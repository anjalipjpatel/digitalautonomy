package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestValidateAmount(t *testing.T) {
	tcs := []struct {
		amount string
		valid  bool
	}{
		{amount: "12", valid: true},
		{amount: "12.50", valid: true},
		{amount: "0.5", valid: true},
		{amount: "0", valid: false},
		{amount: "12.500", valid: false},
		{amount: "1.", valid: false},
		{amount: "1e2", valid: false},
		{amount: "-1", valid: false},
	}

	for _, tc := range tcs {
		err := validateAmount(tc.amount)
		if (err == nil) != tc.valid {
			t.Errorf("validateAmount(%q) error = %v, valid = %t", tc.amount, err, tc.valid)
		}
	}
}

func TestSelectUserExcludesChosenSender(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("1\n"))
	selected, err := selectUser(reader, 0)
	if err != nil {
		t.Fatal(err)
	}
	if selected != 1 {
		t.Fatalf("selected index = %d, want Bob at index 1", selected)
	}
}
