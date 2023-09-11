package main

import (
    "fmt"
    "net/http"
    "sync"
    "crypto/md5"
    "io"
    "encoding/hex"
    "encoding/json"
)

var urlMap = make(map[string]string)
var mu sync.Mutex

type URLMsg struct {
    URL string `json:"url"`
}

func main() {
    fmt.Println("hello world!")

    http.HandleFunc("/", goRedirect) 
    http.HandleFunc("/shorten", goShorten)
    http.ListenAndServe(":8000", nil)

}

func goRedirect(rw http.ResponseWriter, req *http.Request) {
    fmt.Println(req.URL.Path)
    shortURL := req.URL.Path[1:]
    longURL, ok := urlMap[shortURL]

    if !ok {
        http.NotFound(rw, req)
        return
    }

    http.Redirect(rw, req, longURL, http.StatusFound)
}

func generateShortURL(longURL string) string {
    mu.Lock()
    defer mu.Unlock()

    hasher := md5.New()
    io.WriteString(hasher, longURL)
    hash := hex.EncodeToString(hasher.Sum(nil))
    return hash[:6]
}


func goShorten(rw http.ResponseWriter, req *http.Request) {
    var UrlMsg URLMsg
    json.NewDecoder(req.Body).Decode(&UrlMsg)
    

    shortURL := generateShortURL(UrlMsg.URL)

    urlMap[shortURL] = UrlMsg.URL
    fmt.Fprintf(rw, "Shortened URL is http://localhost:8000/%s", shortURL)
}

