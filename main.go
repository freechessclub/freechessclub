// Copyright © 2019 Free Chess Club <help@freechess.club>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
)

var (
	log      = logrus.New()
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkSameOrigin,
	}
)

func checkSameOrigin(r *http.Request) bool {
	if strings.Contains(r.UserAgent(), "Free Chess Club") {
		return true
	}

	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	return u.Host == r.Host
}

func redirectToHTTPSRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proto := r.Header.Get("X-Forwarded-Proto")
		if proto == "http" || proto == "HTTP" {
			http.Redirect(w, r, fmt.Sprintf("https://%s%s", r.Host, r.URL), http.StatusPermanentRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("Using default port 8080")
	}

	root := os.Getenv("WEB_ROOT")
	http.HandleFunc("/ws", handleWebsocket)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "public/www/img/favicon.ico"))
	})
	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "public/privacy.html"))
	})
	http.HandleFunc("/contact/submit", handleContact)
	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "public/contact.html"))
	})
	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "public/play.html"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "public/", r.URL.Path[1:]))
	})
	w := log.Writer()
	defer w.Close()
	loggingRouter := handlers.LoggingHandler(w, http.DefaultServeMux)
	httpsRouter := redirectToHTTPSRouter(loggingRouter)

	log.Println(http.ListenAndServe(":"+port, httpsRouter))
}
