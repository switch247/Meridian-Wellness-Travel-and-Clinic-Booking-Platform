package security

import (
	"regexp"
	"strings"
)

var (
	streetReplacements = map[string]string{
		" street":    " st",
		" st.":       " st",
		" avenue":    " ave",
		" ave.":      " ave",
		" road":      " rd",
		" rd.":       " rd",
		" boulevard": " blvd",
		" blvd.":     " blvd",
	}
	spaceRe = regexp.MustCompile(`\s+`)
)

func NormalizeUSAddress(line1, city, state, postal string) string {
	norm := strings.ToLower(strings.TrimSpace(line1 + " " + city + " " + state + " " + postal))
	norm = spaceRe.ReplaceAllString(norm, " ")
	for old, newV := range streetReplacements {
		norm = strings.ReplaceAll(norm, old, newV)
	}
	return strings.TrimSpace(norm)
}

func InCoverage(postal string, allowed []string) bool {
	for _, code := range allowed {
		if strings.TrimSpace(code) == strings.TrimSpace(postal) {
			return true
		}
	}
	return false
}
