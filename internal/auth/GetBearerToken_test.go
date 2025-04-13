package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		headers    http.Header
		expected   string
		shouldFail bool
	}{
		{
			name: "Valid Bearer Token",
			headers: http.Header{
				"Authorization": []string{"Bearer VALID_TOKEN"},
			},
			expected:   "VALID_TOKEN",
			shouldFail: false,
		},
		{
			name: "Missing Authorization Header",
			headers: http.Header{
				"Authorization": []string{},
			},
			expected:   "",
			shouldFail: true,
		},
		{
			name: "Malformed Authorization Header",
			headers: http.Header{
				"Authorization": []string{"INVALID_FORMAT"},
			},
			expected:   "",
			shouldFail: true,
		},
		{
			name: "Extra Whitespace Around Token",
			headers: http.Header{
				"Authorization": []string{"Bearer   WHITESPACE_TOKEN    "},
			},
			expected:   "WHITESPACE_TOKEN",
			shouldFail: false,
		},
		{
			name: "Case-Sensitive Check",
			headers: http.Header{
				"Authorization": []string{"BEARER ALL_CAPS"},
			},
			expected: "",
			shouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, err := GetBearerToken(test.headers)

			if test.shouldFail && err == nil {
				t.Fatalf("expected an error but got none for case: %v", test.name)
			}
			if !test.shouldFail && err != nil {
				t.Fatalf("did not expect an error but got one: %v for case: %v", err, test.name)
			}
			if !test.shouldFail && token != test.expected {
				t.Errorf("expected token %v, got %v for case: %v", test.expected, token, test.name)
			}
		})
	}
}
