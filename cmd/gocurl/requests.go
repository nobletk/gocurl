package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"os"
	"sync"

	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

type Request struct {
	url      string
	protocol string
	host     string
	port     string
	path     string
	accept   string
	flags    Flags
}

type ConnectionState struct {
	connNum  int
	connPool map[string]int
	connAddr string
}

const (
	BOLD  = "\033[1m"
	RESET = "\033[0m"
)

func (app *application) SendRequest(cl *http.Client, req *Request) error {
	url := fmt.Sprintf("%s://%s:%s%s", req.protocol, req.host, req.port, req.path)

	var reqBody bytes.Buffer
	if req.flags.data != "" {
		reqBody.WriteString(req.flags.data)
	}

	httpReq, err := http.NewRequest(req.flags.method, url, &reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	app.setRequestHeaders(httpReq, req)

	if req.flags.verbose {
		httpReq = httpReq.WithContext(httptrace.WithClientTrace(context.Background(), app.createClientTrace(req)))
	}

	resp, err := cl.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	err = app.handleVerbose(req, resp)
	if err != nil {
		return fmt.Errorf("failed to handle verbose: %w", err)
	}

	if req.flags.method == "HEAD" && !req.flags.verbose {
		app.printHeader(resp)
	}

	if req.flags.method != "HEAD" {
		fmt.Printf("<\n")
		io.Copy(os.Stdout, resp.Body)
	}

	if req.flags.keepAlive {
		fmt.Printf("* Connection #%d to host %s left intact\n", app.connState.connNum, req.host)
	}

	return nil
}

func (app *application) handleVerbose(req *Request, resp *http.Response) error {
	if !req.flags.verbose {
		return nil
	}

	rawResponse, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return err
	}

	lines := bytes.Split(rawResponse, []byte("\r\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		fmt.Printf("< %s\n", line)

		if req.flags.method == "HEAD" {
			app.printMatchingHeaders(resp, line)
		}
	}

	return nil
}

func (app *application) printMatchingHeaders(resp *http.Response, line []byte) {
	for key, values := range resp.Header {
		if bytes.Contains(line, []byte(key)) {
			fmt.Printf("%s%s%s: %s\n", BOLD, key, RESET, values[0])
		}
	}
}

func (app *application) printHeader(resp *http.Response) {
	var headerBuf bytes.Buffer
	headerOrder := []string{
		"Date",
		"Content-Type",
		"Content-Length",
		"Server",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
	}

	headerBuf.WriteString(fmt.Sprintf("HTTP/%d %d\n", resp.ProtoMajor, resp.StatusCode))
	for _, key := range headerOrder {
		if value := resp.Header.Get(key); value != "" {
			headerBuf.WriteString(fmt.Sprintf("%s%s%s: %s\n", BOLD, key, RESET, value))
		}
	}
	fmt.Print(headerBuf.String())
}

func (app *application) CreateHttpClient() (*http.Client, error) {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	if err := http2.ConfigureTransport(tr); err != nil {
		return nil, fmt.Errorf("failed to configure HTTP/2: %w", err)
	}

	cl := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 15,
	}

	return cl, nil
}

func (app *application) setRequestHeaders(httpReq *http.Request, req *Request) {
	httpReq.Header.Set("Accept", req.accept)
	httpReq.Header.Set("User-Agent", "gocurl/1.0")

	if req.flags.keepAlive {
		httpReq.Header.Set("Connection", "keep-alive")
	} else {
		httpReq.Header.Set("Connection", "close")
	}

	if req.flags.header != nil {
		for _, header := range req.flags.header {
			s := strings.Split(header, ":")
			key := strings.TrimSpace(s[0])
			value := strings.TrimSpace(s[1])
			httpReq.Header.Set(key, value)
			if key == "Connection" && value == "keep-alive" {
				req.flags.keepAlive = true
			}
		}
	}

	if (req.flags.method == "POST" || req.flags.method == "PATCH" || req.flags.method == "PUT") && req.flags.data != "" {
		httpReq.Header.Set("Content-Length", strconv.Itoa(len(req.flags.data)))
	}
}

func (app *application) createClientTrace(req *Request) *httptrace.ClientTrace {
	httpVersion := "HTTP/1.1"
	if req.protocol == "https" {
		httpVersion = "HTTP/2"
	}

	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			fmt.Printf("* Trying %s...\n", hostPort)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			curConnAddr := info.Conn.RemoteAddr().String()
			curConnNum := app.connState.connPool[curConnAddr]
			if info.Reused {
				fmt.Printf("* Reusing existing connection #%d with host %s\n", curConnNum, curConnAddr)
			}
		},
		PutIdleConn: func(err error) {
			fmt.Printf("* Connection #%d returned to idle pool\n", app.connState.connNum)
		},
		DNSStart: func(di httptrace.DNSStartInfo) {
			fmt.Printf("* DNS lookup begins...\n")
		},
		DNSDone: func(di httptrace.DNSDoneInfo) {
			fmt.Printf("* DNS lookup table:\n")
			if !di.Coalesced {
				for _, a := range di.Addrs {
					fmt.Printf("  - (%s)\n", a.IP)
				}
			}
		},
		ConnectStart: func(network, addr string) {
			fmt.Println("* Starting connection...")
		},
		ConnectDone: func(network, addr string, err error) {
			var mu sync.Mutex
			mu.Lock()
			defer mu.Unlock()
			if _, exists := app.connState.connPool[addr]; !exists {
				app.connState.connNum++
				app.connState.connPool[addr] = app.connState.connNum
				app.connState.connAddr = addr
			}

			app.connState.connNum = app.connState.connPool[addr]
			app.connState.connAddr = addr

			if err != nil {
				fmt.Printf("* Failed to connect to %s (%s) #%d: %v\n", req.host, addr, app.connState.connNum, err)
			} else {
				fmt.Printf("* Connected to %s (%s) #%d\n", req.host, addr, app.connState.connNum)
			}
		},
		TLSHandshakeStart: func() {
			fmt.Printf("* Starting TLS handshake\n")
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if err != nil {
				fmt.Printf("* TLS handshake failed: %v\n", err)
			} else {
				cert := cs.PeerCertificates[0]
				fmt.Printf("* SSL connection using %s / %s\n", tls.VersionName(cs.Version), tls.CipherSuiteName(cs.CipherSuite))
				fmt.Printf("* Server certificate:\n")
				fmt.Printf("  - Subject: %s\n", cert.Subject)
				fmt.Printf("  - Issuer: %s\n", cert.Issuer)
				fmt.Printf("  - Valid from: %s\n", cert.NotBefore)
				fmt.Printf("  - Valid to: %s\n", cert.NotAfter)
				if cs.HandshakeComplete {
					fmt.Printf("* SSL verification succeded.\n")
					fmt.Println("* using HTTP/2")
				} else {
					fmt.Printf("* SSL verification failed.\n")
				}
			}
		},
		WroteHeaders: func() {
			fmt.Printf("> %s %s %s\n", req.flags.method, req.path, httpVersion)
			fmt.Printf("> Host: %s\n", req.host)
			fmt.Printf("> Accept: %s\n", req.accept)
			if req.flags.header != nil {
				for _, header := range req.flags.header {
					s := strings.Split(header, ":")
					fmt.Printf("> %s: %s\n", strings.TrimSpace(s[0]), strings.TrimSpace(s[1]))
				}
			}
			fmt.Printf(">\n")
		},
	}
}
