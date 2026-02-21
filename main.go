package main

import (
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/mriedmann/acr-docker-credential-helper/acr"
)

func main() {
	// Create ACR helper instance
	helper := acr.NewACRHelper()

	// Serve the credential helper protocol
	// This reads from stdin, routes to appropriate method, writes to stdout
	credentials.Serve(helper)
}
