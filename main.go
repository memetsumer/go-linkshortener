package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
    "strings"
	"net/http"
	"sync"
    "github.com/golang-jwt/jwt/v5"
)

var urlMap = make(map[string]string)
var mu sync.Mutex

type URLMsg struct {
    URL string `json:"url"`
}

type User struct {
    ID int `json:"id"`
    Username string `json:"username"`
}

func main() {
    fmt.Println("hello world!")

    http.HandleFunc("/", goRedirect) 
    http.HandleFunc("/shorten", requireAuth(goShorten))
    http.HandleFunc("/login", goLogin)
    http.ListenAndServe(":8000", nil)

}

func goLogin(w http.ResponseWriter, r *http.Request) {
   var user User
   err := json.NewDecoder(r.Body).Decode(&user)
   if err != nil {
       http.Error(w, "Parsing error", http.StatusBadRequest)
       return
   }
   fmt.Println(user)
   token, err := generateToken(user.ID, user.Username)
   if err != nil {
       fmt.Println(err)
       http.Error(w, "Auth failed", http.StatusUnauthorized)
       return
   }
   fmt.Println(token)

   w.Header().Set("Authorization", "Bearer " + token)
}

func goRedirect(rw http.ResponseWriter, req *http.Request) {
    shortURL := req.URL.Path[1:]
    longURL, ok := urlMap[shortURL]

    if !ok {
        http.NotFound(rw, req)
        return
    }

    http.Redirect(rw, req, longURL, http.StatusFound)
}


func goShorten(rw http.ResponseWriter, req *http.Request) {
    var UrlMsg URLMsg
    json.NewDecoder(req.Body).Decode(&UrlMsg)
    

    shortURL := generateShortURL(UrlMsg.URL)

    urlMap[shortURL] = UrlMsg.URL
    fmt.Fprintf(rw, "Shortened URL is http://localhost:8000/%s", shortURL)
}


func generateShortURL(longURL string) string {
    mu.Lock()
    defer mu.Unlock()

    hasher := md5.New()
    io.WriteString(hasher, longURL)
    hash := hex.EncodeToString(hasher.Sum(nil))
    return hash[:6]

}


func generateToken(userID int, username string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "userID": userID,
        "username": username,
    })

    secretKey := []byte("super-secret-key")
    tokenString, err := token.SignedString(secretKey)
    if err != nil {
        return "", err
    }
    return tokenString, nil
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
