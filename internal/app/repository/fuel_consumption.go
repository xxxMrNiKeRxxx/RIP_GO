package repository

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"metoda/internal/app/ds"
	"metoda/internal/app/serializer"
	"time"
)

// ─── helpers ────────────────────────────────────────────────────────────────
func (r *Repository) ClearDraftFuelConsumptionsOnStartup() {
	r.db.Exec("UPDATE fuel_consumptions SET status = ? WHERE status = ?", ds.StatusDeleted, ds.StatusDraft)
}

func (r *Repository) GetCreatorLogin(creatorID uint) string {
	var u ds.Users
	r.db.Where("id = ?", creatorID).First(&u)
	return u.Login
}

func (r *Repository) GetModeratorLogin(moderatorID *uint) string {
	if moderatorID == nil {
		return ""
	}
	var u ds.Users
	r.db.Where("id = ?", *moderatorID).First(&u)
	return u.Login
}

// GetFuelConsumptionEntriesCount — число записей м-м, в которых fuel_saved не пустое
func (r *Repository) GetFuelConsumptionEntriesCount(fuelConsumptionID uint) int {
	var count int64
	r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ?", fuelConsumptionID).
		Where("fuel_saved IS NOT NULL AND fuel_saved > 0").
		Count(&count)
	return int(count)
}

// ─── HTML-layer methods ───────────────────────────────────────────────────────
func (r *Repository) GetDraftFuelConsumption(creatorID uint) (*ds.FuelConsumption, error) {
	var t ds.FuelConsumption
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetFuelConsumptionWithEntries(fuelConsumptionID uint) (*ds.FuelConsumption, []FuelConsumptionEntryView, error) {
	var t ds.FuelConsumption
	err := r.db.First(&t, fuelConsumptionID).Error
	if err != nil {
		return nil, nil, err
	}

	var items []ds.FuelConsumptionMode
	err = r.db.Where("consumption_id = ?", fuelConsumptionID).
		Preload("Mode").
		Order("id").
		Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	logrus.Infof("GetFuelConsumptionWithEntries: Found %d items", len(items))
	for i, item := range items {
		logrus.Infof("  Item %d: ModeID=%d, ModeName=%s, ImageKey=%s",
			i, item.ModeID, item.Mode.ModeName, item.Mode.ImageKey)
	}

	var views []FuelConsumptionEntryView
	for _, item := range items {
		views = append(views, FuelConsumptionEntryView{
			ID:              item.ID,
			ModeName:        item.Mode.ModeName,
			BaseConsumption: item.Mode.BaseConsumption,
			EconomyPercent:  item.Mode.EconomyPercent,
			ImageKey:        item.Mode.ImageKey,
			RouteDistance:   item.RouteDistance,
			FuelSaved:       item.FuelSaved,
			SortOrder:       item.SortOrder,
		})
	}
	return &t, views, nil
}

type FuelConsumptionEntryView struct {
	ID              uint
	ModeName        string
	BaseConsumption float64
	EconomyPercent  float64
	ImageKey        string
	RouteDistance   int
	FuelSaved       float64
	SortOrder       int
}

func (r *Repository) GetCalculatedFuelSaved(fuelConsumptionID uint) []float64 {
	var fuelSaved []float64
	err := r.db.Raw(`
		SELECT fuel_saved
		FROM fuel_consumption_modes
		WHERE consumption_id = ?
		AND fuel_saved IS NOT NULL AND fuel_saved > 0
	`, fuelConsumptionID).Scan(&fuelSaved).Error
	if err != nil {
		logrus.Errorf("GetCalculatedFuelSaved: %v", err)
	}
	return fuelSaved
}

func (r *Repository) AddModeToFuelConsumption(modeID uint, creatorID uint) error {
	var t ds.FuelConsumption
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&t).Error
	if err != nil {
		t = ds.FuelConsumption{
			Status:     ds.StatusDraft,
			DateCreate: time.Now(),
			CreatorID:  creatorID,
		}
		if err := r.db.Create(&t).Error; err != nil {
			return err
		}
	}

	fc := ds.FuelConsumptionMode{
		ConsumptionID: t.ConsumptionID,
		ModeID:        modeID,
		RouteDistance: 300,
		FuelSaved:     0,
		SortOrder:     0,
	}
	return r.db.Create(&fc).Error
}

func (r *Repository) DeleteFuelConsumptionBySQL(fuelConsumptionID uint) error {
	return r.db.Exec("UPDATE fuel_consumptions SET status = '"+ds.StatusDeleted+"' WHERE consumption_id = ?", fuelConsumptionID).Error
}

func (r *Repository) UpdateFuelConsumptionItem(itemID uint, routeDistance int) error {
	var item ds.FuelConsumptionMode
	if err := r.db.First(&item, itemID).Error; err != nil {
		return err
	}

	if routeDistance <= 0 {
		routeDistance = 300
	}

	updates := map[string]interface{}{
		"route_distance": routeDistance,
	}
	return r.db.Model(&item).Updates(updates).Error
}

func (r *Repository) FormFuelConsumption(fuelConsumptionID uint) error {
	var entries []ds.FuelConsumptionMode
	if err := r.db.Where("consumption_id = ?", fuelConsumptionID).Find(&entries).Error; err != nil {
		return err
	}

	var fc ds.FuelConsumption
	if err := r.db.First(&fc, fuelConsumptionID).Error; err != nil {
		return err
	}

	for _, entry := range entries {
		var mode ds.DrivingMode
		if err := r.db.First(&mode, entry.ModeID).Error; err != nil {
			continue
		}

		fuelWithoutCruise := (mode.BaseConsumption / 100) * float64(entry.RouteDistance)
		fuelSaved := fuelWithoutCruise * (mode.EconomyPercent / 100)

		r.db.Model(&entry).Updates(map[string]interface{}{
			"fuel_saved": fuelSaved,
		})
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      ds.StatusFormed,
		"date_formed": now,
	}
	return r.db.Model(&ds.FuelConsumption{}).Where("consumption_id = ?", fuelConsumptionID).Updates(updates).Error
}

func (r *Repository) GetCartCount(creatorID uint) int64 {
	var fuelConsumptionID uint
	var count int64
	err := r.db.Model(&ds.FuelConsumption{}).
		Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).
		Select("consumption_id").
		First(&fuelConsumptionID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ?", fuelConsumptionID).
		Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records:", err)
	}
	return count
}

func (r *Repository) GetDraftFuelConsumptionID(creatorID uint) uint {
	var fuelConsumptionID uint
	err := r.db.Model(&ds.FuelConsumption{}).
		Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).
		Select("consumption_id").
		First(&fuelConsumptionID).Error
	if err != nil {
		return 0
	}
	return fuelConsumptionID
}

// ─── API methods ─────────────────────────────────────────────────────────────
func (r *Repository) GetFuelConsumptionByID(id int) (ds.FuelConsumption, error) {
	var t ds.FuelConsumption
	err := r.db.Where("consumption_id = ?", id).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.FuelConsumption{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, id)
		}
		return ds.FuelConsumption{}, err
	}

	if t.Status == ds.StatusDeleted {
		return ds.FuelConsumption{}, fmt.Errorf("%w: заявка удалена", ErrNotFound)
	}
	return t, nil
}

func (r *Repository) GetAllFuelConsumptions(from, to time.Time, status string) ([]ds.FuelConsumption, error) {
	var list []ds.FuelConsumption
	sub := r.db.Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

	if !from.IsZero() {
		sub = sub.Where("date_formed >= ?", from)
	}
	if !to.IsZero() {
		sub = sub.Where("date_formed <= ?", to.Add(24*time.Hour))
	}
	if status != "" {
		sub = sub.Where("status = ?", status)
	}

	err := sub.Order("consumption_id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repository) GetFuelConsumptionEntriesAPI(id uint) ([]serializer.FuelConsumptionEntryViewJSON, error) {
	var items []ds.FuelConsumptionMode
	err := r.db.Where("consumption_id = ?", id).Preload("Mode").Order("id").Find(&items).Error
	if err != nil {
		return nil, err
	}

	views := make([]serializer.FuelConsumptionEntryViewJSON, 0, len(items))
	for _, item := range items {
		views = append(views, serializer.FuelConsumptionEntryViewJSON{
			ID:              item.ID,
			ModeID:          item.ModeID,
			ModeName:        item.Mode.ModeName,
			BaseConsumption: item.Mode.BaseConsumption,
			EconomyPercent:  item.Mode.EconomyPercent,
			ImageKey:        item.Mode.ImageKey,
			RouteDistance:   item.RouteDistance,
			FuelSaved:       item.FuelSaved,
			SortOrder:       item.SortOrder,
		})
	}
	return views, nil
}

func (r *Repository) GetCartInfo(creatorID uint) (uint, int64, error) {
	var t ds.FuelConsumption
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	var count int64
	r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ?", t.ConsumptionID).
		Count(&count)
	return t.ConsumptionID, count, nil
}

func (r *Repository) UpdateFuelConsumptionFields(id int, j serializer.FuelConsumptionUpdateJSON) (ds.FuelConsumption, error) {
	t, err := r.GetFuelConsumptionByID(id)
	if err != nil {
		return ds.FuelConsumption{}, err
	}

	if t.Status != ds.StatusDraft {
		return ds.FuelConsumption{}, fmt.Errorf("%w: можно менять только черновик", ErrNotAllowed)
	}

	if int(t.CreatorID) != GetUserID() {
		return ds.FuelConsumption{}, fmt.Errorf("%w: только создатель может редактировать заявку", ErrNotAllowed)
	}

	updates := map[string]interface{}{}
	if j.Origin != "" {
		updates["origin"] = j.Origin
	}
	if j.Destination != "" {
		updates["destination"] = j.Destination
	}
	if j.FuelPrice > 0 {
		updates["fuel_price"] = j.FuelPrice
	}

	if len(updates) > 0 {
		if err := r.db.Model(&t).Updates(updates).Error; err != nil {
			return ds.FuelConsumption{}, err
		}
	}

	return r.GetFuelConsumptionByID(id)
}

func (r *Repository) FormFuelConsumptionAPI(id int) (ds.FuelConsumption, error) {
	t, err := r.GetFuelConsumptionByID(id)
	if err != nil {
		return ds.FuelConsumption{}, err
	}

	if t.Status != ds.StatusDraft {
		return ds.FuelConsumption{}, fmt.Errorf("нельзя сформировать заявку со статусом %s", t.Status)
	}

	if int(t.CreatorID) != GetUserID() {
		return ds.FuelConsumption{}, fmt.Errorf("%w: только создатель может сформировать заявку", ErrNotAllowed)
	}

	var count int64
	r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ?", t.ConsumptionID).
		Count(&count)
	if count == 0 {
		return ds.FuelConsumption{}, fmt.Errorf("нельзя сформировать пустую заявку: добавьте хотя бы один режим")
	}

	var entries []ds.FuelConsumptionMode
	if err := r.db.Where("consumption_id = ?", t.ConsumptionID).Find(&entries).Error; err != nil {
		return ds.FuelConsumption{}, err
	}

	for _, entry := range entries {
		var mode ds.DrivingMode
		if err := r.db.First(&mode, entry.ModeID).Error; err != nil {
			continue
		}

		fuelWithoutCruise := (mode.BaseConsumption / 100) * float64(entry.RouteDistance)
		fuelSaved := fuelWithoutCruise * (mode.EconomyPercent / 100)

		if err := r.db.Model(&entry).Update("fuel_saved", fuelSaved).Error; err != nil {
			return ds.FuelConsumption{}, err
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      ds.StatusFormed,
		"date_formed": now,
	}
	if err := r.db.Model(&t).Updates(updates).Error; err != nil {
		return ds.FuelConsumption{}, err
	}

	return r.GetFuelConsumptionByID(id)
}

func (r *Repository) FinishFuelConsumptionAPI(id int, status string) (ds.FuelConsumption, error) {
	if status != ds.StatusCompleted && status != ds.StatusRejected {
		return ds.FuelConsumption{}, fmt.Errorf("недопустимый статус: ожидается '%s' или '%s'", ds.StatusCompleted, ds.StatusRejected)
	}

	moderator, err := r.GetUserByID(GetUserID())
	if err != nil {
		return ds.FuelConsumption{}, err
	}

	if !moderator.IsModerator {
		return ds.FuelConsumption{}, fmt.Errorf("%w: только модератор может завершить или отклонить заявку", ErrNotAllowed)
	}

	t, err := r.GetFuelConsumptionByID(id)
	if err != nil {
		return ds.FuelConsumption{}, err
	}

	if t.Status != ds.StatusFormed {
		return ds.FuelConsumption{}, fmt.Errorf("завершить или отклонить можно только сформированную заявку; текущий статус — %s (сначала PUT .../form)", t.Status)
	}

	now := time.Now()
	moderatorID := uint(moderator.ID)
	if err := r.db.Model(&t).Updates(map[string]interface{}{
		"status":         status,
		"date_completed": now,
		"moderator_id":   moderatorID,
	}).Error; err != nil {
		return ds.FuelConsumption{}, err
	}

	return r.GetFuelConsumptionByID(id)
}

// ✅ DELETE /api/fuel-consumptions/:id — КРИТИЧНО: функция должна быть определена!
func (r *Repository) DeleteFuelConsumptionAPI(id int) error {
	t, err := r.GetFuelConsumptionByID(id)
	if err != nil {
		return err
	}

	if t.Status != ds.StatusDraft {
		return fmt.Errorf("%w: только черновик может быть удалён создателем", ErrNotAllowed)
	}

	if int(t.CreatorID) != GetUserID() {
		return fmt.Errorf("%w: только создатель может удалить заявку", ErrNotAllowed)
	}

	return r.db.Model(&t).Update("status", ds.StatusDeleted).Error
}

func (r *Repository) AddModeToCartAPI(modeID uint, creatorID uint) (ds.FuelConsumption, bool, error) {
	var mode ds.DrivingMode
	if err := r.db.Where("mode_id = ? AND is_active = ?", modeID, true).First(&mode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.FuelConsumption{}, false, fmt.Errorf("%w: режим с id %d", ErrNotFound, modeID)
		}
		return ds.FuelConsumption{}, false, err
	}

	var fc ds.FuelConsumption
	created := false
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.StatusDraft).First(&fc).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		fc = ds.FuelConsumption{
			Status:     ds.StatusDraft,
			DateCreate: time.Now(),
			CreatorID:  creatorID,
		}
		if err := r.db.Create(&fc).Error; err != nil {
			return ds.FuelConsumption{}, false, err
		}
		created = true
	} else if err != nil {
		return ds.FuelConsumption{}, false, err
	}

	var existing ds.FuelConsumptionMode
	res := r.db.Where("consumption_id = ? AND mode_id = ?", fc.ConsumptionID, modeID).First(&existing)
	if res.Error == nil {
		return ds.FuelConsumption{}, false, fmt.Errorf("%w: режим %d уже добавлен в заявку %d", ErrAlreadyExists, modeID, fc.ConsumptionID)
	}

	fuelWithoutCruise := (mode.BaseConsumption / 100) * 300.0
	fuelSaved := fuelWithoutCruise * (mode.EconomyPercent / 100)

	fcm := ds.FuelConsumptionMode{
		ConsumptionID: fc.ConsumptionID,
		ModeID:        modeID,
		RouteDistance: 300,
		FuelSaved:     fuelSaved,
		SortOrder:     0,
	}
	if err := r.db.Create(&fcm).Error; err != nil {
		return ds.FuelConsumption{}, false, err
	}

	return fc, created, nil
}

func (r *Repository) DeleteModeFromCartAPI(modeID, fuelConsumptionID int) (ds.FuelConsumption, error) {
	t, err := r.GetFuelConsumptionByID(fuelConsumptionID)
	if err != nil {
		return ds.FuelConsumption{}, err
	}

	if t.Status != ds.StatusDraft {
		return ds.FuelConsumption{}, fmt.Errorf("%w: нельзя изменить не черновик", ErrNotAllowed)
	}

	err = r.db.Where("mode_id = ? AND consumption_id = ?", modeID, fuelConsumptionID).Delete(&ds.FuelConsumptionMode{}).Error
	if err != nil {
		return ds.FuelConsumption{}, err
	}
	return t, nil
}

func (r *Repository) UpdateModeInCartAPI(modeID, fuelConsumptionID int, j serializer.FuelConsumptionModeUpdateJSON) (ds.FuelConsumptionMode, error) {
	t, err := r.GetFuelConsumptionByID(fuelConsumptionID)
	if err != nil {
		return ds.FuelConsumptionMode{}, err
	}

	if t.Status != ds.StatusDraft {
		return ds.FuelConsumptionMode{}, fmt.Errorf("%w: нельзя изменить не черновик", ErrNotAllowed)
	}

	var item ds.FuelConsumptionMode
	err = r.db.Where("mode_id = ? AND consumption_id = ?", modeID, fuelConsumptionID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.FuelConsumptionMode{}, fmt.Errorf("%w: связь не найдена", ErrNotFound)
		}
		return ds.FuelConsumptionMode{}, err
	}

	updates := map[string]interface{}{}
	if j.RouteDistance > 0 {
		updates["route_distance"] = j.RouteDistance

		var mode ds.DrivingMode
		if err := r.db.First(&mode, item.ModeID).Error; err == nil {
			fuelWithoutCruise := (mode.BaseConsumption / 100) * float64(j.RouteDistance)
			fuelSaved := fuelWithoutCruise * (mode.EconomyPercent / 100)
			updates["fuel_saved"] = fuelSaved
		}
	}

	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return ds.FuelConsumptionMode{}, err
	}

	r.db.Where("mode_id = ? AND consumption_id = ?", modeID, fuelConsumptionID).First(&item)
	return item, nil
}

func (r *Repository) GetFuelConsumptionEntryByID(id uint) (ds.FuelConsumptionMode, error) {
	var item ds.FuelConsumptionMode
	err := r.db.First(&item, id).Error
	return item, err
}
