# postio-go — Claude Code working notes

Go SDK for `postio-api`. Mirrors `@postio/core` with idiomatic Go
ergonomics. Lives in its own repo because Go's module system expects
the import path to match the repo path (`github.com/postio-uk/postio-go`).

## Stack

- **Go 1.22+**. Uses generics-friendly stdlib only — no external
  dependencies, by design. `math/rand/v2` for full-jitter backoff.
- **HTTP**: stdlib `net/http`. Each `Client` holds one configurable
  `*http.Client`.
- **Tests**: stdlib `testing` + `httptest`. Live tests gated behind a
  `live` build tag.

## Why no codegen

JS/Python/PHP have mature OpenAPI codegen tools. Go has `oapi-codegen`,
but for 6 endpoints with 15 schemas the generated code is heavier than
hand-written models — and the JSON shape changes ride through
`json:"..."` tags without ceremony. Hand-written wins on readability.

If the API surface ever grows past ~15 endpoints, revisit and switch to
`oapi-codegen` for the model file.

## Layout

```
postio-go/
├── client.go              Client, options, transport, retry loop
├── address.go             AddressService (Search/Postcode/UDPRN)
├── email.go               EmailService.Validate
├── phone.go               PhoneService.Validate
├── models.go              every response struct + ErrorEnvelope
├── errors.go              Error, sentinel errors (ErrInvalidKey etc.)
├── version.go             Version constant
├── client_test.go         offline tests (httptest.Server)
├── client_live_test.go    `//go:build live` — live tests against stage
├── README.md
├── CLAUDE.md              this file
├── LICENSE                MIT
├── CHANGELOG.md
└── .github/workflows/
    └── ci.yml             vet + build + test matrix on 1.22 / 1.23
```

## Common commands

```bash
# build + offline tests
go vet ./...
go build ./...
go test -race -count=1 ./...

# live tests (requires POSTIO_API_KEY_STAGE in env or umbrella .env)
set -a && source ../.env && set +a
go test -count=1 -tags=live -run TestLive ./...

# pkg.go.dev preview
godoc -http=:6060   # if you have it installed
```

## Branch + deploy model

- `stage` — working branch.
- `master` — push triggers the `live-test` job. Cutting a release is
  just `git tag vX.Y.Z && git push origin vX.Y.Z`. Go's module proxy
  picks it up; pkg.go.dev indexes within a few minutes of first
  `go get`.
- No release workflow file. The Go module proxy is the registry.

## Versioning

Independent SemVer. Stay on `0.x` until the API contract is committed
stable. Bump to `1.0` only with a real public-API freeze of this SDK
(not the API spec).

## Spec drift

Two known gaps with the OpenAPI spec, papered over in `models.go`:

1. **PhoneResult nullable fields**: spec marks every `string|null`
   field as `required` with `omitempty` excluded; live API drops them
   entirely on invalid input. Pointer fields plus `omitempty` JSON
   tags handle the missing case.
2. **PhoneResult.IsReachable**: spec says `string|null`, live API
   returns `bool`. Typed as `interface{}` for now — reapply when
   regenerating against an aligned spec.

Both will go away once postio-api ships a spec/runtime alignment.

## Secrets the CI needs

| Secret | Used by | Notes |
|---|---|---|
| `POSTIO_API_KEY_STAGE` | `ci.yml` (live-test job) | Same value as the postio-python repo. Pair with `stage-api.postio.co.uk` (handled in `client_live_test.go`). |

No publish secret. Go modules are pull-based; tagging the repo is the
publish step.

## Tone for this repo

Same as the umbrella: terse, casual, status-emoji summaries. Customers
see error messages and godoc — keep those plain English.
