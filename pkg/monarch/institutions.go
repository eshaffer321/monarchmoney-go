package monarch

import (
	"context"

	"github.com/pkg/errors"
)

// institutionService implements the InstitutionService interface
type institutionService struct {
	client *Client
}

// List retrieves connected institutions
func (s *institutionService) List(ctx context.Context) ([]*Institution, error) {
	query := s.client.loadQuery("institutions/list.graphql")

	var result struct {
		Institutions []*Institution `json:"institutions"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get institutions")
	}

	return result.Institutions, nil
}
