package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendRequest(t *testing.T) {
	tests := []struct {
		name string
		reqs []Request
	}{
		{
			name: "HTTP GET Request",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					port:     "80",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						method: "GET",
					},
				},
			},
		},
		{
			name: "HTTP GET Request Verbose",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					port:     "80",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						method:  "GET",
						verbose: true,
					},
				},
			},
		},
		{
			name: "HTTP DELETE Request",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					port:     "80",
					path:     "/delete",
					accept:   "*/*",
					flags: Flags{
						method: "DELETE",
					},
				},
			},
		},
		{
			name: "HTTP POST Request",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					port:     "80",
					path:     "/post",
					accept:   "*/*",
					flags: Flags{
						method: "POST",
						data:   `{"key": "value"}`,
						header: []string{"Content-Type: application/json"},
					},
				},
			},
		},
		{
			name: "HTTP PUT Request",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					port:     "80",
					path:     "/put",
					accept:   "*/*",
					flags: Flags{
						method: "PUT",
						data:   `{"key": "value"}`,
						header: []string{"Content-Type: application/json"},
					},
				},
			},
		},
		{
			name: "HTTP PATCH Request",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					port:     "80",
					path:     "/patch",
					flags: Flags{
						method: "PATCH",
						header: []string{"accept: application/json"},
					},
				},
			},
		},
		{
			name: "HTTPS GET Request",
			reqs: []Request{
				{
					protocol: "https",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						method:  "GET",
						verbose: true,
					},
				},
			},
		},
		{
			name: "HTTPS HEAD Request",
			reqs: []Request{
				{
					protocol: "https",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						method: "HEAD",
					},
				},
			},
		},
		{
			name: "HTTPS HEAD Request Verbose",
			reqs: []Request{
				{
					protocol: "https",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						method:  "HEAD",
						verbose: true,
					},
				},
			},
		},
		{
			name: "HTTP Multiple Requests keepAlive",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						keepAlive: true,
						verbose:   true,
					},
				},
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						keepAlive: true,
						verbose:   true,
					},
				},
			},
		},
		{
			name: "HTTP Three Requests with 2 keepAlive",
			reqs: []Request{
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						keepAlive: true,
						verbose:   true,
					},
				},
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						verbose: true,
					},
				},
				{
					protocol: "http",
					host:     "eu.httpbin.org",
					path:     "/get",
					accept:   "*/*",
					flags: Flags{
						keepAlive: true,
						verbose:   true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			require.NoError(t, err)

			orgStdout := os.Stdout
			os.Stdout = w

			defer func() {
				os.Stdout = orgStdout
			}()

			app := NewApplication()
			cl, err := app.CreateHttpClient()
			require.NoError(t, err)

			for _, req := range tt.reqs {
				err = app.SendRequest(cl, &req)
				require.NoError(t, err)
			}

			w.Close()

			var stdoutBuf bytes.Buffer
			_, err = io.Copy(&stdoutBuf, r)
			require.NoError(t, err)

			t.Logf("\n%s", stdoutBuf.String())
		})
	}
}

func TestSendRequestFailNoHost(t *testing.T) {
	req := Request{
		protocol: "http",
		host:     "eu.httpbi.org",
		port:     "80",
		path:     "/get",
		accept:   "*/*",
		flags: Flags{
			method: "GET",
		},
	}

	app := NewApplication()
	cl, err := app.CreateHttpClient()
	require.NoError(t, err)

	err = app.SendRequest(cl, &req)
	expectedErr := `failed to send request: Get "http://eu.httpbi.org:80/get": dial tcp: lookup eu.httpbi.org on 192.168.1.1:53: no such host`
	require.EqualError(t, err, expectedErr)
}
