package notifications

import (
	"context"
	"os"
	"testing"

	"github.com/dsmes/dsmes-backend/internal/config"
)

// Initializes the Firebase Admin client without sending a push.
func TestLocalFCMConfiguration(t *testing.T) {
	if os.Getenv("RUN_LOCAL_FCM_TEST") != "1" {
		t.Skip("set RUN_LOCAL_FCM_TEST=1 to validate local Firebase credentials")
	}
	t.Chdir("../../..")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FCM.CredentialsJSON == "" {
		t.Fatal("FCM credentials are empty")
	}
	sender, err := NewFCMSender(context.Background(), cfg.FCM.CredentialsJSON)
	if err != nil {
		t.Fatal(err)
	}
	if sender == nil {
		t.Fatal("Firebase sender was not created")
	}
}
