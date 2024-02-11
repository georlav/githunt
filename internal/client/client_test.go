package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/georlav/githunt/internal/client"
)

func TestClient_CheckGit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("result") {
		case "found":
			b, err := os.ReadFile("testdata/response.txt")
			if err != nil {
				t.Fatal(err)
			}
			_, _ = w.Write(b)
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
		client.SetTimeout(time.Second * 5),
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

			isVulnerable, err := c.CheckGit(context.Background(), u)
			if err != nil && !strings.Contains(err.Error(), tc.err) {
				t.Fatal(err)
			}

			if tc.vulnerable != isVulnerable {
				t.Fatal("Unexpected result")
			}
		})
	}
}
