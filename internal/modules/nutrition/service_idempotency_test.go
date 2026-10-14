package nutrition

import (
	"context"
	"testing"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"go.uber.org/zap"
)

type mockNutritionRepo struct {
	meals    map[string]*domain.MealLog
	byClient map[string]*domain.MealLog
}

func newMockNutritionRepo() *mockNutritionRepo {
	return &mockNutritionRepo{
		meals:    make(map[string]*domain.MealLog),
		byClient: make(map[string]*domain.MealLog),
	}
}

func (m *mockNutritionRepo) SearchFoods(ctx context.Context, query string) ([]domain.Food, error) {
	return nil, nil
}
func (m *mockNutritionRepo) FindFoodByID(ctx context.Context, id string) (*domain.Food, error) {
	return &domain.Food{
		BaseModel: domain.BaseModel{ID: id},
		Name:      "Nasi Putih",
		Calories:  200,
	}, nil
}
func (m *mockNutritionRepo) CreateFood(ctx context.Context, f *domain.Food) error {
	return nil
}
func (m *mockNutritionRepo) UpdateFood(ctx context.Context, f *domain.Food) error {
	return nil
}
func (m *mockNutritionRepo) CreateMealLog(ctx context.Context, log *domain.MealLog) error {
	log.ID = "generated-meal-" + time.Now().Format("150405.000")
	m.meals[log.ID] = log
	if log.ClientID != nil && *log.ClientID != "" {
		m.byClient[log.PatientID+":"+*log.ClientID] = log
	}
	return nil
}
func (m *mockNutritionRepo) FindMealLogByID(ctx context.Context, patientID string, id string) (*domain.MealLog, error) {
	return m.meals[id], nil
}
func (m *mockNutritionRepo) FindMealLogByClientID(ctx context.Context, patientID string, clientID string) (*domain.MealLog, error) {
	return m.byClient[patientID+":"+clientID], nil
}
func (m *mockNutritionRepo) UpdateMealLog(ctx context.Context, log *domain.MealLog) error {
	return nil
}
func (m *mockNutritionRepo) DeleteMealLog(ctx context.Context, patientID string, id string) error {
	return nil
}
func (m *mockNutritionRepo) FindMealsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.MealLog, error) {
	return nil, nil
}
func (m *mockNutritionRepo) GetDailyCalorieTarget(ctx context.Context, patientID string) (int, error) {
	return 2000, nil
}
func (m *mockNutritionRepo) UpsertRecentSearch(ctx context.Context, patientID string, foodID string) error {
	return nil
}
func (m *mockNutritionRepo) FindRecentSearches(ctx context.Context, patientID string) ([]domain.RecentFoodSearch, error) {
	return nil, nil
}

func TestLogMeal_Idempotency(t *testing.T) {
	repo := newMockNutritionRepo()
	svc := NewNutritionService(repo, zap.NewNop())

	clientID := "client-meal-uuid-888"
	req := LogMealRequest{
		FoodID:            "food-uuid-1",
		MealType:          domain.MealSarapan,
		PortionMultiplier: 1.0,
		ClientID:          &clientID,
	}

	// 1. Initial request
	resp1, err := svc.LogMeal(context.Background(), "patient-1", req)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// 2. Retry request with same client_id
	resp2, err := svc.LogMeal(context.Background(), "patient-1", req)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	if resp2.ID != resp1.ID {
		t.Errorf("idempotency violated: expected same ID %s, got %s", resp1.ID, resp2.ID)
	}
	if len(repo.meals) != 1 {
		t.Errorf("expected 1 meal record, got %d", len(repo.meals))
	}
}
