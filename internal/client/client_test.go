package client_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/georlav/githunt/internal/client"
)

func TestClient_CheckGit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("result") {
		case "found":
			b, err := ioutil.ReadFile("testdata/response.txt")
			if err != nil {
				t.Fatal(err)
			}
			w.Write(b)
		case "notfound":
			http.NotFound(w, r)
		case "timeout":
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatal("Server got unexpected input")
		}
	}))
	t.Cleanup(func() {
		ts.Close()
	})

	// Initialize http client
	c := client.NewClient(
		client.SetTimeout(time.Second*30),
		client.SetQPS(10),
	)

	testsCases := []struct {
		description string
		url         string
		vulnerable  bool
		err         string
	}{
		{
			description: "Should find a vulnerable target",
			url:         ts.URL + "/.git/config?result=found",
			vulnerable:  true,
		},
		{
			description: "Should not find a vulnerable target",
			url:         ts.URL + "/.git/config?result=notfound",
			vulnerable:  false,
		},
		{
			description: "Should fail due to timeout",
			url:         ts.URL + "/.git/config?result=timeout",
			vulnerable:  false,
			err:         "context deadline exceeded",
		},
	}

	for i := range testsCases {
		tc := testsCases[i]

		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tc.url)
			if err != nil {
				t.Fatal(err)
			}

			result, err := c.CheckGit(context.Background(), u)
			if err != nil && !strings.Contains(err.Error(), tc.err) {
				t.Fatal(err)
			}

			if result != nil && tc.vulnerable != result.Vulnerable {
				t.Fatalf("Invalid result, expected %t got %t", tc.vulnerable, result.Vulnerable)
			}
		})
	}
}
