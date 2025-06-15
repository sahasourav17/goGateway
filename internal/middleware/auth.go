package middleware

import (
	// "context" is no longer needed here for passing the user ID
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func AuthMiddleware(next http.Handler) http.Handler {
	if len(jwtSecret) == 0 {
		log.Println("WARNING: JWT_SECRET_KEY environment variable not set. Using default insecure key.")
		jwtSecret = []byte("a-string-secret-at-least-256-bits-long")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, claimOK := claims["user_id"].(string)
			if !claimOK {
				http.Error(w, "Invalid token: user_id claim missing or not a string", http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-User-ID", userID)
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	})
}
