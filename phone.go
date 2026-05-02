package postio

import (
	"context"
	"net/url"
)

// PhoneService groups the /phone/* endpoints.
type PhoneService struct {
	client *Client
}

// Validate runs format / carrier / reachability checks against a phone
// number. Accepts E.164 (preferred) or other formats; the API normalises.
func (s *PhoneService) Validate(ctx context.Context, number string) (*PhoneEnvelope, error) {
	var out PhoneEnvelope
	path := "/phone/" + url.PathEscape(number)
	if err := s.client.do(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
