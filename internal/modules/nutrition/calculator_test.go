package nutrition_test

import (
	"testing"
	"time"

	"github.com/dsmes/dsmes-backend/internal/modules/nutrition"
)

func TestCalculateDailyCalories_NormalCases(t *testing.T) {
	// Male, 21 years old (birthyear now-21), 176 cm, 80 kg, Moderate Activity (1.55)
	dobMale := time.Now().AddDate(-21, 0, 0).Format("2006-01-02")
	reqMale := nutrition.CalorieCalculationRequest{
		Gender:        "male",
		DateOfBirth:   dobMale,
		HeightCm:      176,
		WeightKg:      80,
		ActivityLevel: "moderate",
	}

	resMale, err := nutrition.CalculateDailyCalories(reqMale)
	if err != nil {
		t.Fatalf("expected no error for male request, got: %v", err)
	}

	t.Logf("===============================================================")
	t.Logf("CALORIE CALCULATION AUDIT LOGS FOR TEST CASE (MALE, 21 Y.O, 176 CM, 80 KG, MODERATE ACTIVE)")
	t.Logf("===============================================================")
	t.Logf("BMR                  : %d kcal", resMale.BMR)
	t.Logf("TDEE                 : %d kcal", resMale.TDEE)
	t.Logf("BMI                  : %.1f (%s)", resMale.BMI, resMale.BMICategory)
	t.Logf("Multiplier           : %.3f", resMale.ActivityMultiplier)
	t.Logf("---------------------------------------------------------------")
	t.Logf("4-Tier Calorie Recommendations:")
	t.Logf("1. Maintain Weight   : %d kcal (%d%%) - %s", 
		resMale.Recommendations.Maintain.Calories, 
		resMale.Recommendations.Maintain.Percentage,
		resMale.Recommendations.Maintain.Title)
	t.Logf("2. Mild Weight Loss  : %d kcal (%d%%) - %s [Target: %s]", 
		resMale.Recommendations.MildLoss.Calories, 
		resMale.Recommendations.MildLoss.Percentage,
		resMale.Recommendations.MildLoss.Title,
		resMale.Recommendations.MildLoss.WeeklyTarget)
	t.Logf("3. Weight Loss       : %d kcal (%d%%) - %s [Target: %s]", 
		resMale.Recommendations.WeightLoss.Calories, 
		resMale.Recommendations.WeightLoss.Percentage,
		resMale.Recommendations.WeightLoss.Title,
		resMale.Recommendations.WeightLoss.WeeklyTarget)
	t.Logf("4. Extreme Weight Loss: %d kcal (%d%%) - %s [Target: %s]", 
		resMale.Recommendations.ExtremeLoss.Calories, 
		resMale.Recommendations.ExtremeLoss.Percentage,
		resMale.Recommendations.ExtremeLoss.Title,
		resMale.Recommendations.ExtremeLoss.WeeklyTarget)
	t.Logf("===============================================================")

	if resMale.Age != 21 {
		t.Errorf("expected age 21, got %d", resMale.Age)
	}
	if resMale.BMI != 25.8 {
		t.Errorf("expected BMI 25.8, got %f", resMale.BMI)
	}
	if resMale.BMICategory != "Obese" { // WHO Asia-Pacific threshold >= 25.0
		t.Errorf("expected BMI category Obese, got %s", resMale.BMICategory)
	}
	if resMale.Recommendations.Maintain.Calories != 2790 {
		t.Errorf("expected Maintain calories 2790, got %d", resMale.Recommendations.Maintain.Calories)
	}
	if resMale.Recommendations.MildLoss.Calories != 2540 { // 2790 - 250
		t.Errorf("expected MildLoss calories 2540, got %d", resMale.Recommendations.MildLoss.Calories)
	}
	if resMale.Recommendations.WeightLoss.Calories != 2290 { // 2790 - 500
		t.Errorf("expected WeightLoss calories 2290, got %d", resMale.Recommendations.WeightLoss.Calories)
	}
	if resMale.Recommendations.ExtremeLoss.Calories != 1790 { // 2790 - 1000
		t.Errorf("expected ExtremeLoss calories 1790, got %d", resMale.Recommendations.ExtremeLoss.Calories)
	}
	// Female, 25 years old (birthyear now-25), 160 cm, 50 kg, Light Activity (1.375)
	dobFemale := time.Now().AddDate(-25, 0, 0).Format("2006-01-02")
	reqFemale := nutrition.CalorieCalculationRequest{
		Gender:        "female",
		DateOfBirth:   dobFemale,
		HeightCm:      160,
		WeightKg:      50,
		ActivityLevel: "light",
	}

	resFemale, err := nutrition.CalculateDailyCalories(reqFemale)
	if err != nil {
		t.Fatalf("expected no error for female request, got: %v", err)
	}

	t.Logf("===============================================================")
	t.Logf("CALORIE CALCULATION AUDIT LOGS FOR TEST CASE (FEMALE, 25 Y.O, 160 CM, 50 KG, LIGHT ACTIVE - SAFETY CLAMPING TEST)")
	t.Logf("===============================================================")
	t.Logf("BMR                  : %d kcal", resFemale.BMR)
	t.Logf("TDEE                 : %d kcal", resFemale.TDEE)
	t.Logf("BMI                  : %.1f (%s)", resFemale.BMI, resFemale.BMICategory)
	t.Logf("Multiplier           : %.3f", resFemale.ActivityMultiplier)
	t.Logf("---------------------------------------------------------------")
	t.Logf("4-Tier Calorie Recommendations:")
	t.Logf("1. Maintain Weight   : %d kcal (%d%%) - %s", 
		resFemale.Recommendations.Maintain.Calories, 
		resFemale.Recommendations.Maintain.Percentage,
		resFemale.Recommendations.Maintain.Title)
	t.Logf("2. Mild Weight Loss  : %d kcal (%d%%) - %s [Target: %s]", 
		resFemale.Recommendations.MildLoss.Calories, 
		resFemale.Recommendations.MildLoss.Percentage,
		resFemale.Recommendations.MildLoss.Title,
		resFemale.Recommendations.MildLoss.WeeklyTarget)
	t.Logf("3. Weight Loss       : %d kcal (%d%%) - %s [Target: %s] (Clamped to 1200)", 
		resFemale.Recommendations.WeightLoss.Calories, 
		resFemale.Recommendations.WeightLoss.Percentage,
		resFemale.Recommendations.WeightLoss.Title,
		resFemale.Recommendations.WeightLoss.WeeklyTarget)
	t.Logf("4. Extreme Weight Loss: %d kcal (%d%%) - %s [Target: %s] (Clamped to 1200)", 
		resFemale.Recommendations.ExtremeLoss.Calories, 
		resFemale.Recommendations.ExtremeLoss.Percentage,
		resFemale.Recommendations.ExtremeLoss.Title,
		resFemale.Recommendations.ExtremeLoss.WeeklyTarget)
	t.Logf("===============================================================")

	if resFemale.Age != 25 {
		t.Errorf("expected age 25, got %d", resFemale.Age)
	}
	if resFemale.BMI != 19.5 {
		t.Errorf("expected BMI 19.5, got %f", resFemale.BMI)
	}
	if resFemale.Recommendations.Maintain.Calories != 1669 {
		t.Errorf("expected Maintain calories 1669, got %d", resFemale.Recommendations.Maintain.Calories)
	}
	if resFemale.Recommendations.MildLoss.Calories != 1419 { // 1669 - 250
		t.Errorf("expected MildLoss calories 1419, got %d", resFemale.Recommendations.MildLoss.Calories)
	}
	if resFemale.Recommendations.WeightLoss.Calories != 1200 { // 1669 - 500 = 1169 -> clamped to 1200
		t.Errorf("expected WeightLoss calories 1200, got %d", resFemale.Recommendations.WeightLoss.Calories)
	}
	if resFemale.Recommendations.ExtremeLoss.Calories != 1200 { // 1669 - 1000 = 669 -> clamped to 1200
		t.Errorf("expected ExtremeLoss calories 1200, got %d", resFemale.Recommendations.ExtremeLoss.Calories)
	}
}

func TestCalculateDailyCalories_EdgeCases(t *testing.T) {
	// Future Date of Birth
	futureDOB := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	_, err := nutrition.CalculateDailyCalories(nutrition.CalorieCalculationRequest{
		Gender:        "male",
		DateOfBirth:   futureDOB,
		HeightCm:      176,
		WeightKg:      80,
		ActivityLevel: "moderate",
	})
	if err == nil {
		t.Errorf("expected error for future date of birth, got nil")
	}

	// Height <= 0
	_, err = nutrition.CalculateDailyCalories(nutrition.CalorieCalculationRequest{
		Gender:        "male",
		DateOfBirth:   "1995-05-15",
		HeightCm:      0,
		WeightKg:      80,
		ActivityLevel: "moderate",
	})
	if err == nil {
		t.Errorf("expected error for height <= 0, got nil")
	}

	// Weight <= 0
	_, err = nutrition.CalculateDailyCalories(nutrition.CalorieCalculationRequest{
		Gender:        "male",
		DateOfBirth:   "1995-05-15",
		HeightCm:      170,
		WeightKg:      -5,
		ActivityLevel: "moderate",
	})
	if err == nil {
		t.Errorf("expected error for weight <= 0, got nil")
	}

	// Unknown Gender
	_, err = nutrition.CalculateDailyCalories(nutrition.CalorieCalculationRequest{
		Gender:        "unknown_gender",
		DateOfBirth:   "1995-05-15",
		HeightCm:      170,
		WeightKg:      60,
		ActivityLevel: "moderate",
	})
	if err == nil {
		t.Errorf("expected error for invalid gender, got nil")
	}
}
