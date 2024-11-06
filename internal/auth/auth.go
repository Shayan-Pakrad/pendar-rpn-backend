package auth

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shayan-Pakrad/rpn-web-service/internal/db"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func InitKeys() {

	// Set relative paths
	privateKeyRelativePath := filepath.Join("internal", "auth", "keys", "private_key.pem")
	publicKeyRelativePath := filepath.Join("internal", "auth", "keys", "public_key.pem")

	// Load private key
	privateKeyBytes, err := os.ReadFile(privateKeyRelativePath)
	if err != nil {
		log.Fatalf("Error reading private key: %v", err)
	}
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		log.Fatalf("Error parsing private key: %v", err)
	}

	// Load public key
	publicKeyBytes, err := os.ReadFile(publicKeyRelativePath)
	if err != nil {
		log.Fatalf("Error reading public key: %v", err)
	}
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		log.Fatalf("Error parsing public key: %v", err)
	}
}

// JWT Middleware for protecting rpn endpoints
func VerifyJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		// Extract and set claims in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			username, ok := claims["username"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid claims")
			}
			c.Set("user", username)
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid claims")
		}

		// Pass control to the next handler
		return next(c)
	}
}

// Subscription checker middleware
func CheckSubsctiption(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		user, ok := c.Get("user").(string)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user")
		}

		var remaining_requests int
		var expiryDate time.Time

		// Query the user's subscription data from the database
		query := `SELECT remaining_requests, expiry_date FROM subscriptions WHERE username = $1 AND expiry_date > NOW()`
		err := db.DB.QueryRowContext(context.Background(), query, user).Scan(&remaining_requests, &expiryDate)
		if err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusForbidden, "No active subscription found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Error checking subscription")
		}

		// Check if capacity is greater than zero
		if remaining_requests <= 0 {
			return echo.NewHTTPError(http.StatusForbidden, "Subscription remaining_requests exceeded")
		}

		// Check if the subscription is not expired
		if time.Now().After(expiryDate) {
			return echo.NewHTTPError(http.StatusForbidden, "Subscription expired")
		}

		// Decrement the remaining_requests
		_, err = db.DB.ExecContext(context.Background(), `UPDATE subscriptions SET remaining_requests = remaining_requests - 1 WHERE username = $1 AND expiry_date = $2`, user, expiryDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to decrement subscription capacity")
		}

		return next(c)
	}
}

// Generate JWT token
func GenerateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"username": username,
		"exp":      jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 24-hour expiration
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Admin authentication middelware
func AdminMiddleware(username, password string) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(user, pass string, c echo.Context) (bool, error) {
		if user == username && pass == password {
			return true, nil
		}
		return false, nil
	})
}
