package acr

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ACR registry URL patterns
const (
	// Standard ACR domain suffix
	ACRDomainSuffix = ".azurecr.io"

	// ACR registry name pattern (alphanumeric, 5-50 chars)
	ACRRegistryNamePattern = `^[a-z0-9]{5,50}$`
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

// ParseAndNormalize validates serverURL and returns normalized host + registry name.
// Normalized host format: <registry-name>.azurecr.io
func (v *RegistryValidator) ParseAndNormalize(serverURL string) (string, string, error) {
	raw := strings.TrimSpace(strings.ToLower(serverURL))
	if raw == "" {
		return "", "", fmt.Errorf("registry URL is empty")
	}

	host, err := normalizeRegistryHost(raw)
	if err != nil {
		return "", "", err
	}

	if !strings.HasSuffix(host, ACRDomainSuffix) {
		return "", "", fmt.Errorf(
			"not an ACR registry: URL must end with %s, got: %s",
			ACRDomainSuffix,
			host,
		)
	}

	registryName := strings.TrimSuffix(host, ACRDomainSuffix)
	if !v.registryNameRegex.MatchString(registryName) {
		return "", "", fmt.Errorf(
			"invalid ACR registry name: must be 5-50 alphanumeric characters, got: %s",
			registryName,
		)
	}

	return host, registryName, nil
}

func normalizeRegistryHost(raw string) (string, error) {
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("invalid registry URL: %w", err)
		}

		if parsed.Host == "" {
			return "", fmt.Errorf("invalid registry URL: missing host")
		}

		if parsed.Port() != "" {
			return "", fmt.Errorf("invalid registry URL: ports are not allowed")
		}

		if parsed.Path != "" && parsed.Path != "/" {
			return "", fmt.Errorf("invalid registry URL: paths are not allowed")
		}

		if parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
			return "", fmt.Errorf("invalid registry URL: query, fragment, and user info are not allowed")
		}

		return strings.TrimSuffix(parsed.Hostname(), "."), nil
	}

	trimmed := strings.TrimSuffix(raw, "/")
	if strings.ContainsAny(trimmed, "/?#") {
		return "", fmt.Errorf("invalid registry URL: paths, query, and fragment are not allowed")
	}

	if strings.Contains(trimmed, ":") {
		return "", fmt.Errorf("invalid registry URL: ports are not allowed")
	}

	if strings.Contains(trimmed, "@") {
		return "", fmt.Errorf("invalid registry URL: user info is not allowed")
	}

	return strings.TrimSuffix(trimmed, "."), nil
}

// IsACRRegistry checks if a URL is an ACR registry without full validation.
func (v *RegistryValidator) IsACRRegistry(serverURL string) bool {
	_, _, err := v.ParseAndNormalize(serverURL)
	return err == nil
}
