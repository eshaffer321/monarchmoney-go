package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erickshaffer/monarchmoney-go/internal/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	loginEndpoint = "/auth/login/"
	mfaEndpoint   = "/auth/login/mfa/"
	userAgent     = "monarchmoney-go/1.0.0"
)

// Service handles authentication operations
type Service struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	session    *types.Session
	logger     types.Logger
}

// NewService creates a new auth service
func NewService(baseURL string, httpClient *http.Client, logger types.Logger) *Service {
	headers := map[string]string{
		"Accept":          "application/json",
		"Content-Type":    "application/json",
		"Client-Platform": "web",
		"User-Agent":      userAgent,
		"Origin":          "https://app.monarchmoney.com",
		"device-uuid":     uuid.New().String(),
	}

	return &Service{
		baseURL:    baseURL,
		httpClient: httpClient,
		headers:    headers,
		logger:     logger,
	}
}

// Login performs authentication
func (s *Service) Login(ctx context.Context, email, password string) error {
	return s.login(ctx, email, password, "")
}

// LoginWithMFA performs login with MFA code
func (s *Service) LoginWithMFA(ctx context.Context, email, password, mfaCode string) error {
	// First attempt login
	err := s.login(ctx, email, password, "")
	if err != nil && err.Error() != "MFA required" {
		return err
	}

	// Then submit MFA code
	return s.submitMFA(ctx, email, password, mfaCode)
}

// LoginWithTOTP performs login with TOTP secret
func (s *Service) LoginWithTOTP(ctx context.Context, email, password, totpSecret string) error {
	// Generate TOTP code
	code, err := generateTOTP(totpSecret)
	if err != nil {
		return errors.Wrap(err, "failed to generate TOTP code")
	}

	// First attempt login
	err = s.login(ctx, email, password, "")
	if err != nil && err.Error() != "MFA required" {
		return err
	}

	// Submit MFA with generated code
	return s.submitMFA(ctx, email, password, code)
}

// GetSession returns the current session
func (s *Service) GetSession() (*types.Session, error) {
	if s.session == nil {
		return nil, errors.New("not authenticated")
	}
	return s.session, nil
}

// SaveSession saves session to file
func (s *Service) SaveSession(path string) error {
	if s.session == nil {
		return errors.New("not authenticated")
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(err, "failed to create session directory")
	}

	// Marshal session to JSON (not pickle like Python)
	data, err := json.MarshalIndent(s.session, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal session")
	}

	// Write to file with restrictive permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return errors.Wrap(err, "failed to write session file")
	}

	if s.logger != nil {
		s.logger.Info("Session saved", "path", path)
	}

	return nil
}

// LoadSession loads session from file
func (s *Service) LoadSession(path string) error {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("not authenticated")
		}
		return errors.Wrap(err, "failed to read session file")
	}

	// Unmarshal session
	var session types.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return errors.Wrap(err, "failed to unmarshal session")
	}

	// Check expiry
	if !session.ExpiresAt.IsZero() && time.Now().After(session.ExpiresAt) {
		return errors.New("session expired")
	}

	s.session = &session

	if s.logger != nil {
		s.logger.Info("Session loaded", "path", path, "email", session.Email)
	}

	return nil
}

// SetSession sets the current session
func (s *Service) SetSession(session *types.Session) {
	s.session = session
}

// login performs the login request
func (s *Service) login(ctx context.Context, email, password, mfaCode string) error {
	// Create login request
	reqBody := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	if mfaCode != "" {
		reqBody["totp"] = mfaCode
	}

	// Marshal request
	body, err := json.Marshal(reqBody)
	if err != nil {
		return errors.Wrap(err, "failed to marshal login request")
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+loginEndpoint, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create login request")
	}

	// Set headers
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	// Log request
	if s.logger != nil {
		s.logger.Debug("Login request", "email", email)
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "login request failed")
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read login response")
	}

	// Log response
	if s.logger != nil {
		s.logger.Debug("Login response", "status", resp.StatusCode)
	}

	// Parse response
	var loginResp loginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return errors.Wrap(err, "failed to parse login response")
	}

	// Check for errors
	if loginResp.ErrorCode != "" {
		switch loginResp.ErrorCode {
		case "MFA_REQUIRED":
			return errors.New("MFA required")
		case "INVALID_CREDENTIALS":
			return errors.New("login failed")
		default:
			return &types.Error{
				Code:    loginResp.ErrorCode,
				Message: loginResp.Message,
			}
		}
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return errors.New("login failed")
		}
		return &types.Error{
			Code:       "LOGIN_FAILED",
			Message:    fmt.Sprintf("login failed with status %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
		}
	}

	// Extract token
	if loginResp.Token == "" {
		return errors.New("no token in login response")
	}

	// Create session
	s.session = &types.Session{
		Token:      loginResp.Token,
		UserID:     loginResp.UserID,
		Email:      email,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // Default 24h expiry
		DeviceUUID: s.headers["device-uuid"],
	}

	if s.logger != nil {
		s.logger.Info("Login successful", "email", email)
	}

	return nil
}

// submitMFA submits MFA code
func (s *Service) submitMFA(ctx context.Context, email, password, code string) error {
	// Create MFA request
	reqBody := map[string]interface{}{
		"email":    email,
		"password": password,
		"totp":     code,
	}

	// Marshal request
	body, err := json.Marshal(reqBody)
	if err != nil {
		return errors.Wrap(err, "failed to marshal MFA request")
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+loginEndpoint, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create MFA request")
	}

	// Set headers
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	// Log request
	if s.logger != nil {
		s.logger.Debug("MFA request", "email", email)
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "MFA request failed")
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read MFA response")
	}

	// Parse response
	var mfaResp loginResponse
	if err := json.Unmarshal(respBody, &mfaResp); err != nil {
		return errors.Wrap(err, "failed to parse MFA response")
	}

	// Check for errors
	if mfaResp.ErrorCode != "" {
		return &types.Error{
			Code:    mfaResp.ErrorCode,
			Message: mfaResp.Message,
		}
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return &types.Error{
			Code:       "MFA_FAILED",
			Message:    fmt.Sprintf("MFA failed with status %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
		}
	}

	// Extract token
	if mfaResp.Token == "" {
		return errors.New("no token in MFA response")
	}

	// Create session
	s.session = &types.Session{
		Token:      mfaResp.Token,
		UserID:     mfaResp.UserID,
		Email:      email,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		DeviceUUID: s.headers["device-uuid"],
	}

	if s.logger != nil {
		s.logger.Info("MFA successful", "email", email)
	}

	return nil
}

// generateTOTP generates a TOTP code from secret
func generateTOTP(secret string) (string, error) {
	// Remove spaces and convert to uppercase
	secret = strings.ReplaceAll(strings.ToUpper(secret), " ", "")

	// Decode base32 secret
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode TOTP secret")
	}

	// Get current time counter (30 second intervals)
	counter := time.Now().Unix() / 30

	// Convert counter to bytes
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	// Generate HMAC
	h := hmac.New(sha1.New, key)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff
	code = code % 1000000

	// Format as 6-digit string
	return fmt.Sprintf("%06d", code), nil
}

// loginResponse represents the login API response
type loginResponse struct {
	Token     string `json:"token"`
	UserID    string `json:"userId"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}
