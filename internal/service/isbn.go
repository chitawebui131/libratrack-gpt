package service

import (
	"errors"
	"strings"
)

var ErrInvalidISBN = errors.New("invalid ISBN")

// NormalizeISBN removes the prefix "ISBN", hyphens and spaces, upper-cases a final X and checks the length and the check digit.
// An empty input returns "" and no error. ISBN-10 is kept as ISBN-10 (it is NOT converted).
func NormalizeISBN(s string) (string, error) {
	// 1. s = strings.TrimSpace(s). If s == "", return "", nil.
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}

	// 2. If strings.HasPrefix(strings.ToUpper(s), "ISBN"), cut those 4 letters (and an optional ":" after them).
	upperS := strings.ToUpper(s)
	if strings.HasPrefix(upperS, "ISBN") {
		s = s[4:]
		if len(s) > 0 && s[0] == ':' {
			s = s[1:]
		}
		s = strings.TrimSpace(s)
	}

	// 3. Remove every "-" and every whitespace (strings.Map). s = strings.ToUpper(s).
	s = strings.Map(func(r rune) rune {
		if r == '-' || r == ' ' {
			return -1
		}
		return r
	}, s)
	s = strings.ToUpper(s)

	// 4. If len(s) == 10: sum := 0; for i := 0; i < 10; i++ { c := s[i]; v := 0; if c == 'X' && i == 9 { v = 10 } else if c >= '0' && c <= '9' { v = int(c - '0') } else { return "", ErrInvalidISBN }; sum += (10 - i) * v }. If sum%11 != 0 return "", ErrInvalidISBN. Return s, nil.
	if len(s) == 10 {
		sum := 0
		for i := 0; i < 10; i++ {
			c := s[i]
			v := 0
			if c == 'X' && i == 9 {
				v = 10
			} else if c >= '0' && c <= '9' {
				v = int(c - '0')
			} else {
				return "", ErrInvalidISBN
			}
			sum += (10 - i) * v
		}
		if sum%11 != 0 {
			return "", ErrInvalidISBN
		}
		return s, nil
	}

	// 5. If len(s) == 13: sum := 0; for i := 0; i < 13; i++ { c := s[i]; if c < '0' || c > '9' { return "", ErrInvalidISBN }; d := int(c - '0'); if i%2 == 1 { d *= 3 }; sum += d }. If sum%10 != 0 return "", ErrInvalidISBN. Return s, nil.
	if len(s) == 13 {
		sum := 0
		for i := 0; i < 13; i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return "", ErrInvalidISBN
			}
			d := int(c - '0')
			if i%2 == 1 {
				d *= 3
			}
			sum += d
		}
		if sum%10 != 0 {
			return "", ErrInvalidISBN
		}
		return s, nil
	}

	// 6. Return "", ErrInvalidISBN.
	return "", ErrInvalidISBN
}
