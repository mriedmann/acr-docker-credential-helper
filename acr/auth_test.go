package acr

import "testing"

func TestExchangeForACRTokenRejectsNonACRHosts(t *testing.T) {
	a := NewAzureAuthenticator()
	for _, host := range []string{
		"evil.example.com",
		"myregistry.azurecr.io.evil.com",
		"myregistry.azurecr.io:8443",
		"user@myregistry.azurecr.io",
		"127.0.0.1",
	} {
		if _, err := a.ExchangeForACRToken(host, "tenant", "token"); err == nil {
			t.Errorf("expected error for host %q, got nil", host)
		}
	}
}
