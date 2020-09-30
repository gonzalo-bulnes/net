package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
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
	u := strings.SplitN(url, "://", 2)
	if len(u) != 2 {
		err = fmt.Errorf("Malformed URL: %s", url)
		return
	}
	if protocol := u[0]; protocol != "http" {
		err = fmt.Errorf("Unsupported protocol: %s", protocol)
		return
	}

	// establish a TCP connection
	conn, err := net.Dial("tcp", u[1])
	if err != nil {
		err = fmt.Errorf("Connection error: %w", err)
		return
	}

	// write a request
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")

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
