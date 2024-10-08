package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitArgsByNext(t *testing.T) {
	args := []string{"-v", "http://eu.httpbin.org/get", "--next", "-v",
		"-X", "Connection: close", "http://eu.httpbin.org/get"}

	app := NewApplication()
	actual := app.SplitArgsByNext(args)

	expected := [][]string{
		{"-v", "http://eu.httpbin.org/get"},
		{"-v", "-X", "Connection: close", "http://eu.httpbin.org/get"},
	}
	assert.Equal(t, expected, actual)
}

func TestParseArgsWithNoErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Request
	}{
		{
			name: "Verbose HTTP Request",
			args: []string{"-v", "http://eu.httpbin.org/get"},
			expected: Request{
				url:      "http://eu.httpbin.org/get",
				protocol: "http",
				host:     "eu.httpbin.org",
				port:     "80",
				path:     "/get",
				accept:   "*/*",
				flags: Flags{
					verbose:   true,
					method:    "GET",
					header:    []string{},
					data:      "",
					keepAlive: false,
				},
			},
		},
		{
			name: "Verbose HTTP Request With Port",
			args: []string{"-v", "http://eu.httpbin.org:80/get"},
			expected: Request{
				url:      "http://eu.httpbin.org:80/get",
				protocol: "http",
				host:     "eu.httpbin.org",
				port:     "80",
				path:     "/get",
				accept:   "*/*",
				flags: Flags{
					verbose:   true,
					method:    "GET",
					header:    []string{},
					data:      "",
					keepAlive: false,
				},
			},
		},
		{
			name: "Verbose HTTPS Request",
			args: []string{"-v", "https://eu.httpbin.org/get"},
			expected: Request{
				url:      "https://eu.httpbin.org/get",
				protocol: "https",
				host:     "eu.httpbin.org",
				port:     "443",
				path:     "/get",
				accept:   "*/*",
				flags: Flags{
					verbose:   true,
					method:    "GET",
					header:    []string{},
					data:      "",
					keepAlive: false,
				},
			},
		},
		{
			name: "Verbose keepAlive",
			args: []string{"-v", "-k", "http://eu.httpbin.org/get"},
			expected: Request{
				url:      "http://eu.httpbin.org/get",
				protocol: "http",
				host:     "eu.httpbin.org",
				port:     "80",
				path:     "/get",
				accept:   "*/*",
				flags: Flags{
					verbose:   true,
					method:    "GET",
					header:    []string{},
					data:      "",
					keepAlive: true,
				},
			},
		},
		{
			name: "Verbose POST Custom Multiple Headers And Data",
			args: []string{
				"-v",
				"-X",
				"POST",
				"-H",
				"Content-Type: application/json",
				"-H",
				"Connection: keep-alive",
				"-d",
				`{"key\": "value"}`,
				"http://eu.httpbin.org/post",
			},
			expected: Request{
				url:      "http://eu.httpbin.org/post",
				protocol: "http",
				host:     "eu.httpbin.org",
				port:     "80",
				path:     "/post",
				accept:   "*/*",
				flags: Flags{
					verbose: true,
					method:  "POST",
					header: []string{
						"Content-Type: application/json",
						"Connection: keep-alive",
					},
					data:      `{"key\": "value"}`,
					keepAlive: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApplication()
			actual, err := app.ParseArgs(tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestParseArgsWithErrors(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "Unknown Shorthand Flag",
			args:          []string{"-r", "http://eu.httpbin.org/get"},
			expectedError: "unknown shorthand flag: 'r' in -r",
		},
		{
			name:          "Incorrect Argument Format",
			args:          []string{"-H", "http://eu.httpbin.org/get"},
			expectedError: "incorrect argument format",
		},
		{
			name:          "Incorrect Argument Format",
			args:          []string{"-v", "Connection: close", "http://eu.httpbin.org/get"},
			expectedError: "incorrect argument format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApplication()
			_, err := app.ParseArgs(tt.args)
			assert.EqualError(t, err, tt.expectedError)
		})
	}
}
