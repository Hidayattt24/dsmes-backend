package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/resend/resend-go/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type resendEmailService struct {
	client    *resend.Client
	fromEmail string
	templates *TemplateEngine
	log       *zap.Logger
}

// NewResendEmailService constructs a Resend-backed EmailService.
func NewResendEmailService(cfg *config.Config, log *zap.Logger) EmailService {
	// Root execution is relative to backend directory, so "templates" is the default path
	tmplEngine := NewTemplateEngine("templates")

	client := resend.NewClient(cfg.Email.ResendAPIKey)

	return &resendEmailService{
		client:    client,
		fromEmail: cfg.Email.ResendFromEmail,
		templates: tmplEngine,
		log:       log,
	}
}

// SendVerificationOTP implements EmailService.
func (s *resendEmailService) SendVerificationOTP(ctx context.Context, to string, code string) error {
	const templateName = "verification.html"
	s.log.Info("email: rendering verification template", zap.String("recipient", to))

	html, err := s.templates.Render(templateName, map[string]any{
		"Code":       code,
		"ExpiresMin": 5,
	})
	if err != nil {
		s.log.Error("email: failed to render template", zap.String("template", templateName), zap.Error(err))
		return errs.NewInternal("failed to generate verification email template", err)
	}

	return s.send(ctx, to, "Kode Verifikasi Akun DSMES Aceh", html, templateName)
}

// SendPasswordResetOTP implements EmailService.
func (s *resendEmailService) SendPasswordResetOTP(ctx context.Context, to string, code string) error {
	const templateName = "forgot-password.html"
	s.log.Info("email: rendering forgot-password template", zap.String("recipient", to))

	html, err := s.templates.Render(templateName, map[string]any{
		"Code":       code,
		"ExpiresMin": 5,
	})
	if err != nil {
		s.log.Error("email: failed to render template", zap.String("template", templateName), zap.Error(err))
		return errs.NewInternal("failed to generate password reset email template", err)
	}

	return s.send(ctx, to, "Atur Ulang Kata Sandi DSMES Aceh", html, templateName)
}

// SendWelcomeEmail implements EmailService.
func (s *resendEmailService) SendWelcomeEmail(ctx context.Context, to string, name string) error {
	const templateName = "welcome.html"
	s.log.Info("email: rendering welcome template", zap.String("recipient", to))

	html, err := s.templates.Render(templateName, map[string]any{
		"Name": name,
	})
	if err != nil {
		s.log.Error("email: failed to render template", zap.String("template", templateName), zap.Error(err))
		return errs.NewInternal("failed to generate welcome email template", err)
	}

	return s.send(ctx, to, "Selamat Datang di DSMES Aceh", html, templateName)
}

// SendGenericEmail implements EmailService.
func (s *resendEmailService) SendGenericEmail(ctx context.Context, to string, subject string, htmlContent string) error {
	return s.send(ctx, to, subject, htmlContent, "generic")
}

// send wraps the Resend Go SDK client call, implements detailed error mapping and security logging.
func (s *resendEmailService) send(ctx context.Context, to, subject, html, templateName string) error {
	// Guard: check if API key or from email is missing
	if s.fromEmail == "" {
		s.log.Error("email: cannot send email, RESEND_FROM_EMAIL is not configured")
		return errs.NewInternal("email service misconfigured: missing sender email", nil)
	}

	from := s.fromEmail
	if !strings.Contains(from, "<") {
		from = fmt.Sprintf("DSMES Aceh <%s>", from)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	s.log.Info("email: sending email via Resend",
		zap.String("recipient", to),
		zap.String("template", templateName),
	)

	// Trigger API call
	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		errStr := err.Error()
		s.log.Error("email: failed to send email via Resend",
			zap.String("recipient", to),
			zap.String("template", templateName),
			zap.Error(err),
		)

		// Map Resend error code responses appropriately
		if strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "API key") {
			return errs.NewUnauthorized("email service authentication failed: invalid API key")
		}
		if strings.Contains(errStr, "429") || strings.Contains(errStr, "Rate limit") {
			return errs.NewInternal("email service rate limit exceeded", err)
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return errs.NewInternal("email service request timed out", err)
		}

		return errs.NewInternal(fmt.Sprintf("failed to deliver email: %s", errStr), err)
	}

	s.log.Info("email: sent successfully",
		zap.String("recipient", to),
		zap.String("template", templateName),
		zap.String("resend_id", sent.Id),
	)

	return nil
}
