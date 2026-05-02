# Postio Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/postio-uk/postio-go.svg)](https://pkg.go.dev/github.com/postio-uk/postio-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/postio-uk/postio-go)](https://goreportcard.com/report/github.com/postio-uk/postio-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go SDK for the [Postio API](https://postio.co.uk) — UK address, email, and
phone validation. Backed by Royal Mail PAF and Ordnance Survey.
Stdlib `net/http` only, zero dependencies.

## Install

```bash
go get github.com/postio-uk/postio-go
```

Requires Go 1.22+.

## 30-second example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/postio-uk/postio-go"
)

func main() {
    client, err := postio.NewClient(postio.WithAPIKey("pk_live_..."))
    if err != nil {
        log.Fatal(err)
    }

    result, err := client.Address.Search(context.Background(), "downing street", nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, hit := range result.Results {
        fmt.Println(hit.UDPRN, hit.Suggestion)
    }
    fmt.Println("request id:", result.Meta.RequestID)
}
```

API key may also come from `POSTIO_API_KEY`.

## API

| Method | Returns |
|---|---|
| `client.Address.Search(ctx, q, opts)` | `*AddressSearchEnvelope` |
| `client.Address.Postcode(ctx, postcode, opts)` | `*AddressPostcodeEnvelope` |
| `client.Address.UDPRN(ctx, udprn)` | `*AddressUDPRNEnvelope` |
| `client.Email.Validate(ctx, address)` | `*EmailEnvelope` |
| `client.Phone.Validate(ctx, number)` | `*PhoneEnvelope` |
| `client.Connect(ctx)` | `*ConnectSuccess` |

`opts` may be `nil` for default behaviour. Every method takes a
`context.Context` for cancellation/timeout — pass `context.Background()`
if you don't have one.

## Errors

Every non-2xx response returns a `*postio.Error`. Match by sentinel
with `errors.Is`, or by struct with `errors.As` for full detail.

```go
result, err := client.Address.Postcode(ctx, "not-a-postcode", nil)
if errors.Is(err, postio.ErrValidation) {
    var e *postio.Error
    errors.As(err, &e)
    fmt.Printf("validation failed (request_id=%s): %s\n", e.RequestID, e.Details)
}
```

| Sentinel | HTTP |
|---|---|
| `postio.ErrValidation` | 400 / 422 |
| `postio.ErrInvalidKey` | 401 |
| `postio.ErrOutOfCredit` | 402 |
| `postio.ErrForbidden` | 403 |
| `postio.ErrNotFound` | 404 |
| `postio.ErrRateLimit` | 429 (`.RetryAfter` populated when sent) |
| `postio.ErrServer` | 5xx |
| `postio.ErrTimeout` | local request timeout |
| `postio.ErrConnection` | transport error |

The `*Error` struct exposes `Status`, `Code`, `Message`, `Details`,
`RequestID`, `RetryAfter`, the raw `Envelope`, and a `Cause` wrapping
the underlying transport error.

## Configuration

```go
client, err := postio.NewClient(
    postio.WithAPIKey("pk_live_..."),
    postio.WithBaseURL("https://api.postio.co.uk/v1"),  // default
    postio.WithTimeout(10 * time.Second),                // default
    postio.WithRetries(2),                               // default; 0 to disable
    postio.WithRetryBackoff(500*time.Millisecond, 8*time.Second),
    postio.WithHeader("x-tracking-id", "..."),
)
```

Default retry policy: 2 retries on 408/409/429/5xx + network/timeout,
exponential backoff with full jitter (500ms → 8s cap).

## Links

- [pkg.go.dev/github.com/postio-uk/postio-go](https://pkg.go.dev/github.com/postio-uk/postio-go)
- [Docs](https://postio.co.uk/docs)
- [API reference (OpenAPI)](https://postio.co.uk/openapi.json)
- [Issues](https://github.com/postio-uk/postio-go/issues)

## License

MIT — see [LICENSE](./LICENSE).

> *Postio is a trading name of Onno Group Limited, registered in
> England & Wales (company no. 08622799). Registered office:
> Suite 22 Trym Lodge, 1 Henbury Road, Westbury-On-Trym, Bristol BS9 3HQ.*
