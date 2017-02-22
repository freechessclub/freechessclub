// Copyright © 2017 The Free Chess Club <help@freechess.club>
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
)

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
	log.Println(contact)
	w.Write([]byte("Message submitted successfully. Thank you!"))
}
