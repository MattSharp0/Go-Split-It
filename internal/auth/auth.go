package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"

	// Cookie names
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
	CSRFTokenCookieName    = "csrf_token"

	// Token types
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrRevokedToken     = errors.New("token revoked")
	ErrInvalidCSRFToken = errors.New("invalid CSRF token")
)

type JWTClaims struct {
	UserID    int64  `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateAccessToken generates a short-lived JWT access token for a user
func GenerateAccessToken(userID int64) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET environment variable is not set")
	}

	expirationMinutes := 30 // default 30 minutes
	if expStr := os.Getenv("ACCESS_TOKEN_EXPIRATION_MINUTES"); expStr != "" {
		if parsed, err := strconv.Atoi(expStr); err == nil {
			expirationMinutes = parsed
		}
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(expirationMinutes) * time.Minute)

	claims := JWTClaims{
		UserID:    userID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a long-lived refresh token string (to be hashed before storage)
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashRefreshToken hashes a refresh token using SHA256 for fast database lookups
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateToken validates a JWT access token and returns the user ID
func ValidateToken(tokenString string) (int64, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, errors.New("JWT_SECRET environment variable is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrExpiredToken
		}
		return 0, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	// Ensure this is an access token
	if claims.TokenType != TokenTypeAccess {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}

// ExtractTokenFromHeader extracts the Bearer token from the Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header must be in format: Bearer <token>")
	}

	return parts[1], nil
}

// SetUserID sets the user ID in the context
func SetUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID gets the user ID from the context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// ValidateRefreshToken validates a refresh token against the database
func ValidateRefreshToken(ctx context.Context, querier db.Querier, token string) (int64, error) {
	// Hash token for lookup
	tokenHash := HashRefreshToken(token)

	// Get token from database
	refreshToken, err := querier.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return 0, ErrInvalidToken
	}

	// Check if token is revoked
	if refreshToken.RevokedAt.Valid {
		return 0, ErrRevokedToken
	}

	// Check if token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		return 0, ErrExpiredToken
	}

	return refreshToken.UserID, nil
}

// RevokeRefreshToken revokes a refresh token in the database
func RevokeRefreshToken(ctx context.Context, querier db.Querier, token string) error {
	tokenHash := HashRefreshToken(token)
	return querier.RevokeRefreshToken(ctx, tokenHash)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func RevokeAllUserTokens(ctx context.Context, querier db.Querier, userID int64) error {
	return querier.RevokeAllUserTokens(ctx, userID)
}

// Cookie helper functions
func SetAuthCookie(w http.ResponseWriter, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   getCookieSecure(),
		SameSite: http.SameSiteStrictMode,
	}
	if domain := os.Getenv("COOKIE_DOMAIN"); domain != "" {
		cookie.Domain = domain
	}
	http.SetCookie(w, cookie)
}

func SetCSRFCookie(w http.ResponseWriter, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     CSRFTokenCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: false, // CSRF token needs to be accessible to JavaScript
		Secure:   getCookieSecure(),
		SameSite: http.SameSiteStrictMode,
	}
	if domain := os.Getenv("COOKIE_DOMAIN"); domain != "" {
		cookie.Domain = domain
	}
	http.SetCookie(w, cookie)
}

func ClearAuthCookies(w http.ResponseWriter) {
	cookies := []string{AccessTokenCookieName, RefreshTokenCookieName, CSRFTokenCookieName}
	for _, name := range cookies {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: name != CSRFTokenCookieName,
			Secure:   getCookieSecure(),
			SameSite: http.SameSiteStrictMode,
		}
		if domain := os.Getenv("COOKIE_DOMAIN"); domain != "" {
			cookie.Domain = domain
		}
		http.SetCookie(w, cookie)
	}
}

func GetTokenFromCookie(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func getCookieSecure() bool {
	secure := os.Getenv("COOKIE_SECURE")
	return secure == "true" || secure == "1"
}

// CSRF token functions
func GenerateCSRFToken(userID int64) (string, error) {
	secret := os.Getenv("CSRF_SECRET")
	if secret == "" {
		// Fallback to JWT_SECRET if CSRF_SECRET is not set
		secret = os.Getenv("JWT_SECRET")
		if secret == "" {
			return "", errors.New("CSRF_SECRET or JWT_SECRET environment variable is not set")
		}
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // CSRF tokens expire in 24 hours

	claims := JWTClaims{
		UserID:    userID,
		TokenType: "csrf",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign CSRF token: %w", err)
	}

	return tokenString, nil
}

func ValidateCSRFToken(tokenString string, userID int64) error {
	secret := os.Getenv("CSRF_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
		if secret == "" {
			return errors.New("CSRF_SECRET or JWT_SECRET environment variable is not set")
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ErrExpiredToken
		}
		return fmt.Errorf("%w: %v", ErrInvalidCSRFToken, err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return ErrInvalidCSRFToken
	}

	// Verify token type and user ID match
	if claims.TokenType != "csrf" {
		return ErrInvalidCSRFToken
	}

	if claims.UserID != userID {
		return ErrInvalidCSRFToken
	}

	return nil
}
