package food

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type FoodRepository interface {
	GetFoods(ctx context.Context, query FoodFilterQuery) ([]domain.FoodMaster, int64, error)
	FindByID(ctx context.Context, id string) (*domain.FoodMaster, error)
	Create(ctx context.Context, f *domain.FoodMaster) error
	Update(ctx context.Context, f *domain.FoodMaster) error
	Delete(ctx context.Context, id string) error
	BulkInsertInBatches(ctx context.Context, foods []domain.FoodMaster, batchSize int) (int, error)
	GetStats(ctx context.Context) (*FoodStatsResponse, error)
	GetAllForExport(ctx context.Context, query FoodFilterQuery) ([]domain.FoodMaster, error)
}

type FoodService interface {
	GetFoods(ctx context.Context, query FoodFilterQuery) ([]FoodMasterResponse, int64, error)
	GetByID(ctx context.Context, id string) (*FoodMasterResponse, error)
	Create(ctx context.Context, req CreateFoodRequest) (*FoodMasterResponse, error)
	Update(ctx context.Context, id string, req UpdateFoodRequest) (*FoodMasterResponse, error)
	Delete(ctx context.Context, id string) error
	SearchFoods(ctx context.Context, query FoodFilterQuery) ([]FoodMasterResponse, int64, error)
	PreviewExcelImport(ctx context.Context, fileBytes []byte) (*ExcelImportPreviewResponse, error)
	ConfirmExcelImport(ctx context.Context, req ExcelImportConfirmRequest) (*ExcelImportConfirmResponse, error)
	ExportFoods(ctx context.Context, query FoodFilterQuery, format string) ([]byte, string, string, error)
	GetStats(ctx context.Context) (*FoodStatsResponse, error)
}
