package main

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
)

var (
	log         = logrus.WithField("cmd", "go-freechessclub")
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	runRedis()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/" + r.URL.Path[1:])
	})
	http.HandleFunc("/chat/", handleChat)
	http.HandleFunc("/ws", handleWebsocket)
	log.Println(http.ListenAndServe(":"+port, nil))
}
