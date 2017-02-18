package main

import (
	"net/http"
	"os"
	"time"

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

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/chat/", http.FileServer(http.Dir("./chat/")))
	http.HandleFunc("/ws", handleWebsocket)
	log.Println(http.ListenAndServe(":"+port, nil))
}
