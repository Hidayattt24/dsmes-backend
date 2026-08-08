package domain

import (
	"testing"
	"time"
)

func dob(y, m, d int) *time.Time {
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return &t
}

func TestClassifyGDP(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		expected GlucoseCategory
	}{
		{"GDP 65 (hypoglycemia)", 65, CategoryHypoglycemia},
		{"GDP 90 (normal)", 90, CategoryNormal},
		{"GDP 110 (prediabetes)", 110, CategoryPrediabetes},
		{"GDP 135 (hyperglycemia)", 135, CategoryHyperglycemia},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ClassifyBloodGlucose(tt.val, TimeFasting, dob(1985, 6, 15))
			if r.Category != tt.expected {
				t.Errorf("got %s, want %s", r.Category, tt.expected)
			}
		})
	}
}

func TestClassifyGD2PP(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		expected GlucoseCategory
	}{
		{"GD2PP 130 (normal)", 130, CategoryNormal},
		{"GD2PP 160 (prediabetes)", 160, CategoryPrediabetes},
		{"GD2PP 230 (hyperglycemia)", 230, CategoryHyperglycemia},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ClassifyBloodGlucose(tt.val, TimeAfterMeal, dob(1985, 6, 15))
			if r.Category != tt.expected {
				t.Errorf("got %s, want %s", r.Category, tt.expected)
			}
		})
	}
}

func TestClassifyGDS(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		expected GlucoseCategory
	}{
		{"GDS 100 (normal)", 100, CategoryNormal},
		{"GDS 180 (normal)", 180, CategoryNormal},
		{"GDS 250 (hyperglycemia)", 250, CategoryHyperglycemia},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ClassifyBloodGlucose(tt.val, TimeRandom, dob(1985, 6, 15))
			if r.Category != tt.expected {
				t.Errorf("got %s, want %s", r.Category, tt.expected)
			}
		})
	}
}

func TestClassifyBeforeBed(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		expected GlucoseCategory
	}{
		{"BeforeBed 90 (target)", 90, CategoryTarget},
		{"BeforeBed 160 (elevated)", 160, CategoryElevated},
		{"BeforeBed 220 (hyperglycemia)", 220, CategoryHyperglycemia},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ClassifyBloodGlucose(tt.val, TimeBeforeBed, dob(1985, 6, 15))
			if r.Category != tt.expected {
				t.Errorf("got %s, want %s", r.Category, tt.expected)
			}
		})
	}
}

func TestClassifyHypoglycemiaAllTypes(t *testing.T) {
	types := []MeasurementTime{TimeFasting, TimeBeforeMeal, TimeAfterMeal, TimeBeforeBed, TimeRandom}
	for _, mType := range types {
		r := ClassifyBloodGlucose(55, mType, nil)
		if r.Category != CategoryHypoglycemia {
			t.Errorf("%s: 55 should be hypoglycemia, got %s", mType, r.Category)
		}
	}
}

func TestClassifyBeforeMealTarget(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		expected GlucoseCategory
	}{
		{"BeforeMeal 85 (target)", 85, CategoryTarget},
		{"BeforeMeal 150 (elevated)", 150, CategoryElevated},
		{"BeforeMeal 210 (hyperglycemia)", 210, CategoryHyperglycemia},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ClassifyBloodGlucose(tt.val, TimeBeforeMeal, dob(1985, 6, 15))
			if r.Category != tt.expected {
				t.Errorf("got %s, want %s", r.Category, tt.expected)
			}
		})
	}
}

func TestCategoryInfoCompleteness(t *testing.T) {
	categories := []GlucoseCategory{
		CategoryHypoglycemia,
		CategoryNormal,
		CategoryTarget,
		CategoryPrediabetes,
		CategoryElevated,
		CategoryHyperglycemia,
	}
	for _, c := range categories {
		info, ok := categoryInfo[c]
		if !ok {
			t.Errorf("missing categoryInfo for %s", c)
			continue
		}
		if info.Label == "" {
			t.Errorf("%s has empty Label", c)
		}
		if info.Color == "" {
			t.Errorf("%s has empty Color", c)
		}
		if info.Severity == "" {
			t.Errorf("%s has empty Severity", c)
		}
	}
}

func TestBackwardCompatibility(t *testing.T) {
	r1 := CalculateBloodSugarMedicalResult(90, TimeFasting, nil)
	r2 := ClassifyBloodGlucose(90, TimeFasting, nil)
	if r1.Category != r2.Category {
		t.Error("CalculateBloodSugarMedicalResult differs from ClassifyBloodGlucose")
	}

	s := CalculateGlucoseStatus(100, TimeRandom)
	if s != CategoryNormal {
		t.Errorf("CalculateGlucoseStatus returned %s, want normal", s)
	}
}
