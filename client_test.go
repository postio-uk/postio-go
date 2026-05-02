package postio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// helper: spin up a test server with a fixed handler.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(
		WithAPIKey("pk_test"),
		WithBaseURL(srv.URL),
		WithRetries(0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return srv, c
}

func TestSearchSendsHeaders(t *testing.T) {
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "pk_test" {
			t.Errorf("x-api-key = %q, want pk_test", got)
		}
		if got := r.Header.Get("accept"); got != "application/json" {
			t.Errorf("accept = %q, want application/json", got)
		}
		if got := r.Header.Get("user-agent"); got == "" || got[:11] != "postio-go/0" {
			t.Errorf("user-agent = %q, want postio-go/0...", got)
		}
		if got := r.URL.Query().Get("q"); got != "downing" {
			t.Errorf("q = %q, want downing", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"results": []map[string]any{
				{"udprn": 12345, "suggestion": "10 Downing Street"},
			},
			"meta": map[string]any{
				"countResults": 1,
				"requestId":    "test-req-id",
				"performance":  map[string]int{"workerMs": 5, "lookupMs": 2},
			},
		})
	})
	defer srv.Close()

	r, err := c.Address.Search(context.Background(), "downing", &SearchOptions{MaxResults: 5})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if !r.Success || len(r.Results) != 1 || r.Results[0].UDPRN != 12345 {
		t.Errorf("unexpected response: %+v", r)
	}
	if r.Meta.RequestID != "test-req-id" {
		t.Errorf("Meta.RequestID = %q", r.Meta.RequestID)
	}
}

func TestErrorMappings(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"401 → ErrInvalidKey", 401, `{"success":false,"error":"invalid_api_key","results":[],"meta":{"countResults":0,"requestId":"r-401","performance":{"workerMs":1,"lookupMs":0}}}`, ErrInvalidKey},
		{"402 → ErrOutOfCredit", 402, `{"success":false,"error":"out_of_credit","results":[],"meta":{"countResults":0,"requestId":"r-402","performance":{"workerMs":1,"lookupMs":0}}}`, ErrOutOfCredit},
		{"403 → ErrForbidden", 403, `{"success":false,"error":"forbidden","results":[],"meta":{"countResults":0,"requestId":"r-403","performance":{"workerMs":1,"lookupMs":0}}}`, ErrForbidden},
		{"404 → ErrNotFound", 404, `{"success":false,"error":"not_found","results":[],"meta":{"countResults":0,"requestId":"r-404","performance":{"workerMs":1,"lookupMs":0}}}`, ErrNotFound},
		{"400 → ErrValidation", 400, `{"success":false,"error":"bad_request","results":[],"meta":{"countResults":0,"requestId":"r-400","performance":{"workerMs":1,"lookupMs":0}}}`, ErrValidation},
		{"500 → ErrServer", 500, `{"success":false,"error":"internal","results":[],"meta":{"countResults":0,"requestId":"r-500","performance":{"workerMs":1,"lookupMs":0}}}`, ErrServer},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			defer srv.Close()
			_, err := c.Connect(context.Background())
			if !errors.Is(err, tc.sentinel) {
				t.Errorf("err=%v, want errors.Is(%s) → true", err, tc.sentinel)
			}
			var apiErr *Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *Error, got %T", err)
			}
			if apiErr.Status != tc.status {
				t.Errorf("Status = %d, want %d", apiErr.Status, tc.status)
			}
			if apiErr.RequestID == "" {
				t.Errorf("RequestID empty (expected populated)")
			}
		})
	}
}

func TestRateLimitRetryAfter(t *testing.T) {
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "12")
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"success":false,"error":"rate_limited","results":[],"meta":{"countResults":0,"requestId":"r-429","performance":{"workerMs":1,"lookupMs":0}}}`))
	})
	defer srv.Close()

	_, err := c.Connect(context.Background())
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.RetryAfter != 12 {
		t.Errorf("RetryAfter = %v, want 12", apiErr.RetryAfter)
	}
	if !errors.Is(err, ErrRateLimit) {
		t.Errorf("expected errors.Is(ErrRateLimit) true")
	}
}

func TestRetryOnServerError(t *testing.T) {
	calls := 0
	srv, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls < 2 {
			w.WriteHeader(503)
			_, _ = w.Write([]byte(`{"success":false,"error":"unavailable","results":[],"meta":{"countResults":0,"requestId":"r1","performance":{"workerMs":1,"lookupMs":0}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"meta":{"requestId":"r-ok","performance":{"workerMs":5,"lookupMs":2}}}`))
	})
	defer srv.Close()

	c, err := NewClient(
		WithAPIKey("pk_test"),
		WithBaseURL(srv.URL),
		WithRetries(2),
		WithRetryBackoff(0, 0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	r, err := c.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect (retried): %v", err)
	}
	if r.Meta.RequestID != "r-ok" {
		t.Errorf("RequestID = %q", r.Meta.RequestID)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (one retry)", calls)
	}
}

func TestNoAPIKey(t *testing.T) {
	t.Setenv("POSTIO_API_KEY", "")
	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when no api key")
	}
}

func TestContextCancellation(t *testing.T) {
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Connect(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
