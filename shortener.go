package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
	"hash/crc32"
	"encoding/hex"
	"encoding/binary"

	"github.com/garyburd/redigo/redis"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	db, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatalln("Error connecting to redis:", err)
	}
	http.ListenAndServe(":8080", &handler{db})
}

type handler struct{ redis.Conn }

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		start := time.Now()
		h.redirectShortURL(w, r)
		elapsed := time.Since(start)
		log.Printf("Redirection took %s\n",elapsed)

	case "POST":
		start1 := time.Now()
		h.createShortURL(w, r)
		elapsed1 := time.Since(start1)
		log.Printf("Creation took %s\n",elapsed1)
	}
}

func (h *handler) redirectShortURL(w http.ResponseWriter, r *http.Request) {
	url, err := h.Do("GET", r.URL.Path[1:])
	if err != nil || url == nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, string(url.([]byte)), 301)
}

func (h *handler) createShortURL(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL must be provided", 400)
		return
	}
	code := getCode(url)
	_, err := h.Do("SET", code, url)
	if err != nil {
		http.Error(w, "Something failed internally: "+err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "http://localhost:8080/%s\n",code)
}

func getCode(url string) string {
    p := crc32.NewIEEE()
    p.Write([]byte(url))
    v := p.Sum32()
    b := make([]byte, 4)
    binary.LittleEndian.PutUint32(b[:], uint32(v))
    str := hex.EncodeToString(b)
    return str
}
