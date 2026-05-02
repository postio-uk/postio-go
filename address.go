package postio

import (
	"context"
	"net/url"
	"strconv"
)

// AddressService groups the /address/* endpoints.
type AddressService struct {
	client *Client
}

// SearchOptions tunes a free-text address search.
type SearchOptions struct {
	// MaxResults caps the number of hits returned. API default is 10
	// when zero. Maximum 100.
	MaxResults int
}

// PostcodeOptions tunes a postcode lookup.
type PostcodeOptions struct {
	MaxResults int
}

// Search performs a free-text typeahead lookup. Pass nil for default
// options.
func (s *AddressService) Search(ctx context.Context, q string, opts *SearchOptions) (*AddressSearchEnvelope, error) {
	values := url.Values{"q": {q}}
	if opts != nil && opts.MaxResults > 0 {
		values.Set("max_results", strconv.Itoa(opts.MaxResults))
	}
	var out AddressSearchEnvelope
	if err := s.client.do(ctx, "/address/search", values, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Postcode returns the full PAF address records for a UK postcode.
func (s *AddressService) Postcode(ctx context.Context, postcode string, opts *PostcodeOptions) (*AddressPostcodeEnvelope, error) {
	values := url.Values{}
	if opts != nil && opts.MaxResults > 0 {
		values.Set("max_results", strconv.Itoa(opts.MaxResults))
	}
	var out AddressPostcodeEnvelope
	path := "/address/postcode/" + url.PathEscape(postcode)
	if err := s.client.do(ctx, path, values, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UDPRN looks up a single PAF address by its UDPRN.
func (s *AddressService) UDPRN(ctx context.Context, udprn int) (*AddressUDPRNEnvelope, error) {
	var out AddressUDPRNEnvelope
	path := "/address/udprn/" + strconv.Itoa(udprn)
	if err := s.client.do(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
