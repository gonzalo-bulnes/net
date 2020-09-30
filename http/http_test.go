package http

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	srv := http.FileServer(http.Dir("testdata"))
	http.Handle("/", http.StripPrefix("testdata", srv))
	go http.ListenAndServe(":8080", srv)

	resp, err := Get("http://localhost:8080")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatalf("Expected response, got none")
	}
	if actual, expected := resp.StatusCode, http.StatusOK; actual != expected {
		t.Errorf("Expected HTTP %d, got HTTP %d", expected, actual)
	}

	if expected := int64(398); resp.ContentLength != expected {
		t.Errorf("Expected content length to be %d bytes, got %d", expected, resp.ContentLength)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if actual := int64(len(body)); actual != resp.ContentLength {
		t.Errorf("Expected body length to be %d bytes), got %d", resp.ContentLength, actual)
	}

	if known := "<p>swoop!</p>"; !strings.Contains(string(body), known) {
		t.Errorf("Expected body to contain known string '%s', got\n%s\n", known, body)
	}
}
