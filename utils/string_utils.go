// Copyright (C) 2026 Jan Mewes
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func ToKebabCase(s string) string {
	runes := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsLetter(r) {
			if unicode.IsUpper(r) {
				runes = append(runes, '-')
				runes = append(runes, unicode.ToLower(r))
			} else {
				runes = append(runes, r)
			}
		} else if unicode.IsDigit(r) {
			runes = append(runes, r)
		} else {
			runes = append(runes, '-')
		}
	}

	result := strings.Trim(string(runes), "-")
	result = regexp.MustCompile(`-+`).ReplaceAllString(result, "-")
	return result
}
