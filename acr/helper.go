package acr

import (
	"os"

	"github.com/docker/docker-credential-helpers/credentials"
)

// ACRHelper implements the credentials.Helper interface for ACR
type ACRHelper struct {
	authenticator *AzureAuthenticator
	validator     *RegistryValidator
}

// NewACRHelper creates a new ACR credential helper
func NewACRHelper() *ACRHelper {
	return &ACRHelper{
		authenticator: NewAzureAuthenticator(),
		validator:     NewRegistryValidator(),
	}
}

// Get retrieves credentials for the specified server URL
// Returns: username (null GUID), password (refresh token), error
func (h *ACRHelper) Get(serverURL string) (string, string, error) {
	// 1. Validate server URL is an ACR registry
	registryName, err := h.validator.ValidateAndExtract(serverURL)
	if err != nil {
		return "", "", err
	}

	// 2. Get Azure access token
	azureToken, err := h.authenticator.GetAzureAccessToken()
	if err != nil {
		return "", "", WrapAzureAuthError(err)
	}

	// 3. Determine tenant ID: try extracting from JWT first, then fall back to env var
	tenantID, err := h.authenticator.ExtractTenantIDFromToken(azureToken)
	if err != nil || tenantID == "" {
		// Fall back to environment variable
		tenantID = os.Getenv("AZURE_TENANT_ID")
		if tenantID == "" {
			return "", "", NewMissingTenantIDError()
		}
	}

	// 4. Exchange for ACR refresh token
	refreshToken, err := h.authenticator.ExchangeForACRToken(
		registryName,
		serverURL,
		tenantID,
		azureToken,
	)
	if err != nil {
		return "", "", WrapACRTokenExchangeError(err)
	}

	// 5. Return Docker credentials
	// Username: null GUID (standard for ACR refresh tokens)
	// Password: ACR refresh token
	return "00000000-0000-0000-0000-000000000000", refreshToken, nil
}

// Add is not implemented (credential storage not required)
func (h *ACRHelper) Add(*credentials.Credentials) error {
	return NewNotImplementedError("Add")
}

// Delete is not implemented (credential removal not required)
func (h *ACRHelper) Delete(serverURL string) error {
	return NewNotImplementedError("Delete")
}

// List is not implemented (credential listing not required)
func (h *ACRHelper) List() (map[string]string, error) {
	return map[string]string{}, nil // Return empty map (no stored credentials)
}
