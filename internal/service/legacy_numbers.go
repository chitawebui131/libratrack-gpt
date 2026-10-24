package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var inventarRange = regexp.MustCompile(`^(\d+)-(\d+)$`)

// parseCount parses countBooks. An empty value means 1.
func parseCount(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 100000 {
		return 0, errors.New("countBooks must be a whole number from 1 to 100000")
	}
	return n, nil
}

// parseYear returns the year, or 0 plus a warning when the value is empty or invalid.
func parseYear(s string) (int, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ""
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1000 || n > 2100 {
		return 0, "year ignored: " + s
	}
	return n, ""
}

// parseInt0 returns the number, or 0 for an empty or invalid value.
func parseInt0(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// parseInventar expands the inventory cell into one number per copy.
//   - spaces inside a number are removed: "28 935" -> "28935";
//   - a range "22282-22298" expands to every number in it;
//   - several values may be separated with ";";
//   - when there is exactly ONE number but count > 1, numbers base-1 ... base-count are generated and a warning is returned;
//   - when the number of numbers differs from count, a warning is returned and the numbers from the cell are used.
func parseInventar(s string, count int) (list []string, warnings []string, err error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil, errors.New("inventar is empty")
	}

	for _, token := range strings.Split(s, ";") {
		t := strings.Join(strings.Fields(token), "")
		if t == "" {
			continue
		}

		m := inventarRange.FindStringSubmatch(t)
		if m != nil {
			a, _ := strconv.Atoi(m[1])
			b, _ := strconv.Atoi(m[2])
			if b < a || b-a >= 100000 {
				return nil, nil, errors.New("invalid inventar range: " + t)
			}
			for n := a; n <= b; n++ {
				list = append(list, strconv.Itoa(n))
			}
		} else {
			list = append(list, t)
		}
	}

	if len(list) == 0 {
		return nil, nil, errors.New("inventar is empty")
	}

	if len(list) == 1 && count > 1 {
		base := list[0]
		list = nil
		for i := 1; i <= count; i++ {
			list = append(list, fmt.Sprintf("%s-%d", base, i))
		}
		warnings = append(warnings, "inventar expanded to "+strconv.Itoa(count)+" numbers")
	} else if len(list) != count {
		warnings = append(warnings, fmt.Sprintf("inventar has %d numbers but countBooks is %d", len(list), count))
	}

	seen := map[string]bool{}
	for _, v := range list {
		if seen[v] {
			return nil, nil, errors.New("duplicate inventar number: " + v)
		}
		seen[v] = true
	}

	return list, warnings, nil
}
