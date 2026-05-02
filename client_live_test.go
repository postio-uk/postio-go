// Live tests against api.postio.co.uk (or stage). Skipped when no key
// is in the environment. Build tag keeps these out of `go test ./...`
// by default — opt in with `go test -tags=live`.

//go:build live

package postio_test

import (
	"context"
	"os"
	"testing"

	"github.com/postio-uk/postio-go"
)

func liveClient(t *testing.T) *postio.Client {
	t.Helper()
	key := os.Getenv("POSTIO_API_KEY_STAGE")
	baseURL := "https://stage-api.postio.co.uk/v1"
	if key == "" {
		key = os.Getenv("POSTIO_API_KEY_PROD")
		baseURL = "https://api.postio.co.uk/v1"
	}
	if key == "" {
		key = os.Getenv("POSTIO_API_KEY")
		baseURL = "https://api.postio.co.uk/v1"
	}
	if key == "" {
		t.Skip("no POSTIO_API_KEY* env var set")
	}
	c, err := postio.NewClient(postio.WithAPIKey(key), postio.WithBaseURL(baseURL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestLiveConnect(t *testing.T) {
	c := liveClient(t)
	r, err := c.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if !r.Success || r.Meta.RequestID == "" {
		t.Errorf("unexpected response: %+v", r)
	}
}

func TestLiveAddressSearch(t *testing.T) {
	c := liveClient(t)
	r, err := c.Address.Search(context.Background(), "downing street", &postio.SearchOptions{MaxResults: 5})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(r.Results) == 0 {
		t.Fatal("no results for downing street")
	}
	if r.Results[0].Suggestion == "" || r.Results[0].UDPRN == 0 {
		t.Errorf("unexpected first hit: %+v", r.Results[0])
	}
}

func TestLiveEmailValidate(t *testing.T) {
	c := liveClient(t)
	r, err := c.Email.Validate(context.Background(), "admin@postio.co.uk")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(r.Results) != 1 || !r.Results[0].IsValidSyntax {
		t.Errorf("unexpected response: %+v", r)
	}
}

func TestLivePhoneValidate(t *testing.T) {
	c := liveClient(t)
	r, err := c.Phone.Validate(context.Background(), "+442079460000")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(r.Results) != 1 || !r.Results[0].IsValid {
		t.Errorf("unexpected response: %+v", r)
	}
}
