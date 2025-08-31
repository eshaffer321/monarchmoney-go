package monarch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/pkg/errors"
)

// adminService implements the AdminService interface
type adminService struct {
	client *Client
}

// GetSubscription retrieves subscription details
func (s *adminService) GetSubscription(ctx context.Context) (*Subscription, error) {
	query := s.client.loadQuery("admin/subscription.graphql")

	var result struct {
		Subscription *Subscription `json:"subscription"`
	}

	if err := s.client.executeGraphQL(ctx, query, nil, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get subscription details")
	}

	return result.Subscription, nil
}

// UploadBalanceHistory uploads CSV balance history
func (s *adminService) UploadBalanceHistory(ctx context.Context, accountID string, csvData []byte) error {
	// This uses a different endpoint, not GraphQL
	url := fmt.Sprintf("%s/account-balance-history/upload/", s.client.baseURL)

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add account ID field
	if err := writer.WriteField("account_id", accountID); err != nil {
		return errors.Wrap(err, "failed to write account_id field")
	}

	// Add CSV file
	part, err := writer.CreateFormFile("file", "balance_history.csv")
	if err != nil {
		return errors.Wrap(err, "failed to create form file")
	}

	if _, err := io.Copy(part, bytes.NewReader(csvData)); err != nil {
		return errors.Wrap(err, "failed to write CSV data")
	}

	if err := writer.Close(); err != nil {
		return errors.Wrap(err, "failed to close multipart writer")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return errors.Wrap(err, "failed to create upload request")
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add auth header if we have a session
	if s.client.session != nil && s.client.session.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", s.client.session.Token))
	}

	// Execute request
	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to upload balance history")
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return &Error{
			Code:       "UPLOAD_FAILED",
			Message:    fmt.Sprintf("upload failed with status %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return nil
}
