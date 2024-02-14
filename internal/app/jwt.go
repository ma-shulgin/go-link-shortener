package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ma-shulgin/go-link-shortener/internal/app-context"
)

var jwtSecret []byte

const authCookieName string = "auth_token"

func InitializeJWT(secret string) {
	jwtSecret = []byte(secret)
}

func GenerateJWT(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)

	return tokenString, err
}

func ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrHashUnavailable
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}

func JwtAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authCookieName)
		if err != nil || cookie.Value == "" {
			userID, err := GenerateRandomUserID(4)
			if err != nil {
				http.Error(w, "Could not generate userID", http.StatusInternalServerError)
				return
			}
			tokenString, err := GenerateJWT(userID)
			if err != nil {
				http.Error(w, "Could not generate token", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     authCookieName,
				Value:    tokenString,
				HttpOnly: true,
			})
			ctx := context.WithValue(r.Context(), appContext.KeyUserID, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		} else {
			claims, err := ValidateJWT(cookie.Value)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), appContext.KeyUserID, claims.Subject)
			h.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func GenerateRandomUserID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
