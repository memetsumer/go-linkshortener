package main

import (
    "strings"
    "github.com/golang-jwt/jwt/v5"
    "net/http"
    "time"
)


type User struct {
    ID int `json:"id"`
    Username string `json:"username"`
}

func extractTokenFromRequest(r *http.Request) string {
    authHeader := r.Header.Get("Authorization")

    if authHeader == "" {
        return ""
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return ""
    }

    return parts[1]
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenString := extractTokenFromRequest(r)

        if tokenString == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte("super-secret-key"), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    }
}

func generateToken(userID int, username string) (string, error) {
    expiresAt := time.Now().Add(time.Hour)
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "userID": userID,
        "username": username,
        "iat": time.Now().Unix(),
        "exp": expiresAt.Unix(),
    })

    secretKey := []byte("super-secret-key")
    tokenString, err := token.SignedString(secretKey)
    if err != nil {
        return "", err
    }
    return tokenString, nil
}
