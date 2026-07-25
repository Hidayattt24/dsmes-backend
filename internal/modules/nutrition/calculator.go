package nutrition

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

// ParseDOB parses date strings in multiple formats (YYYY-MM-DD, RFC3339, DD-MM-YYYY).
func ParseDOB(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date of birth is empty")
	}

	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"02-01-2006",
		"02/01/2006",
	}

	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// CalculateAge computes user age in years from birthdate.
func CalculateAge(dob time.Time) (int, error) {
	now := time.Now()
	if dob.After(now) {
		return 0, fmt.Errorf("date of birth cannot be in the future")
	}

	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}

	if age < 0 || age > 120 {
		return 0, fmt.Errorf("calculated age %d is out of valid range (0-120)", age)
	}

	return age, nil
}

// NormalizeGender validates and normalizes gender into standard "male" or "female".
func NormalizeGender(gender string) (string, bool, error) {
	g := strings.ToLower(strings.TrimSpace(gender))
	switch g {
	case "male", "l", "laki-laki", "pria", "m":
		return "male", true, nil
	case "female", "p", "perempuan", "wanita", "f":
		return "female", false, nil
	default:
		return "", false, errs.NewBadRequest("invalid gender: must be 'male' or 'female'")
	}
}

// NormalizeActivityLevel resolves activity level name into multiplier factor.
func NormalizeActivityLevel(level string) (float64, error) {
	l := strings.ToLower(strings.TrimSpace(level))
	switch l {
	case "very low", "sangat rendah", "1.2", "1.20":
		return 1.20, nil
	case "light", "ringan", "1.375":
		return 1.375, nil
	case "moderate", "sedang", "1.55":
		return 1.55, nil
	case "high", "aktif", "1.725":
		return 1.725, nil
	case "very high", "sangat aktif", "1.9", "1.90":
		return 1.90, nil
	default:
		return 0, errs.NewBadRequest("invalid activity level: must be one of 'Very Low', 'Light', 'Moderate', 'High', 'Very High'")
	}
}

// CalculateDailyCalories executes Mifflin-St Jeor equation and TDEE multiplier targets.
// Note: blood_type is intentionally ignored in calculation as required.
func CalculateDailyCalories(req CalorieCalculationRequest) (*CalorieCalculationResponse, error) {
	// 1. Validation: Height & Weight
	if req.HeightCm <= 0 {
		return nil, errs.NewBadRequest("height must be greater than 0 cm")
	}
	if req.WeightKg <= 0 {
		return nil, errs.NewBadRequest("weight must be greater than 0 kg")
	}

	// 2. Validate & Normalize Gender
	normalizedGender, isMale, err := NormalizeGender(req.Gender)
	if err != nil {
		return nil, err
	}

	// 3. Resolve Age
	var age int
	if strings.TrimSpace(req.DateOfBirth) != "" {
		dob, parseErr := ParseDOB(req.DateOfBirth)
		if parseErr != nil {
			return nil, errs.NewBadRequest(parseErr.Error())
		}
		calcAge, ageErr := CalculateAge(dob)
		if ageErr != nil {
			return nil, errs.NewBadRequest(ageErr.Error())
		}
		age = calcAge
	} else if req.Age > 0 && req.Age <= 120 {
		age = req.Age
	} else {
		return nil, errs.NewBadRequest("valid date_of_birth or age is required")
	}

	// 4. Resolve Activity Multiplier
	multiplier, err := NormalizeActivityLevel(req.ActivityLevel)
	if err != nil {
		return nil, err
	}

	// 5. Calculate BMR (Mifflin-St Jeor Equation)
	// Male: BMR = (10 × W) + (6.25 × H) - (5 × A) + 5
	// Female: BMR = (10 × W) + (6.25 × H) - (5 × A) - 161
	var bmrFloat float64
	if isMale {
		bmrFloat = (10 * req.WeightKg) + (6.25 * req.HeightCm) - (5 * float64(age)) + 5
	} else {
		bmrFloat = (10 * req.WeightKg) + (6.25 * req.HeightCm) - (5 * float64(age)) - 161
	}
	bmr := int(math.Round(bmrFloat))

	// 6. Calculate TDEE
	tdeeFloat := bmrFloat * multiplier
	tdee := int(math.Round(tdeeFloat))

	// 7. Calculate Recommended Calorie Targets with safety clamping
	// Weight Loss: TDEE - 500 (Min Female: 1200, Min Male: 1500)
	minWeightLossCal := 1200
	if isMale {
		minWeightLossCal = 1500
	}

	weightLoss := tdee - 500
	if weightLoss < minWeightLossCal {
		weightLoss = minWeightLossCal
	}

	maintenance := tdee
	weightGain := tdee + 500

	return &CalorieCalculationResponse{
		Age:                age,
		Gender:             normalizedGender,
		Height:             req.HeightCm,
		Weight:             req.WeightKg,
		BMR:                bmr,
		ActivityMultiplier: multiplier,
		TDEE:               tdee,
		RecommendedCalories: RecommendedCalories{
			WeightLoss:  weightLoss,
			Maintenance: maintenance,
			WeightGain:  weightGain,
		},
	}, nil
}
