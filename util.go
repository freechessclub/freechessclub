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
	"strconv"
)

func atoi(b []byte) int {
	i, _ := strconv.Atoi(string(b))
	return i
}

func style12ToFEN(b []byte) string {
	str := string(b[:])
	var fen string
	count := 0
	for i := 0; i < 8; i++ {
		if str[i] == '-' {
			count++
			if i == 7 {
				fen += strconv.Itoa(count)
			}
		} else {
			if count > 0 {
				fen += strconv.Itoa(count)
				count = 0
			}
			fen += string(str[i])
		}
	}
	return fen
}
