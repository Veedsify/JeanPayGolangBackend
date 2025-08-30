package initializers

import (
	"os"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func InitializeGoogleAuth() {
	GOOGLE_CLIENT_ID := os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_CLIENT_SECRET := os.Getenv("GOOGLE_CLIENT_SECRET")
	GOOGLE_CALLBACK_URL := libs.GetEnvOrDefault("GOOGLE_CALLBACK_URL", "http://localhost:8080/api/auth/google/callback")
	if GOOGLE_CLIENT_ID == "" || GOOGLE_CLIENT_SECRET == "" {
		panic("Google OAuth credentials are not set in environment variables")
	}
	goth.UseProviders(
		google.New(GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_CALLBACK_URL, "email", "profile"),
	)
}
