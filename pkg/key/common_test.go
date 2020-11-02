package key

import (
	"strconv"
	"testing"
)

func Test_MaxBatchSizeIsValid(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		workers int
		valid   bool
	}{
		{
			name:    "case 0: int - simple value",
			input:   "5",
			workers: 5,
			valid:   true,
		},
		{
			name:    "case 1: int - big value",
			input:   "200",
			workers: 300,
			valid:   true,
		},
		{
			name:    "case 2: int - invalid value - negative number",
			input:   "-10",
			workers: 5,
			valid:   false,
		},
		{
			name:    "case 2: int - invalid value - zero",
			input:   "0",
			workers: 5,
			valid:   false,
		},
		{
			name:    "case 3: int - invalid value - value bigger than worker count",
			input:   "20",
			workers: 5,
			valid:   false,
		},
		{
			name:    "case 4: percentage - simple value",
			input:   "0.5",
			workers: 10,
			valid:   true,
		},
		{
			name:    "case 5: percentage - rounding",
			input:   "0.35",
			workers: 10,
			valid:   true,
		},
		{
			name:    "case 6: percentage - rounding",
			input:   "0.32",
			workers: 10,
			valid:   true,
		},
		{
			name:    "case 7: percentage - invalid value - too big",
			input:   "1.5",
			workers: 10,
			valid:   false,
		},
		{
			name:    "case 8: percentage - invalid value - negative",
			input:   "-0.5",
			workers: 10,
			valid:   false,
		},
		{
			name:  "case 9: invalid value - '50%'",
			input: "50%",
			valid: false,
		},
		{
			name:  "case 10: invalid value - string",
			input: "test",
			valid: false,
		},
		{
			name:  "case 11: invalid value - number and string",
			input: "5erft",
			valid: false,
		},
		{
			name:  "case 12: invalid value - float and string",
			input: "0.5erft",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			isValid := MaxBatchSizeIsValid(tc.input, tc.workers)

			if isValid != tc.valid {
				t.Fatalf("%s - expected '%t' got '%t'\n", tc.name, tc.valid, isValid)
			}
		})
	}
}

func Test_PauseTimeIsValid(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "case 0: simple value",
			value: "PT15M",
			valid: true,
		},
		{
			name:  "case 2: simple value",
			value: "PT10S",
			valid: true,
		},
		{
			name:  "case 3: simple value",
			value: "PT1H2M10S",
			valid: true,
		},
		{
			name:  "case 4: simple value",
			value: "PT2M10S",
			valid: true,
		},
		{
			name:  "case 5: invalid value value",
			value: "10m",
			valid: false,
		},
		{
			name:  "case 6: invalid value value",
			value: "10s",
			valid: false,
		},
		{
			name:  "case 7: invalid value value",
			value: "10",
			valid: false,
		},
		{
			name:  "case 8: invalid value value",
			value: "1 hour",
			valid: false,
		},
		{
			name:  "case 9: invalid value value",
			value: "random string",
			valid: false,
		},
		{
			name:  "case 10: invalid value value",
			value: "",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := PauseTimeIsValid(tc.value)

			if result != tc.valid {
				t.Fatalf("%s -  expected '%t' got '%t'\n", tc.name, tc.valid, result)
			}
		})
	}
}
