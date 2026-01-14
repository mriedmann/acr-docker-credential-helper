package acr

import (
	"fmt"
	"regexp"
	"strings"
)

// ACR registry URL patterns
const (
	// Standard ACR domain suffix
	ACRDomainSuffix = ".azurecr.io"

	// ACR registry name pattern (alphanumeric, 5-50 chars)
	ACRRegistryNamePattern = `^[a-zA-Z0-9]{5,50}$`
)

// RegistryValidator validates and extracts information from ACR registry URLs
type RegistryValidator struct {
	registryNameRegex *regexp.Regexp
}

// NewRegistryValidator creates a new registry validator
func NewRegistryValidator() *RegistryValidator {
	return &RegistryValidator{
		registryNameRegex: regexp.MustCompile(ACRRegistryNamePattern),
	}
}

// ValidateAndExtract validates that the server URL is an ACR registry
// and extracts the registry name
// Returns: registry name, error
func (v *RegistryValidator) ValidateAndExtract(serverURL string) (string, error) {
	// Remove common URL components
	serverURL = strings.TrimPrefix(serverURL, "https://")
	serverURL = strings.TrimPrefix(serverURL, "http://")
	serverURL = strings.TrimSuffix(serverURL, "/")

	// Check if it's an ACR domain
	if !strings.HasSuffix(serverURL, ACRDomainSuffix) {
		return "", fmt.Errorf(
			"not an ACR registry: URL must end with %s, got: %s",
			ACRDomainSuffix,
			serverURL,
		)
	}

	// Extract registry name
	registryName := strings.TrimSuffix(serverURL, ACRDomainSuffix)

	// Validate registry name format
	if !v.registryNameRegex.MatchString(registryName) {
		return "", fmt.Errorf(
			"invalid ACR registry name: must be 5-50 alphanumeric characters, got: %s",
			registryName,
		)
	}

	return registryName, nil
}

// IsACRRegistry checks if a URL is an ACR registry without validation
func (v *RegistryValidator) IsACRRegistry(serverURL string) bool {
	serverURL = strings.TrimPrefix(serverURL, "https://")
	serverURL = strings.TrimPrefix(serverURL, "http://")
	return strings.HasSuffix(serverURL, ACRDomainSuffix)
}
