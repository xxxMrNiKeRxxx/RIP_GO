package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	minioClient "metoda/internal/app/minioClient"
	// ❌ УДАЛИТЕ ЭТОТ ИМПОРТ ИЗ repository.go, ЕСЛИ МЕТОДЫ ТАМ НЕ ИСПОЛЬЗУЮТСЯ
	// "metoda/internal/app/ds"
)

// ─── Errors ─────────────────────────────────────────────────────────────────
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAllowed    = errors.New("not allowed")
	ErrNoDraft       = errors.New("no draft for this user")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token expired")
)

// ─── Repository ─────────────────────────────────────────────────────────────

type Repository struct {
	db     *gorm.DB
	mc     *minio.Client
	rd     *redis.Client // Redis для blacklist JWT
	userID int           // ID текущего пользователя (в рамках запроса)
}

// ─── Constructor ────────────────────────────────────────────────────────────

func New(dsn string) (*Repository, error) {
	// PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	// MinIO
	mc, err := minioClient.InitMinio()
	if err != nil {
		return nil, fmt.Errorf("minio: %w", err)
	}

	// Redis (опционально, через ENV)
	var rdb *redis.Client
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		})
		// Проверка подключения с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			// Если Redis не критичен, можно вернуть warning вместо ошибки:
			// logrus.Warnf("redis ping failed: %v", err)
			// return &Repository{db: db, mc: mc, rd: nil, userID: 1}, nil
			return nil, fmt.Errorf("redis ping: %w", err)
		}
	}

	return &Repository{
		db:     db,
		mc:     mc,
		rd:     rdb,
		userID: 1, // default, будет перезаписан при авторизации
	}, nil
}

// ─── User Context Management (instance methods) ─────────────────────────────

func (r *Repository) GetUserID() int {
	return r.userID
}

func (r *Repository) SetUserID(id int) {
	r.userID = id
}

func (r *Repository) SignOut() {
	r.userID = 0
}

// ─── Redis: JWT Blacklist ───────────────────────────────────────────────────

// AddToBlacklist добавляет токен в чёрный список на время его жизни
func (r *Repository) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	if r.rd == nil {
		return nil // Redis не подключён — пропускаем
	}
	// Ключ вида "blacklist:<token>", значение "1"
	return r.rd.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

// IsBlacklisted проверяет, отозван ли токен
func (r *Repository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if r.rd == nil {
		return false, nil
	}
	val, err := r.rd.Get(ctx, "blacklist:"+token).Result()
	if err == redis.Nil {
		return false, nil // Токена в блэклисте нет
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

// ─── Helpers for GORM / Minio / Redis ───────────────────────────────────────

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) Minio() *minio.Client {
	return r.mc
}

func (r *Repository) Redis() *redis.Client {
	return r.rd
}

// ❌ УДАЛЕНО: GetCreatorLogin, GetModeratorLogin
// Они определены в fuel_consumption.go