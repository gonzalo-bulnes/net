package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	_url "net/url"
	"strings"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

// Response represents the response from an HTTP request.
type Response struct {
	StatusCode int
	header     string
	// Header     map[string][]string
	ContentLength int64
	Body          io.ReadCloser
}

// Get issues a GET to the specified URL.
func Get(url string) (resp *Response, err error) {
	resp = &Response{
		ContentLength: -1, // indicates that the length is unknown
	}

	// preliminary checks
	u, err := _url.Parse(url)
	if err != nil {
		err = fmt.Errorf("Error parsin URL '%s': %v", url, err)
		return
	}
	if u.Scheme != "http" {
		err = fmt.Errorf("Unsupported protocol: %s", u.Scheme)
		return
	}
	address := u.Host
	if !strings.Contains(address, ":") {
		switch u.Scheme {
		case "http":
			address = fmt.Sprintf("%s%s", address, ":80")
		}
	}
	path := "/"
	if u.Path != "" {
		path = u.Path
	}

	// establish a TCP connection
	conn, err := net.Dial("tcp", address)
	if err != nil {
		err = fmt.Errorf("Connection error: %w", err)
		return
	}

	// write a request
	const requestTermination = "\r\n"
	fmt.Fprintf(conn, fmt.Sprintf("GET %s HTTP/1.0\r\n", path))
	// send a Host header for HTTP/1.1 servers that require it
	// e.g. servers that serve multiple domains
	fmt.Fprintf(conn, fmt.Sprintf("Host: %s\r\n%s", u.Hostname(), requestTermination))

	r := bufio.NewReader(conn)

	// read the first line of the response
	status, err := r.ReadString('\n')
	if err != nil {
		err = fmt.Errorf("Malformed response: %w", err)
		return
	}

	// interpret the first line of the response
	switch status {
	case "HTTP/1.0 200 OK\r\n":
		resp.StatusCode = 200
	case "HTTP/1.1 200 OK\r\n":
		resp.StatusCode = 200
	default:
		err = fmt.Errorf("Unknown status: %s", status)
		return
	}

	const mandatoryEmptyLine = "\r\n"
	for header := ""; header != mandatoryEmptyLine; {

		// read some more
		header, err = r.ReadString('\n')
		if err != nil {
			err = fmt.Errorf("Malformed response: %w", err)
			return
		}

		// just store the raw header for now
		switch header {
		default:
			resp.header = resp.header + header
		}
	}

	// read an eventual body
	body, err := ioutil.ReadAll(r)
	if err != nil {
		err = fmt.Errorf("Malformed response: %w", err)
		return
	}
	resp.ContentLength = int64(len(body))
	resp.Body = nopCloser{bytes.NewBuffer(body)}

	return
}
