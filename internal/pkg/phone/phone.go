package phone

import (
	"errors"
	"strings"
)

var (
	ErrEmptyPhone    = errors.New("nomor handphone tidak boleh kosong")
	ErrInvalidPhone  = errors.New("nomor handphone harus berupa format Indonesia yang valid (contoh: 081234567890 atau +6281234567890)")
	ErrInvalidLength = errors.New("panjang nomor handphone harus antara 10 hingga 15 digit")
)

// Normalize converts Indonesian phone number formats (08..., 628..., +628...) into standardized digit format (628...).
func Normalize(phone string) (string, error) {
	// Remove non-digit characters (spaces, dashes, +, etc.)
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	if cleaned == "" {
		return "", ErrEmptyPhone
	}

	// Convert leading 0 to 62, or prepend 62 if input starts directly with 8
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "62" + cleaned[1:]
	} else if strings.HasPrefix(cleaned, "8") {
		cleaned = "62" + cleaned
	}

	// Check if prefix is 62
	if !strings.HasPrefix(cleaned, "62") {
		return "", ErrInvalidPhone
	}

	// Validate overall digit length (typically 10 to 15 digits for Indonesian numbers)
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return "", ErrInvalidLength
	}

	return cleaned, nil
}
