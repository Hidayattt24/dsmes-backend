package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
)

// mockAuthRepo implements AuthRepository for unit tests
type mockAuthRepo struct {
	mu          sync.Mutex
	sessions    map[string]*AuthSession
	findErr     error
	rotateErr   error
	rotateCalls int
	findCalls   int
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		sessions: make(map[string]*AuthSession),
	}
}

func (m *mockAuthRepo) FindStaffByEmail(ctx context.Context, email string) (*StaffAccount, error) {
	return nil, nil
}
func (m *mockAuthRepo) FindStaffByID(ctx context.Context, id string) (*StaffAccount, error) {
	return nil, nil
}
func (m *mockAuthRepo) UpdateStaffPassword(ctx context.Context, id, hash string) error {
	return nil
}
func (m *mockAuthRepo) FindPatientByEmail(ctx context.Context, email string) (*domain.Patient, error) {
	return nil, nil
}
func (m *mockAuthRepo) FindPatientByPhoneNumber(ctx context.Context, phone string) (*domain.Patient, error) {
	return nil, nil
}
func (m *mockAuthRepo) FindPatientByID(ctx context.Context, id string) (*domain.Patient, error) {
	return nil, nil
}
func (m *mockAuthRepo) UpdatePatientPassword(ctx context.Context, id, hash string) error {
	return nil
}
func (m *mockAuthRepo) CreateResetToken(ctx context.Context, token *PasswordResetToken) error {
	return nil
}
func (m *mockAuthRepo) FindActiveResetTokens(ctx context.Context, email string, ownerType OwnerType) ([]PasswordResetToken, error) {
	return nil, nil
}
func (m *mockAuthRepo) MarkTokenUsed(ctx context.Context, id string) error {
	return nil
}
func (m *mockAuthRepo) CreateSession(ctx context.Context, session *AuthSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.RefreshToken] = session
	return nil
}
func (m *mockAuthRepo) RecordPatientLogin(ctx context.Context, login *domain.PatientLogin) error {
	return nil
}
func (m *mockAuthRepo) FindSession(ctx context.Context, refreshToken string) (*AuthSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.findCalls++
	if m.findErr != nil {
		return nil, m.findErr
	}
	s, ok := m.sessions[refreshToken]
	if !ok || time.Now().After(s.ExpiresAt) {
		return nil, errs.NewNotFound("session not found or expired")
	}
	return s, nil
}
func (m *mockAuthRepo) RevokeSession(ctx context.Context, refreshToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, refreshToken)
	return nil
}
func (m *mockAuthRepo) RevokeAllSessions(ctx context.Context, ownerType OwnerType, ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.sessions {
		if v.OwnerType == ownerType && v.OwnerID == ownerID {
			delete(m.sessions, k)
		}
	}
	return nil
}
func (m *mockAuthRepo) RotateSession(ctx context.Context, oldRefreshTokenHash string, newSession *AuthSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rotateCalls++
	if m.rotateErr != nil {
		return m.rotateErr
	}
	if _, ok := m.sessions[oldRefreshTokenHash]; !ok {
		return errs.NewNotFound("session not found or expired")
	}
	delete(m.sessions, oldRefreshTokenHash)
	m.sessions[newSession.RefreshToken] = newSession
	return nil
}

func setupTestService(repo AuthRepository) (AuthService, *jwtpkg.Manager) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:          "very-secure-test-secret-at-least-32-chars-long",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "dsmes-test",
		},
	}
	jwtMgr := jwtpkg.NewManager(cfg)
	log := zap.NewNop()
	svc := NewAuthService(repo, jwtMgr, nil, log)
	return svc, jwtMgr
}

func TestHashRefreshToken(t *testing.T) {
	a := HashRefreshToken("some-random-token")
	b := HashRefreshToken("some-random-token")
	if a != b {
		t.Error("expected HashRefreshToken to be deterministic")
	}
	if len(a) != 64 {
		t.Errorf("expected 64-char hex digest, got %d chars: %s", len(a), a)
	}
	c := HashRefreshToken("some-other-token")
	if a == c {
		t.Error("expected different tokens to produce different hashes")
	}
}

func TestRefreshToken_Success(t *testing.T) {
	repo := newMockAuthRepo()
	svc, jwtMgr := setupTestService(repo)

	tokens, err := jwtMgr.GenerateTokenPair("user-123", "user@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	hashT1 := HashRefreshToken(tokens.RefreshToken)
	repo.sessions[hashT1] = &AuthSession{
		OwnerType:    OwnerTypePatient,
		OwnerID:      "user-123",
		RefreshToken: hashT1,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	resp, err := svc.RefreshToken(context.Background(), RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("expected successful refresh, got error: %v", err)
	}

	if resp.Tokens.AccessToken == "" || resp.Tokens.RefreshToken == "" {
		t.Error("expected non-empty rotated tokens")
	}

	// Verify old session hash was removed and new session hash was created
	if _, ok := repo.sessions[hashT1]; ok {
		t.Error("expected old session hash to be revoked in database")
	}
	hashT2 := HashRefreshToken(resp.Tokens.RefreshToken)
	if _, ok := repo.sessions[hashT2]; !ok {
		t.Error("expected new session hash to be present in database")
	}
}

func TestRefreshToken_ExpiredOrNotFound(t *testing.T) {
	repo := newMockAuthRepo()
	svc, jwtMgr := setupTestService(repo)

	tokens, err := jwtMgr.GenerateTokenPair("user-123", "user@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	// Do NOT add to repo.sessions -> record not found
	_, err = svc.RefreshToken(context.Background(), RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	})
	if err == nil {
		t.Fatal("expected error for missing session, got nil")
	}

	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code != 401 {
		t.Errorf("expected 401 Unauthorized for missing session, got: %v", err)
	}
}

func TestRefreshToken_DatabaseError_Returns500(t *testing.T) {
	repo := newMockAuthRepo()
	repo.findErr = errs.NewInternal("database connection refused", errors.New("conn lost"))
	svc, jwtMgr := setupTestService(repo)

	tokens, err := jwtMgr.GenerateTokenPair("user-123", "user@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	_, err = svc.RefreshToken(context.Background(), RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	})
	if err == nil {
		t.Fatal("expected error for database failure, got nil")
	}

	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code != 500 {
		t.Errorf("expected 500 InternalServerError for DB disruption, got: %v", err)
	}
}

func TestRefreshToken_LostResponse_IdempotentReplay(t *testing.T) {
	repo := newMockAuthRepo()
	svc, jwtMgr := setupTestService(repo)

	tokens, err := jwtMgr.GenerateTokenPair("user-123", "user@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	hashT1 := HashRefreshToken(tokens.RefreshToken)
	repo.sessions[hashT1] = &AuthSession{
		OwnerType:    OwnerTypePatient,
		OwnerID:      "user-123",
		RefreshToken: hashT1,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	// First call rotates T1 -> T2
	resp1, err := svc.RefreshToken(context.Background(), RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("first refresh failed: %v", err)
	}

	// Client lost HTTP response and retries with T1 within 15s window
	resp2, err := svc.RefreshToken(context.Background(), RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("expected idempotent replay to succeed, got: %v", err)
	}

	// Must return the exact same rotated token pair
	if resp1.Tokens.AccessToken != resp2.Tokens.AccessToken {
		t.Error("expected idempotent access token to match original rotation")
	}
	if resp1.Tokens.RefreshToken != resp2.Tokens.RefreshToken {
		t.Error("expected idempotent refresh token to match original rotation")
	}

	// DB rotate was called only once!
	if repo.rotateCalls != 1 {
		t.Errorf("expected 1 DB rotate call, got %d", repo.rotateCalls)
	}
}

func TestRefreshToken_ConcurrentCalls(t *testing.T) {
	repo := newMockAuthRepo()
	svc, jwtMgr := setupTestService(repo)

	tokens, err := jwtMgr.GenerateTokenPair("user-123", "user@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	hashT1 := HashRefreshToken(tokens.RefreshToken)
	repo.sessions[hashT1] = &AuthSession{
		OwnerType:    OwnerTypePatient,
		OwnerID:      "user-123",
		RefreshToken: hashT1,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	// Run 5 concurrent refresh requests with the same token
	var wg sync.WaitGroup
	results := make([]*TokenResponse, 5)
	errsList := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r, e := svc.RefreshToken(context.Background(), RefreshTokenRequest{
				RefreshToken: tokens.RefreshToken,
			})
			results[idx] = r
			errsList[idx] = e
		}(i)
	}

	wg.Wait()

	// Every concurrent request must succeed
	for i := 0; i < 5; i++ {
		if errsList[i] != nil {
			t.Errorf("concurrent call %d failed: %v", i, errsList[i])
		}
		if results[i] == nil || results[i].Tokens.AccessToken == "" {
			t.Errorf("concurrent call %d returned empty tokens", i)
		}
	}

	// Database RotateSession was called only once!
	if repo.rotateCalls != 1 {
		t.Errorf("expected exactly 1 DB rotation call, got %d", repo.rotateCalls)
	}
}

func TestRefreshToken_ReplayAfterWindow_Fails(t *testing.T) {
	repo := newMockAuthRepo()
	svc, jwtMgr := setupTestService(repo)

	tokens, err := jwtMgr.GenerateTokenPair("user-123", "user@example.com", "user")
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	hashT1 := HashRefreshToken(tokens.RefreshToken)
	repo.sessions[hashT1] = &AuthSession{
		OwnerType:    OwnerTypePatient,
		OwnerID:      "user-123",
		RefreshToken: hashT1,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	// First call rotates
	_, err = svc.RefreshToken(context.Background(), RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	})
	if err != nil {
		t.Fatalf("first refresh failed: %v", err)
	}

	// Simulate passage of time past the 15-second replay window
	authSvcImpl := svc.(*authService)
	authSvcImpl.replayMu.Lock()
	entry := authSvcImpl.replayBuffer[hashT1]
	entry.expiresAt = time.Now().Add(-1 * time.Second) // expired
	authSvcImpl.replayBuffer[hashT1] = entry
	authSvcImpl.replayMu.Unlock()

	// Second call past the replay window must be rejected with 401
	_, err = svc.RefreshToken(context.Background(), RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	})
	if err == nil {
		t.Fatal("expected expired replay token to be rejected, got nil")
	}

	var appErr *errs.AppError
	if !errors.As(err, &appErr) || appErr.Code != 401 {
		t.Errorf("expected 401 Unauthorized for expired replay token, got: %v", err)
	}
}
