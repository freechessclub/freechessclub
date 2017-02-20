package main

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
)

var (
	log = logrus.New()
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("Using default port 8080")
	}

	http.HandleFunc("/ws", handleWebsocket)
	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/privacy.html")
	})
	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/contact.html")
	})
	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/play.html")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/" + r.URL.Path[1:])
	})
	log.Println(http.ListenAndServe(":"+port, nil))
}
