package acr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/docker/docker-credential-helpers/credentials"
)

// fakeAuthenticator implements Authenticator for testing
type fakeAuthenticator struct {
	accessToken     string
	accessTokenErr  error
	tenantID        string
	tenantIDErr     error
	refreshToken    string
	refreshTokenErr error
}

func (f *fakeAuthenticator) GetAzureAccessToken() (string, error) {
	return f.accessToken, f.accessTokenErr
}

func (f *fakeAuthenticator) ExtractTenantIDFromToken(_ string) (string, error) {
	return f.tenantID, f.tenantIDErr
}

func (f *fakeAuthenticator) ExchangeForACRToken(_, _, _, _ string) (string, error) {
	return f.refreshToken, f.refreshTokenErr
}

// successAuthenticator returns a fakeAuthenticator that succeeds with standard values
func successAuthenticator() *fakeAuthenticator {
	return &fakeAuthenticator{
		accessToken:  "fake-azure-token",
		tenantID:     "fake-tenant-id",
		refreshToken: "fake-refresh-token-12345",
	}
}

// runCommand executes a credential helper action and captures output
func runCommand(helper credentials.Helper, action, input string) (string, error) {
	in := strings.NewReader(input)
	out := new(bytes.Buffer)
	err := credentials.HandleCommand(helper, action, in, out)
	return out.String(), err
}

func TestGet_ValidACRRegistry(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	output, err := runCommand(helper, "get", "myregistry.azurecr.io")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var creds credentials.Credentials
	if err := json.Unmarshal([]byte(output), &creds); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput was: %s", err, output)
	}

	if creds.Username != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("expected null GUID username, got: %s", creds.Username)
	}
	if creds.Secret != "fake-refresh-token-12345" {
		t.Errorf("expected fake refresh token, got: %s", creds.Secret)
	}
	if creds.ServerURL != "myregistry.azurecr.io" {
		t.Errorf("expected server URL to be preserved, got: %s", creds.ServerURL)
	}
}

func TestGet_WithHTTPSPrefix(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	output, err := runCommand(helper, "get", "https://myregistry.azurecr.io")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var creds credentials.Credentials
	if err := json.Unmarshal([]byte(output), &creds); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if creds.Username != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("expected null GUID username, got: %s", creds.Username)
	}
	if creds.Secret != "fake-refresh-token-12345" {
		t.Errorf("expected fake refresh token, got: %s", creds.Secret)
	}
}

func TestGet_NonACRRegistry(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	_, err := runCommand(helper, "get", "registry-1.docker.io")
	if err == nil {
		t.Fatal("expected error for non-ACR registry, got nil")
	}
	if !strings.Contains(err.Error(), "not an ACR registry") {
		t.Errorf("expected 'not an ACR registry' error, got: %v", err)
	}
}

func TestGet_InvalidRegistryName(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	_, err := runCommand(helper, "get", "ab.azurecr.io")
	if err == nil {
		t.Fatal("expected error for invalid registry name, got nil")
	}
	if !strings.Contains(err.Error(), "invalid ACR registry name") {
		t.Errorf("expected 'invalid ACR registry name' error, got: %v", err)
	}
}

func TestGet_EmptyInput(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	_, err := runCommand(helper, "get", "")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestGet_AzureAuthFailure(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(&fakeAuthenticator{
		accessTokenErr: fmt.Errorf("no credential providers found"),
	})

	_, err := runCommand(helper, "get", "myregistry.azurecr.io")
	if err == nil {
		t.Fatal("expected error for auth failure, got nil")
	}
	if !strings.Contains(err.Error(), "Azure authentication failed") {
		t.Errorf("expected 'Azure authentication failed' error, got: %v", err)
	}
}

func TestGet_TokenExchangeFailure(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(&fakeAuthenticator{
		accessToken:     "fake-token",
		tenantID:        "fake-tenant",
		refreshTokenErr: fmt.Errorf("exchange endpoint returned 401"),
	})

	_, err := runCommand(helper, "get", "myregistry.azurecr.io")
	if err == nil {
		t.Fatal("expected error for token exchange failure, got nil")
	}
	if !strings.Contains(err.Error(), "ACR token exchange failed") {
		t.Errorf("expected 'ACR token exchange failed' error, got: %v", err)
	}
}

func TestGet_MissingTenantID(t *testing.T) {
	t.Setenv("AZURE_TENANT_ID", "")

	helper := NewACRHelperWithAuthenticator(&fakeAuthenticator{
		accessToken: "fake-token",
		tenantIDErr: fmt.Errorf("tid claim not found"),
	})

	_, err := runCommand(helper, "get", "myregistry.azurecr.io")
	if err == nil {
		t.Fatal("expected error for missing tenant ID, got nil")
	}
	if !strings.Contains(err.Error(), "Unable to determine tenant ID") {
		t.Errorf("expected 'Unable to determine tenant ID' error, got: %v", err)
	}
}

func TestStore_NotImplemented(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	input := `{"ServerURL":"myregistry.azurecr.io","Username":"user","Secret":"pass"}`
	_, err := runCommand(helper, "store", input)
	if err == nil {
		t.Fatal("expected error for store, got nil")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("expected 'not implemented' error, got: %v", err)
	}
}

func TestErase_NotImplemented(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	_, err := runCommand(helper, "erase", "myregistry.azurecr.io")
	if err == nil {
		t.Fatal("expected error for erase, got nil")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("expected 'not implemented' error, got: %v", err)
	}
}

func TestList_ReturnsEmptyMap(t *testing.T) {
	helper := NewACRHelperWithAuthenticator(successAuthenticator())

	output, err := runCommand(helper, "list", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput was: %s", err, output)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got: %v", result)
	}
}
