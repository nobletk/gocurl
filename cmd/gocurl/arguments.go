package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/pflag"
)

type Flags struct {
	verbose   bool
	method    string
	header    []string
	data      string
	keepAlive bool
}

func (app *application) SplitArgsByNext(args []string) [][]string {
	var splitArgs [][]string
	var curArg []string

	for _, arg := range args {
		if arg == "--next" {
			splitArgs = append(splitArgs, curArg)
			curArg = nil
		} else {
			curArg = append(curArg, arg)
		}
	}
	splitArgs = append(splitArgs, curArg)

	return splitArgs
}

func (app *application) ParseArgs(arg []string) (Request, error) {
	var req Request

	f := pflag.NewFlagSet("request", pflag.ContinueOnError)

	f.BoolVarP(&req.flags.verbose, "verbose", "v", false, "Enable verbose mode")
	f.StringVarP(&req.flags.method, "method", "X", "GET", "Pass request method to server")
	f.StringSliceVarP(&req.flags.header, "header", "H", []string{}, "Pass custom header(s) to server")
	f.StringVarP(&req.flags.data, "data", "d", "", "HTTP POST data")
	f.BoolVarP(&req.flags.keepAlive, "keepAlive", "k", false, "Pass connection keep-alive to server")

	f.Usage = func() {
		var buf bytes.Buffer

		buf.WriteString("Usage:\n")
		buf.WriteString("  gocurl [-options] <url> ...\n")
		buf.WriteString("  gocurl [-options] <url> --next [-options] <url>...\n")
		buf.WriteString("Options:\n")

		fmt.Fprintf(os.Stderr, buf.String())
		f.PrintDefaults()
		fmt.Fprintf(os.Stderr, "      --next\t\t Make next URL use its separate set of options\n\n")
	}

	err := f.Parse(arg)
	if err != nil {
		if err != pflag.ErrHelp {
			f.Usage()
		}
		return req, err
	}

	if f.NArg() != 1 {
		f.Usage()
		return req, fmt.Errorf("incorrect argument format")
	}

	req.url = f.Arg(0)

	err = app.processURL(&req)
	if err != nil {
		return req, err
	}

	return req, nil
}

func (app *application) processURL(req *Request) error {
	parsedURL, err := url.ParseRequestURI(req.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	req.protocol = parsedURL.Scheme
	req.host = parsedURL.Hostname()
	req.path = parsedURL.Path
	if req.path == "" {
		req.path = "/"
	}
	req.port = parsedURL.Port()
	if req.port == "" {
		switch req.protocol {
		case "http":
			req.port = "80"
		case "https":
			req.port = "443"
		}
	}

	req.accept = "*/*"

	return nil
}
