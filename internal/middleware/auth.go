package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func AuthMiddleware(next http.Handler) http.Handler {
	setDefaultJWTSecret()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Invalid token: user_id claim missing or not a string", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", userID)
		next.ServeHTTP(w, r)
	})
}

func setDefaultJWTSecret() {
	if len(jwtSecret) == 0 {
		log.Println("WARNING: JWT_SECRET_KEY environment variable not set. Using default insecure key.")
		jwtSecret = []byte("a-string-secret-at-least-256-bits-long")
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header format")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func validateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
