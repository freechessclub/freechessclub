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
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Contact information
type Contact struct {
	Email string `json:"email"`
	Type  string `json:"type"`
	Msg   string `json:"message"`
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	var contact Contact
	d, _ := httputil.DumpRequest(r, true)
	log.Println(string(d))
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&contact)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()
	from := mail.NewEmail(contact.Email, contact.Email)
	to := mail.NewEmail("Free Chess Club", "feedback@freechess.club")
	content := mail.NewContent("text/plain", contact.Msg)
	m := mail.NewV3MailInit(from, contact.Type, to, content)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(response.StatusCode)
		log.Println(response.Body)
		log.Println(response.Headers)
	}
	w.Write([]byte("Message submitted successfully. Thank you!"))
}
