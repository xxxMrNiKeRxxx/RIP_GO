package repository

import (
	"errors"
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	minioClient "metoda/internal/app/minioClient"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotAllowed    = errors.New("not allowed")
	ErrNoDraft       = errors.New("no draft for this user")
)

// ─── SINGLETON: пользователь (для лаб 1-3, до авторизации) ─────────────────
var currentUserID int = 1

func GetUserID() int {
	return currentUserID
}

func SetUserID(id int) {
	currentUserID = id
}

func SignOut() {
	currentUserID = 0
}

// ─── Repository ─────────────────────────────────────────────────────────────
type Repository struct {
	db *gorm.DB
	mc *minio.Client
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	mc, err := minioClient.InitMinio()
	if err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
		mc: mc,
	}, nil
}
