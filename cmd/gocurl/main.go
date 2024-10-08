package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type application struct {
	reqs      []Request
	connState ConnectionState
}

func main() {
	app := NewApplication()
	args := os.Args[1:]
	splitArgs := app.SplitArgsByNext(args)

	for _, arg := range splitArgs {
		req, err := app.ParseArgs(arg)
		if err != nil {
			if err == pflag.ErrHelp {
				fmt.Println(err)
				os.Exit(2)
			}
			fmt.Printf("Error parsing argument: %s\n", err)
			os.Exit(2)
		}
		app.reqs = append(app.reqs, req)
	}

	c, err := app.CreateHttpClient()
	if err != nil {
		fmt.Printf("Error creating http client: %s\n", err)
		os.Exit(1)
	}

	for _, req := range app.reqs {
		err := app.SendRequest(c, &req)
		if err != nil {
			fmt.Printf("Error sending request: %s\n", err)
			os.Exit(1)
		}
	}
}

func NewApplication() *application {
	return &application{
		reqs: []Request{},
		connState: ConnectionState{
			connNum:  -1,
			connPool: make(map[string]int),
			connAddr: "",
		},
	}
}
