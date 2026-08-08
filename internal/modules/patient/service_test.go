package patient

import (
	"testing"
	"time"
)

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
	if breakdown.BloodSugarScore != 25 {
		t.Errorf("expected blood sugar score 25, got %v", breakdown.BloodSugarScore)
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
