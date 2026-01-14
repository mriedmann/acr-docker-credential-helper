package acr

import (
	"fmt"
)

// Error types for different failure scenarios

// MissingTenantIDError indicates AZURE_TENANT_ID is not set
type MissingTenantIDError struct{}

func (e *MissingTenantIDError) Error() string {
	return "Unable to determine tenant ID: not found in access token and " +
		"AZURE_TENANT_ID environment variable is not set. " +
		"Please set AZURE_TENANT_ID to your Azure tenant ID."
}

func NewMissingTenantIDError() error {
	return &MissingTenantIDError{}
}

// NotImplementedError indicates an operation is not supported
type NotImplementedError struct {
	Operation string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf(
		"operation '%s' is not implemented by docker-credential-acr. "+
			"This helper only supports credential retrieval (Get).",
		e.Operation,
	)
}

func NewNotImplementedError(operation string) error {
	return &NotImplementedError{Operation: operation}
}

// AzureAuthError wraps Azure authentication failures
type AzureAuthError struct {
	Cause error
}

func (e *AzureAuthError) Error() string {
	return fmt.Sprintf(
		"Azure authentication failed: %v. "+
			"Ensure you are logged in via Azure CLI, have a managed identity, "+
			"or have set appropriate environment variables (AZURE_CLIENT_ID, "+
			"AZURE_CLIENT_SECRET, AZURE_TENANT_ID).",
		e.Cause,
	)
}

func WrapAzureAuthError(err error) error {
	return &AzureAuthError{Cause: err}
}

// ACRTokenExchangeError wraps ACR token exchange failures
type ACRTokenExchangeError struct {
	Cause error
}

func (e *ACRTokenExchangeError) Error() string {
	return fmt.Sprintf(
		"ACR token exchange failed: %v. "+
			"Verify that AZURE_TENANT_ID is correct and that you have "+
			"permission to access the registry.",
		e.Cause,
	)
}

func WrapACRTokenExchangeError(err error) error {
	return &ACRTokenExchangeError{Cause: err}
}
