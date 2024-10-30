# GoCurl

A Command line tool to transfer data with URL syntax, for making HTTP/HTTPS requests. 


## Features

* HTTP and HTTPS protocols.

* GET, POST, PUT, DELETE, HEAD and PATCH methods.

* Pass custom header(s) to the server.

* Send data in the request body.

* Verbose mode for detailed info.

* Keep-Alive connection.

## Usage

```
gocurl [-options] <url> ...
```

or chain URLs with `--next`

```
gocurl [-options] <url> --next [-options] <url>...
```

The options are:

* `-d` or `--data` : HTTP POST data 

```
gocurl -d '{"key\": "value"}' <url> ...
```

* `-H` or `--header` : Pass custom header(s) to server 

```
gocurl -H "Content-Type: application/json" <url> ...
```

* `-k` or  `--keepAlive` : Pass connection keep-alive to server 

```
gocurl -k <url> ...
```

* `-X` or  `--method` : Pass request method to server (default "GET") 

```
gocurl -X <method> <url> ...
```

* `-v` or `--verbose` : Enable verbose mode 

```
gocurl -v <url> ...
```

## Getting started

### Clone the repo

```shell
git clone https://github.com/nobletk/gocurl
# then build the binary
make build
```

### Go
```shell
go install https://github.com/nobletk/gocurl/cmd/gocurl@latest
```
