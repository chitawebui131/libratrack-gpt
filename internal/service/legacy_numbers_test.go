package service

import (
	"testing"
)

func TestParseInventar(t *testing.T) {
	tests := []struct {
		input string
		count int
		want  []string
		warns int
		err   bool
	}{
		{"28 935", 1, []string{"28935"}, 0, false},
		{"22282-22298", 17, []string{
			"22282", "22283", "22284", "22285", "22286",
			"22287", "22288", "22289", "22290", "22291",
			"22292", "22293", "22294", "22295", "22296",
			"22297", "22298",
		}, 0, false},
		{"22282-22298", 5, []string{
			"22282", "22283", "22284", "22285", "22286",
			"22287", "22288", "22289", "22290", "22291",
			"22292", "22293", "22294", "22295", "22296",
			"22297", "22298",
		}, 1, false},
		{"100", 3, []string{"100-1", "100-2", "100-3"}, 1, false},
		{"10;12", 2, []string{"10", "12"}, 0, false},
		{"", 1, nil, 0, true},
		{"5-3", 1, nil, 0, true},
		{"7;7", 2, nil, 0, true},
	}

	for _, tt := range tests {
		list, warnings, err := parseInventar(tt.input, tt.count)
		if (err != nil) != tt.err {
			t.Errorf("parseInventar(%q, %d) error = %v, wantErr %v", tt.input, tt.count, err, tt.err)
			continue
		}
		if tt.err {
			continue
		}
		if len(warnings) != tt.warns {
			t.Errorf("parseInventar(%q, %d) warnings = %v, want %d", tt.input, tt.count, len(warnings), tt.warns)
		}
		if len(list) != len(tt.want) {
			t.Errorf("parseInventar(%q, %d) list length = %d, want %d", tt.input, tt.count, len(list), len(tt.want))
			continue
		}
		for i, v := range list {
			if v != tt.want[i] {
				t.Errorf("parseInventar(%q, %d) list[%d] = %q, want %q", tt.input, tt.count, i, v, tt.want[i])
			}
		}
	}
}

func TestParseCountYear(t *testing.T) {
	tests := []struct {
		input string
		want  int
		err   bool
	}{
		{"", 1, false},
		{"17", 17, false},
		{"0", 0, true},
		{"x", 0, true},
	}

	for _, tt := range tests {
		got, err := parseCount(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("parseCount(%q) error = %v, wantErr %v", tt.input, err, tt.err)
			continue
		}
		if !tt.err && got != tt.want {
			t.Errorf("parseCount(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}

	testsYear := []struct {
		input   string
		want    int
		warning string
	}{
		{"1995", 1995, ""},
		{"", 0, ""},
		{"95", 0, "year ignored: 95"},
		{"3000", 0, "year ignored: 3000"},
	}

	for _, tt := range testsYear {
		got, warning := parseYear(tt.input)
		if got != tt.want {
			t.Errorf("parseYear(%q) = %d, want %d", tt.input, got, tt.want)
		}
		if warning != tt.warning {
			t.Errorf("parseYear(%q) warning = %q, want %q", tt.input, warning, tt.warning)
		}
	}
}
