package util

import (
	"strings"
	"unicode"

	"github.com/gnames/gnlib"
)

func ToTitleCaseWord(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s) // Convert string to slice of runes to handle Unicode characters correctly

	// Capitalize the first rune
	runes[0] = unicode.ToTitle(runes[0])

	// Lowercase the rest of the runes
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes) // Convert back to string
}

// NormalizeAuthors normalizes authors from "Doe, J, Berg, JM Jr & Smith, 1988"
// to parseable by gnparser "J. Doe, J.M. Berg & Smith, 1988"
func NormalizeAuthors(au string) string {
	au = strings.Replace(au, " & ", ", ", 1)
	aus := strings.Split(au, ",")

	l := len(aus)
	if l == 1 {
		return au
	}

	aus = gnlib.Map(aus, func(s string) string {
		return strings.TrimSpace(s)
	})

	var res []string
	for i := 0; i < l; i++ {
		if i+1 == l {
			res = append(res, aus[i])
			break
		}
		if isInitials(aus[i+1]) {
			init, suff := toInitials(aus[i+1])
			if aus[i][0] == '(' {
				au = "(" + init + " " + aus[i][1:] + " " + suff
			} else {
				au = init + " " + aus[i] + " " + suff
			}
			res = append(res, strings.TrimSpace(au))
			i++
			continue
		}
		res = append(res, aus[i])
	}

	if len(res) == 1 {
		return res[0]
	}

	yr := ""
	if isYear(res[len(res)-1]) {
		yr = res[len(res)-1]
		res = res[0 : len(res)-1]
		if len(res) == 1 {
			return res[0] + ", " + yr
		}
	}

	au = strings.Join(res[0:len(res)-1], ", ")
	au = au + " & " + res[len(res)-1]

	if yr != "" {
		au = au + ", " + yr
	}
	return au
}

func isInitials(s string) bool {
	ss := strings.Split(s, " ")
	if ss[0][0] == '1' || ss[0][0] == '2' {
		return false
	}
	if len(ss[0]) < 5 && ss[0] == strings.ToUpper(ss[0]) {
		return true
	}
	return false
}

func toInitials(s string) (string, string) {
	ss := strings.Split(s, " ")
	init := ss[0]
	inits := strings.Split(init, "")
	init = strings.Join(inits, ".")
	init = init + "."
	suff := ""
	if len(ss) > 1 {
		suff = strings.Join(ss[1:], "")
		suff = strings.TrimSpace(suff)
	}
	return init, suff
}

func isYear(s string) bool {
	if len(s) == 0 {
		return false
	}

	firstChar := s[0]

	return firstChar == '1' || firstChar == '2'
}
