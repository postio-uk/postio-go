package postio

import (
	"context"
	"net/url"
)

// EmailService groups the /email/* endpoints.
type EmailService struct {
	client *Client
}

// Validate runs syntax / MX / SMTP checks against an email address.
func (s *EmailService) Validate(ctx context.Context, address string) (*EmailEnvelope, error) {
	var out EmailEnvelope
	path := "/email/" + url.PathEscape(address)
	if err := s.client.do(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
