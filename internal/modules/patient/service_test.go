package patient

import (
	"math"
	"testing"
	"time"
)

func TestProfileMetricsChanged(t *testing.T) {
	tests := []struct {
		name                 string
		oldHeight, oldWeight float64
		newHeight, newWeight float64
		want                 bool
	}{
		{"profile-only update", 170, 65, 170, 65, false},
		{"weight changed", 170, 65, 170, 66, true},
		{"height changed", 170, 65, 171, 65, true},
		{"float noise", 170, 65, 170 + math.SmallestNonzeroFloat64, 65, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := profileMetricsChanged(tt.oldHeight, tt.oldWeight, tt.newHeight, tt.newWeight); got != tt.want {
				t.Fatalf("profileMetricsChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
}

func fullAggregates() map[string]*DailyLogsAggregate {
	now := fixedNow()
	agg := make(map[string]*DailyLogsAggregate)
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		agg[day] = &DailyLogsAggregate{
			BloodSugarCount:          1,
			TotalMealCalories:        2000,
			MealCount:                3,
			TotalActivityMinutes:     30,
			MedicationCompletedCount: 2,
			MedicationScheduledCount: 2,
		}
	}
	return agg
}

func TestComplianceFromAggregatesEmpty(t *testing.T) {
	now := fixedNow()
	score, label, breakdown := complianceFromAggregates(
		map[string]*DailyLogsAggregate{},
		2000,
		now.AddDate(0, 0, -30),
		now,
	)

	if score != 0 {
		t.Errorf("expected score 0 for empty data, got %d", score)
	}
	if label != "Tidak Patuh" {
		t.Errorf("expected label 'Tidak Patuh', got %s", label)
	}
	if breakdown == nil {
		t.Fatal("expected a compliance breakdown")
	}
}

func TestComplianceFromAggregatesFull(t *testing.T) {
	now := fixedNow()
	score, label, breakdown := complianceFromAggregates(
		fullAggregates(),
		2000,
		now.AddDate(0, 0, -30),
		now,
	)

	if score != 100 {
		t.Errorf("expected score 100 for perfect compliance, got %d", score)
	}
	if label != "Sangat Patuh" {
		t.Errorf("expected label 'Sangat Patuh', got %s", label)
	}
	if breakdown.BloodSugarScore != 0 {
		t.Errorf("expected blood sugar score 0, got %v", breakdown.BloodSugarScore)
	}
}

func TestComplianceExcludesBloodSugar(t *testing.T) {
	now := fixedNow()
	// Scenario: Same food, activity, and medication, but BloodSugarCount is 0 vs 10
	aggWithoutBS := make(map[string]*DailyLogsAggregate)
	aggWithBS := make(map[string]*DailyLogsAggregate)
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		aggWithoutBS[day] = &DailyLogsAggregate{
			BloodSugarCount:          0,
			TotalMealCalories:        2000,
			MealCount:                3,
			TotalActivityMinutes:     30,
			MedicationCompletedCount: 2,
			MedicationScheduledCount: 2,
		}
		aggWithBS[day] = &DailyLogsAggregate{
			BloodSugarCount:          10,
			TotalMealCalories:        2000,
			MealCount:                3,
			TotalActivityMinutes:     30,
			MedicationCompletedCount: 2,
			MedicationScheduledCount: 2,
		}
	}

	scoreWithoutBS, _, _ := complianceFromAggregates(aggWithoutBS, 2000, now.AddDate(0, 0, -30), now)
	scoreWithBS, _, _ := complianceFromAggregates(aggWithBS, 2000, now.AddDate(0, 0, -30), now)

	if scoreWithoutBS != scoreWithBS {
		t.Errorf("expected compliance scores to be identical regardless of blood sugar, got %d vs %d", scoreWithoutBS, scoreWithBS)
	}
}

func TestComplianceFromAggregatesDefaultTarget(t *testing.T) {
	now := fixedNow()
	// dailyTarget <= 0 must fall back to the domain default (2000), so the food
	// ratio stays 1.0 and keeps a perfect food score.
	score, _, _ := complianceFromAggregates(
		fullAggregates(),
		0,
		now.AddDate(0, 0, -30),
		now,
	)
	if score != 100 {
		t.Errorf("expected score 100 with default target, got %d", score)
	}
}

func TestComplianceFromAggregatesEvalWindowClamp(t *testing.T) {
	now := fixedNow()
	// Newly registered patient (created today) -> evalWindow clamps to 1 day,
	// so a single perfect day still yields 100.
	score, _, _ := complianceFromAggregates(
		fullAggregates(),
		2000,
		now,
		now,
	)
	if score != 100 {
		t.Errorf("expected score 100 for single-day perfect compliance, got %d", score)
	}
}
