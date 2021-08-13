package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-redis/redis"
)

type db_test struct{}

func (m *db_test) Get(key string) (string, error) {
	switch key {
	case "abc-1":
		return "yes", nil
	case "abc-2":
		return "", redis.Nil
	default:
		return "", errors.New("Internal Error")
	}
}
func (m *db_test) Search(key string) ([]string, error) {
	switch key {
	case "abc*":
		return []string{"abc-1", "abc-2"}, nil
	case "*-1":
		return []string{"abc-1", "xyz-1"}, nil
	case "xyz*":
		return []string{}, nil
	case "*-2":
		return []string{}, nil
	default:
		return []string{}, errors.New("Internal Error")
	}
}

func (m *db_test) Set(key string, value interface{}) error {
	switch key {
	case "xyz-4":
		return nil
	default:
		return errors.New("error in setting")
	}
}
func (m *db_test) TotalKeys() int {
	return 4
}

func newTestApplication(t *testing.T) *application {
	return &application{
		config: config{},
		logger: log.New(ioutil.Discard, "", 0),
		db:     &db_test{},
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewServer(h)
	return &testServer{ts}

}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)

	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
}
func (ts *testServer) post(t *testing.T, urlPath string, payload string) (int, http.Header, []byte) {
	rs, err := ts.Client().Post(ts.URL+urlPath, "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	return rs.StatusCode, rs.Header, body
}

func TestPing(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/healthcheck")

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}
}
func TestGet(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"key exists", "/get/abc-1", http.StatusOK, []byte("yes")},
		{"key doesn't exist", "/get/abc-2", http.StatusNotFound, nil},
		{"empty key", "/get", http.StatusNotFound, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}
func TestSet(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		payload  string
		wantCode int
	}{
		{"set key", "/set", `{"key":"xyz-4", "value":"f"}`, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := ts.post(t, tt.urlPath, tt.payload)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"prefix keys exists", "/search?prefix=abc", http.StatusOK, []byte(`{"keys":["abc-1","abc-2"]}`)},
		{"prefix keys doesn't exist", "/search?prefix=xyz", http.StatusOK, []byte(`{"keys":[]}`)},
		{"suffix keys exists", "/search?suffix=-1", http.StatusOK, []byte(`{"keys":["abc-1","xyz-1"]}`)},
		{"suffix keys doesn't exist", "/search?suffix=-2", http.StatusOK, []byte(`{"keys":[]}`)},
		{"empty filters", "/search", http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}
