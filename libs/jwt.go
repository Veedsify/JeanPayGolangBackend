package libs

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the custom claims for JWT tokens
type JWTClaims struct {
	ID        uint   `json:"id"`
	UserID    uint32 `json:"user_id"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	TokenID   string `json:"token_id"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey            []byte
	RefreshSecretKey     []byte
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// JWTService handles JWT operations
type JWTService struct {
	config *JWTConfig
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken        string    `json:"access_token"`
	RefreshToken       string    `json:"refresh_token"`
	ExpiresAt          time.Time `json:"expires_at"`
	TokenType          string    `json:"token_type"`
	IsAdmin            bool      `json:"is_admin"` // Optional, can be nil if not applicable
	IsTwoFactorEnabled *bool     `json:"is_two_factor_enabled"`
}

// UserInfo represents user information for token generation
type UserInfo struct {
	ID      uint   `json:"id"`
	UserID  uint32 `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// NewJWTService creates a new JWT service instance
func NewJWTService(config *JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

// NewJWTServiceFromEnv creates JWT service from environment variables
func NewJWTServiceFromEnv() (*JWTService, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET_KEY environment variable is required")
	}

	refreshSecretKey := os.Getenv("JWT_REFRESH_SECRET_KEY")
	if refreshSecretKey == "" {
		refreshSecretKey = secretKey + "_refresh"
	}

	accessTokenDuration := getEnvDurationOrDefault("JWT_ACCESS_TOKEN_DURATION", 60*time.Minute)
	refreshTokenDuration := getEnvDurationOrDefault("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour)
	issuer := getEnvOrDefault("JWT_ISSUER", "JeanPay")

	config := &JWTConfig{
		SecretKey:            []byte(secretKey),
		RefreshSecretKey:     []byte(refreshSecretKey),
		AccessTokenDuration:  accessTokenDuration,
		RefreshTokenDuration: refreshTokenDuration,
		Issuer:               issuer,
	}

	return NewJWTService(config), nil
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(userInfo *UserInfo) (*TokenPair, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token ID: %w", err)
	}

	now := time.Now()
	accessTokenExpiry := now.Add(j.config.AccessTokenDuration)
	refreshTokenExpiry := now.Add(j.config.RefreshTokenDuration)

	// Generate access token
	accessTokenClaims := &JWTClaims{
		ID:        userInfo.ID,
		UserID:    userInfo.UserID,
		Email:     userInfo.Email,
		IsAdmin:   userInfo.IsAdmin,
		TokenID:   tokenID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("%d", userInfo.UserID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(j.config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshTokenClaims := &JWTClaims{
		ID:        userInfo.ID,
		UserID:    userInfo.UserID,
		Email:     userInfo.Email,
		IsAdmin:   userInfo.IsAdmin,
		TokenID:   tokenID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("%d", userInfo.UserID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(j.config.RefreshSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessTokenExpiry,
		TokenType:    "Bearer",
		IsAdmin:      userInfo.IsAdmin,
	}, nil
}

// ValidateAccessToken validates an access token and returns claims
func (j *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, j.config.SecretKey, "access")
}

// ValidateRefreshToken validates a refresh token and returns claims
func (j *JWTService) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, j.config.RefreshSecretKey, "refresh")
}

// RefreshTokens generates new token pair using a valid refresh token
func (j *JWTService) RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userInfo := &UserInfo{
		ID:      claims.ID,
		UserID:  claims.UserID,
		Email:   claims.Email,
		IsAdmin: claims.IsAdmin,
	}

	return j.GenerateTokenPair(userInfo)
}

// ExtractTokenFromHeader extracts token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token is required")
	}

	return token, nil
}

// GetTokenClaims extracts and validates claims from a token string
func (j *JWTService) GetTokenClaims(tokenString string) (*JWTClaims, error) {
	return j.ValidateAccessToken(tokenString)
}

// IsTokenExpired checks if a token is expired
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	_, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		// Check if it's specifically an expiration error
		if errors.Is(err, jwt.ErrTokenExpired) {
			return true
		}
	}
	return false
}

// GeneratePasswordResetToken generates a token for password reset
func (j *JWTService) GeneratePasswordResetToken(id uint, userID uint32, email string) (string, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}

	now := time.Now()
	expiry := now.Add(1 * time.Hour) // Password reset tokens expire in 1 hour

	claims := &JWTClaims{
		ID:        id,
		UserID:    userID,
		Email:     email,
		TokenID:   tokenID,
		TokenType: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.config.SecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign password reset token: %w", err)
	}

	return tokenString, nil
}

// ValidatePasswordResetToken validates a password reset token
func (j *JWTService) ValidatePasswordResetToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, j.config.SecretKey, "password_reset")
}

// GenerateEmailVerificationToken generates a token for email verification
func (j *JWTService) GenerateEmailVerificationToken(ID uint, userID uint32, email string) (string, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}

	now := time.Now()
	expiry := now.Add(24 * time.Hour) // Email verification tokens expire in 24 hours

	claims := &JWTClaims{
		ID:        ID,
		UserID:    userID,
		Email:     email,
		TokenID:   tokenID,
		TokenType: "email_verification",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.config.SecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign email verification token: %w", err)
	}

	return tokenString, nil
}

// ValidateEmailVerificationToken validates an email verification token
func (j *JWTService) ValidateEmailVerificationToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, j.config.SecretKey, "email_verification")
}

// validateToken is a helper method to validate tokens with specific secret and type
func (j *JWTService) validateToken(tokenString string, secret []byte, expectedTokenType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if claims.TokenType != expectedTokenType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedTokenType, claims.TokenType)
	}

	return claims, nil
}

// generateTokenID generates a unique token ID
func generateTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		// Try parsing as seconds if duration parsing fails
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

// TokenBlacklist interface for token blacklisting (implement as needed)
type TokenBlacklist interface {
	IsBlacklisted(tokenID string) bool
	BlacklistToken(tokenID string, expiry time.Time) error
}

// BlacklistToken adds token to blacklist (requires TokenBlacklist implementation)
func (j *JWTService) BlacklistToken(tokenString string, blacklist TokenBlacklist) error {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return fmt.Errorf("failed to validate token for blacklisting: %w", err)
	}

	return blacklist.BlacklistToken(claims.TokenID, claims.ExpiresAt.Time)
}

// IsTokenBlacklisted checks if token is blacklisted (requires TokenBlacklist implementation)
func (j *JWTService) IsTokenBlacklisted(tokenString string, blacklist TokenBlacklist) bool {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return true // Invalid tokens are considered blacklisted
	}

	return blacklist.IsBlacklisted(claims.TokenID)
}
