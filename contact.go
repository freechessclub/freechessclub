// Copyright © 2017 Free Chess Club <help@freechess.club>
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
	"net/http"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func handleContact(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if params == nil {
		return
	}

	email := ""
	typ := ""
	msg := ""

	if len(params["email"]) > 0 {
		email = params["email"][0]
	}
	if len(params["type"]) > 0 {
		typ = params["type"][0]
	}
	if len(params["message"]) > 0 {
		msg = params["message"][0]
	}

	if email == "" {
		w.Write([]byte("\"Email address\" not specified"))
		return
	}

	if msg == "" {
		w.Write([]byte("\"Message\" not specified"))
		return
	}

	from := mail.NewEmail(email, email)
	to := mail.NewEmail("Free Chess Club", "feedback@freechess.club")
	m := mail.NewSingleEmail(from, typ, to, msg, "")
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(m)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(response.StatusCode)
		log.Println(response.Body)
		log.Println(response.Headers)
	}
	w.Write([]byte("Message submitted successfully. Thank you!"))
}
