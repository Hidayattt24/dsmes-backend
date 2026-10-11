package nutrition_test

import (
	"testing"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/modules/nutrition"
)

func TestToFoodResponse_Nutrients(t *testing.T) {
	food := &domain.Food{
		Name:                "Nasi Putih",
		DefaultServingUnit:  "100 g",
		DefaultServingGrams: 100,
		Calories:            130,
		CarbsG:              28.2,
		ProteinG:            2.7,
		FatG:                0.3,
		SugarG:              0.1,
		SodiumMg:            1.0,
		FiberG:              0.4,
	}

	resp := nutrition.ToFoodResponse(food)
	if resp.Calories != 130 || resp.SugarG != 0.1 || resp.SodiumMg != 1.0 || resp.FiberG != 0.4 {
		t.Errorf("expected food response to carry complete nutrition values, got: %+v", resp)
	}
}
