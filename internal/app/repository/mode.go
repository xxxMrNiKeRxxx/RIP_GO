package repository

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
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
	var mode ds.DrivingMode
	err := r.db.Where("mode_id = ? AND is_active = ?", id, true).First(&mode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: режим с id %d", ErrNotFound, id)
		}
		return nil, err
	}
	return &mode, nil
}

func (r *Repository) SearchModesByName(title string) ([]ds.DrivingMode, error) {
	var modes []ds.DrivingMode
	err := r.db.Where("mode_name ILIKE ? AND is_active = ?", "%"+title+"%", true).Find(&modes).Error
	if err != nil {
		return nil, err
	}
	return modes, nil
}

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
	t, err := r.GetModeByID(modeID)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	if err := minio.EnsureBucket(ctx, r.mc, minio.ServicesBucket); err != nil {
		return ds.DrivingMode{}, err
	}

	objectName, err := minio.UploadFile(ctx, r.mc, minio.ServicesBucket, file)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	t.ImageKey = objectName
	if err := r.db.Save(t).Error; err != nil {
		return ds.DrivingMode{}, err
	}
	return *t, nil
}

func (r *Repository) UploadModeVideo(ctx context.Context, modeID int, file *multipart.FileHeader) (ds.DrivingMode, error) {
	t, err := r.GetModeByID(modeID)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	if err := minio.EnsureBucket(ctx, r.mc, minio.ServicesBucket); err != nil {
		return ds.DrivingMode{}, err
	}

	objectName, err := minio.UploadFile(ctx, r.mc, minio.ServicesBucket, file)
	if err != nil {
		return ds.DrivingMode{}, err
	}

	t.VideoKey = objectName
	if err := r.db.Save(t).Error; err != nil {
		return ds.DrivingMode{}, err
	}
	return *t, nil
}
