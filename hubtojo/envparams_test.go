package main

import (
	"strings"
	"testing"
)

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      bool
		wantError bool
	}{
		{name: "true", value: "true", want: true},
		{name: "uppercase true", value: "TRUE", want: true},
		{name: "one", value: "1", want: true},
		{name: "false", value: "false"},
		{name: "zero", value: "0"},
		{name: "invalid", value: "yes", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_BOOL", tt.value)
			got, err := GetEnvBool("TEST_BOOL", false)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected boolean parsing error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parse boolean: %v", err)
			}
			if got != tt.want {
				t.Fatalf("value = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestGetEnvIntReturnsParsingError(t *testing.T) {
	t.Setenv("TEST_INT", "not-a-number")

	_, err := GetEnvInt("TEST_INT", 5)
	if err == nil || !strings.Contains(err.Error(), "must be an integer") {
		t.Fatalf("error = %v, want integer parsing error", err)
	}
}
