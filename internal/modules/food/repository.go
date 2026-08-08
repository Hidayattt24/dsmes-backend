package food

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type foodRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewFoodRepository(db *gorm.DB, log *zap.Logger) FoodRepository {
	return &foodRepository{
		db:  db,
		log: log,
	}
}

func (r *foodRepository) GetFoods(ctx context.Context, query FoodFilterQuery) ([]domain.FoodMaster, int64, error) {
	var foods []domain.FoodMaster
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.FoodMaster{})

	db = r.applyFilters(db, query)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("foodRepo.GetFoods count error: %w", err)
	}

	// Sorting
	sortBy := "created_at"
	sortOrder := "DESC"
	if query.SortBy != "" {
		allowedSorts := map[string]bool{
			"name":         true,
			"energy_kcal":  true,
			"created_at":   true,
			"status":       true,
			"manufacturer": true,
		}
		if allowedSorts[strings.ToLower(query.SortBy)] {
			sortBy = strings.ToLower(query.SortBy)
		}
	}
	if strings.EqualFold(query.SortOrder, "asc") {
		sortOrder = "ASC"
	}
	db = db.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Pagination
	page := query.Page
	if page <= 0 {
		page = 1
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	if err := db.Offset(offset).Limit(limit).Find(&foods).Error; err != nil {
		return nil, 0, fmt.Errorf("foodRepo.GetFoods find error: %w", err)
	}

	return foods, total, nil
}

func (r *foodRepository) applyFilters(db *gorm.DB, query FoodFilterQuery) *gorm.DB {
	if query.Q != "" {
		q := "%" + strings.TrimSpace(query.Q) + "%"
		db = db.Where("name ILIKE ? OR manufacturer ILIKE ? OR barcode ILIKE ?", q, q, q)
	}

	if query.Manufacturer != "" {
		db = db.Where("manufacturer ILIKE ?", "%"+strings.TrimSpace(query.Manufacturer)+"%")
	}

	if query.MinCalories > 0 {
		db = db.Where("energy_kcal >= ?", query.MinCalories)
	}

	if query.MaxCalories > 0 {
		db = db.Where("energy_kcal <= ?", query.MaxCalories)
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	return db
}

func (r *foodRepository) FindByID(ctx context.Context, id string) (*domain.FoodMaster, error) {
	var food domain.FoodMaster
	if err := r.db.WithContext(ctx).First(&food, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &food, nil
}

func (r *foodRepository) Create(ctx context.Context, f *domain.FoodMaster) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(f).Error
}

func (r *foodRepository) Update(ctx context.Context, f *domain.FoodMaster) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *foodRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.FoodMaster{}, "id = ?", id).Error
}

func (r *foodRepository) BulkInsertInBatches(ctx context.Context, foods []domain.FoodMaster, batchSize int) (int, error) {
	if len(foods) == 0 {
		return 0, nil
	}

	if batchSize <= 0 {
		batchSize = 500
	}

	startTime := time.Now()
	r.log.Info("starting food bulk import batch insert", zap.Int("total_records", len(foods)), zap.Int("batch_size", batchSize))

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&foods, batchSize).Error
	if err != nil {
		r.log.Error("failed food bulk batch insert", zap.Error(err))
		return 0, err
	}

	r.log.Info("completed food bulk import batch insert successfully",
		zap.Int("rows_imported", len(foods)),
		zap.Duration("duration", time.Since(startTime)),
	)
	return len(foods), nil
}

func (r *foodRepository) GetStats(ctx context.Context) (*FoodStatsResponse, error) {
	var stats FoodStatsResponse
	db := r.db.WithContext(ctx).Model(&domain.FoodMaster{})

	// Total Foods
	if err := db.Count(&stats.TotalFoods).Error; err != nil {
		return nil, err
	}

	// Active Foods
	if err := r.db.WithContext(ctx).Model(&domain.FoodMaster{}).Where("status = ?", "active").Count(&stats.ActiveFoods).Error; err != nil {
		return nil, err
	}

	// Today Imported Foods
	todayStart := time.Now().Truncate(24 * time.Hour)
	if err := r.db.WithContext(ctx).Model(&domain.FoodMaster{}).Where("created_at >= ?", todayStart).Count(&stats.TodayImportedFoods).Error; err != nil {
		return nil, err
	}

	// Total Unique Manufacturers
	if err := r.db.WithContext(ctx).Model(&domain.FoodMaster{}).
		Where("manufacturer IS NOT NULL AND manufacturer != ''").
		Select("COUNT(DISTINCT manufacturer)").
		Scan(&stats.TotalManufacturers).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *foodRepository) GetAllForExport(ctx context.Context, query FoodFilterQuery) ([]domain.FoodMaster, error) {
	var foods []domain.FoodMaster
	db := r.db.WithContext(ctx).Model(&domain.FoodMaster{})
	db = r.applyFilters(db, query)
	db = db.Order("name ASC")

	limit := query.Limit
	if limit <= 0 || limit > 50000 {
		limit = 50000
	}
	db = db.Limit(limit)

	if err := db.Find(&foods).Error; err != nil {
		return nil, err
	}
	return foods, nil
}
