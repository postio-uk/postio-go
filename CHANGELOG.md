# Changelog

All notable changes to `github.com/postio-uk/postio-go` are documented
here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
versioning follows [SemVer](https://semver.org/).

## [Unreleased]

## [0.1.2] — 2026-08-12

### Removed

- `county` from the address model. It came from ONS administrative geography
  rather than Royal Mail data, resolved for only part of the country, and is
  not part of a correct UK postal address — Royal Mail removed postal counties
  from PAF in December 2000. `district` and `ward` remain, and are ONS
  administrative geography rather than postal fields.

## [0.1.1] — 2026-05-02

### Changed

- `PhoneResult.IsReachable` typed `*bool` (was `interface{}`). postio-api
  1.0.3 aligned the spec with the runtime — HLR returns boolean.
- `PhoneResult` nullable fields drop `omitempty`. The runtime now always
  emits explicit nulls for every field, so the SDK round-trips without
  silently dropping fields.

## [0.1.0] — 2026-05-02

Initial release. First Postio Go SDK.

### Added

- `postio.Client` constructed via `NewClient(WithAPIKey, ...)`.
- Functional options pattern: `WithAPIKey`, `WithBaseURL`,
  `WithHTTPClient`, `WithTimeout`, `WithRetries`, `WithRetryBackoff`,
  `WithHeader`.
- Address endpoints: `client.Address.Search`, `client.Address.Postcode`,
  `client.Address.UDPRN`.
- Email endpoint: `client.Email.Validate`.
- Phone endpoint: `client.Phone.Validate`.
- Health probe: `client.Connect(ctx)`.
- Typed `*postio.Error` plus sentinel values (`ErrInvalidKey`,
  `ErrOutOfCredit`, `ErrForbidden`, `ErrNotFound`, `ErrValidation`,
  `ErrRateLimit`, `ErrServer`, `ErrTimeout`, `ErrConnection`) for
  `errors.Is` matching.
- Default retry policy: 2 retries with exp backoff + full jitter on
  408/409/429/5xx and network/timeout errors. Mirrors `@postio/node`.
- `POSTIO_API_KEY` env var fallback when `WithAPIKey` is not passed.
- Stdlib-only — no external dependencies.

### Notes

- `PhoneResult.IsReachable` is typed as `interface{}` rather than
  `*string` because the live API returns booleans there even though
  the spec says string-only. Patched once the API/spec align.

[Unreleased]: https://github.com/postio-uk/postio-go/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/postio-uk/postio-go/releases/tag/v0.1.1
[0.1.0]: https://github.com/postio-uk/postio-go/releases/tag/v0.1.0
