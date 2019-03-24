package db

import (
	"testing"
)

func TestGetDigits(t *testing.T) {
	tests := [][2]string{
		[2]string{"01234567890", "01234567890"},
		[2]string{"87503118", "87503118"},
		[2]string{"(875) 031-18", "87503118"},
		[2]string{"+8 (750) 3118", "87503118"},
		[2]string{"87503118`~!@#$%^&*()-_=+[{]};:',<.>/?", "87503118"},
	}

	for index := range tests {
		input := tests[index][0]
		expected := tests[index][1]

		output := getDigits(input)

		if output != expected {
			t.Errorf("expected %s, got %s", expected, output)
		}
	}
}
