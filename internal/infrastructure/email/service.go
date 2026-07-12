package email

import "context"

// EmailService defines the repository/infrastructure layer contract for sending emails.
type EmailService interface {
	// SendVerificationOTP sends a 6-digit verification code to the target email.
	SendVerificationOTP(ctx context.Context, to string, code string) error

	// SendPasswordResetOTP sends a 6-digit password reset OTP code.
	SendPasswordResetOTP(ctx context.Context, to string, code string) error

	// SendWelcomeEmail sends a welcoming email after successful registration.
	SendWelcomeEmail(ctx context.Context, to string, name string) error

	// SendGenericEmail sends a generic custom HTML email.
	SendGenericEmail(ctx context.Context, to string, subject string, htmlContent string) error
}
