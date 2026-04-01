package repository

import (
	"context"
	"fmt"
	"metoda/internal/app/ds"
	minio "metoda/internal/app/minioClient"
	"metoda/internal/app/serializer"
	"mime/multipart"
)

func (r *Repository) GetAllModes() ([]ds.DrivingMode, error) {
	var modes []ds.DrivingMode
	err := r.db.Where("is_active = ?", true).Find(&modes).Error
	if err != nil {
		return nil, err
	}
	return modes, nil
}

func (r *Repository) GetModeByID(id int) (*ds.DrivingMode, error) {
	var t ds.DrivingMode
	err := r.db.Where("mode_id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) SearchModesByName(title string) ([]ds.DrivingMode, error) {
	var modes []ds.DrivingMode
	err := r.db.Where("mode_name ILIKE ? AND is_active = ?", "%"+title+"%", true).Find(&modes).Error
	if err != nil {
		return nil, err
	}
	return modes, nil
}

// CreateTire — только для модератора
func (r *Repository) CreateMode(j serializer.ModeJSON) (ds.DrivingMode, error) {
	if j.ModeName == "" {
		return ds.DrivingMode{}, fmt.Errorf("поле mode_name обязательно")
	}
	if j.BaseConsumption == 0 {
		return ds.DrivingMode{}, fmt.Errorf("поле base_consumption обязательно")
	}
	if j.EconomyPercent == 0 {
		return ds.DrivingMode{}, fmt.Errorf("поле economy_percent обязательно")
	}
	if j.Description == "" {
		return ds.DrivingMode{}, fmt.Errorf("поле description обязательно")
	}

	t := serializer.ModeFromJSON(j)
	if err := r.db.Create(&t).Error; err != nil {
		return ds.DrivingMode{}, err
	}
	return t, nil
}

func (r *Repository) UploadModeImage(ctx context.Context, modeID int, file *multipart.FileHeader) (ds.DrivingMode, error) {
	// 1. Получаем режим по ID (уверены, что ModeID заполнен)
	t, err := r.GetModeByID(modeID)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	// 2. Убеждаемся, что режим существует и ModeID корректен
	if t.ModeID == 0 {
		return ds.DrivingMode{}, fmt.Errorf("invalid mode_id: %d", modeID)
	}

	// 3. Создаём бакет, если его нет
	if err := minio.EnsureBucket(ctx, r.mc, minio.ServicesBucket); err != nil {
		return ds.DrivingMode{}, err
	}

	// 4. Загружаем файл в MinIO
	objectName, err := minio.UploadFile(ctx, r.mc, minio.ServicesBucket, file)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	// 5. Обновляем ТОЛЬКО поле image_key через WHERE — это безопасно и не затрагивает ModeID
	if err := r.db.Model(&ds.DrivingMode{}).
		Where("mode_id = ?", modeID).
		Update("image_key", objectName).Error; err != nil {
		return ds.DrivingMode{}, err
	}

	// 6. Обновляем локальную копию t.ImageKey, чтобы вернуть актуальные данные
	t.ImageKey = objectName

	return *t, nil
}

func (r *Repository) UploadModeVideo(ctx context.Context, modeID int, file *multipart.FileHeader) (ds.DrivingMode, error) {
	// 1. Получаем режим по ID
	t, err := r.GetModeByID(modeID)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	// 2. Проверка ModeID
	if t.ModeID == 0 {
		return ds.DrivingMode{}, fmt.Errorf("invalid mode_id: %d", modeID)
	}

	// 3. Убеждаемся, что бакет существует
	if err := minio.EnsureBucket(ctx, r.mc, minio.ServicesBucket); err != nil {
		return ds.DrivingMode{}, err
	}

	// 4. Загружаем видео в MinIO
	objectName, err := minio.UploadFile(ctx, r.mc, minio.ServicesBucket, file)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	// 5. Обновляем только video_key
	if err := r.db.Model(&ds.DrivingMode{}).
		Where("mode_id = ?", modeID).
		Update("video_key", objectName).Error; err != nil {
		return ds.DrivingMode{}, err
	}

	// 6. Обновляем локальную копию
	t.VideoKey = objectName

	return *t, nil
}