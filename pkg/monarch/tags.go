package monarch

import (
	"context"

	"github.com/pkg/errors"
)

// tagService implements the TagService interface
type tagService struct {
	client *Client
}

// List retrieves all tags
func (s *tagService) List(ctx context.Context) ([]*Tag, error) {
	query := s.client.loadQuery("tags/list.graphql")

	// No variables by default
	var variables map[string]interface{}
	
	var result struct {
		HouseholdTransactionTags []*Tag `json:"householdTransactionTags"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get tags")
	}

	return result.HouseholdTransactionTags, nil
}

// Create creates a new tag
func (s *tagService) Create(ctx context.Context, name, color string) (*Tag, error) {
	query := s.client.loadQuery("tags/create.graphql")

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"name":  name,
			"color": color,
		},
	}

	var result struct {
		CreateTransactionTag struct {
			Tag    *Tag `json:"tag"`
			Errors []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"createTransactionTag"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return nil, errors.Wrap(err, "failed to create tag")
	}

	if len(result.CreateTransactionTag.Errors) > 0 {
		return nil, &Error{
			Code:    result.CreateTransactionTag.Errors[0].Code,
			Message: result.CreateTransactionTag.Errors[0].Message,
		}
	}

	return result.CreateTransactionTag.Tag, nil
}

// SetTransactionTags sets tags on a transaction
func (s *tagService) SetTransactionTags(ctx context.Context, transactionID string, tagIDs ...string) error {
	query := s.client.loadQuery("tags/set.graphql")

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"transactionId": transactionID,
			"tagIds":        tagIDs,
		},
	}

	var result struct {
		SetTransactionTags struct {
			Transaction *struct {
				ID   string `json:"id"`
				Tags []*Tag `json:"tags"`
			} `json:"transaction"`
			Errors []struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"errors"`
		} `json:"setTransactionTags"`
	}

	if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
		return errors.Wrap(err, "failed to set transaction tags")
	}

	if len(result.SetTransactionTags.Errors) > 0 {
		return &Error{
			Code:    result.SetTransactionTags.Errors[0].Code,
			Message: result.SetTransactionTags.Errors[0].Message,
		}
	}

	return nil
}
