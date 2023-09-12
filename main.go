package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

var urlMap = make(map[string]string)
var mu sync.Mutex

type URLMsg struct {
    URL string `json:"url"`
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



