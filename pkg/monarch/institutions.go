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
		Credentials []struct {
			ID             string `json:"id"`
			UpdateRequired bool   `json:"updateRequired"`
			DataProvider   string `json:"dataProvider"`
			Institution    struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"institution"`
		} `json:"credentials"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get institutions")
	}

	// Transform credentials into institutions
	var institutions []*Institution
	for _, cred := range result.Credentials {
		inst := &Institution{
			ID:             cred.Institution.ID,
			Name:           cred.Institution.Name,
			URL:            cred.Institution.URL,
			CredentialID:   cred.ID,
			UpdateRequired: cred.UpdateRequired,
			DataProvider:   cred.DataProvider,
		}
		institutions = append(institutions, inst)
	}

	return institutions, nil
}
