# docker-credential-acr

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mriedmann/acr-docker-credential-helper)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/mriedmann/acr-docker-credential-helper)](https://github.com/mriedmann/acr-docker-credential-helper/releases)
[![Container](https://img.shields.io/badge/container-ghcr.io-blue)](https://ghcr.io/mriedmann/acr-docker-credential-helper)

Docker credential helper for Azure Container Registry (ACR) using Azure SDK's DefaultAzureCredential.

This helper enables seamless authentication to Azure Container Registry by leveraging Azure's identity framework. It automatically exchanges Azure access tokens for ACR refresh tokens, allowing you to use Docker commands without manually managing credentials.

## Installation

### From GitHub Releases (Recommended)

Download the latest pre-built binary for Linux x86_64:

```bash
# Download the latest release
wget https://github.com/mriedmann/acr-docker-credential-helper/releases/latest/download/docker-credential-acr-linux-amd64

# Make executable
chmod +x docker-credential-acr-linux-amd64

# Install to PATH
sudo mv docker-credential-acr-linux-amd64 /usr/local/bin/docker-credential-acr
```

### From Container Image

Extract the binary from the container:

```bash
# Pull the container
docker pull ghcr.io/mriedmann/acr-docker-credential-helper:latest

# Extract binary
docker create --name temp ghcr.io/mriedmann/acr-docker-credential-helper:latest
docker cp temp:/docker-credential-acr /usr/local/bin/
docker rm temp
```

### From Source

```bash
go install github.com/mriedmann/acr-docker-credential-helper@latest
```

Ensure `$GOPATH/bin` or `$GOBIN` is in your PATH.

### Build Locally

```bash
git clone https://github.com/mriedmann/acr-docker-credential-helper.git
cd docker-credential-acr
CGO_ENABLED=0 go build -ldflags="-s -w" -o docker-credential-acr
sudo mv docker-credential-acr /usr/local/bin/
```

## Configuration

### 1. (Optional) Set Tenant ID Environment Variable

The helper can automatically extract the tenant ID from your Azure access token. However, if you prefer to set it explicitly or if your token doesn't include the `tid` claim, you can set:

```bash
export AZURE_TENANT_ID="your-azure-tenant-id"
```

To find your tenant ID:
```bash
az account show --query tenantId -o tsv
```

**Note**: In most cases, this environment variable is not needed. The helper will extract the tenant ID from the Azure access token JWT automatically.

### 2. Configure Azure Authentication

This helper uses Azure SDK's `DefaultAzureCredential`, which automatically tries multiple authentication methods in order:

1. **Environment variables** (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
2. **Workload Identity** (for AKS and other Kubernetes environments)
3. **Managed Identity** (for Azure VMs, App Service, Container Instances, etc.)
4. **Azure CLI** (`az login`)
5. **Azure PowerShell**
6. **Interactive browser** (if enabled)

#### For Local Development (Azure CLI)

The most common approach for local development:

```bash
az login --tenant $AZURE_TENANT_ID
```

#### For Azure-Hosted Applications

Use Managed Identity - no additional configuration needed, just ensure:
- The managed identity is enabled on your resource
- The identity has `AcrPull` or `AcrPush` role on the container registry

#### For Service Principals

Set environment variables:

```bash
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"
export AZURE_TENANT_ID="your-tenant-id"
```

### 3. Configure Docker

Edit `~/.docker/config.json` to use the credential helper.

#### Option A: Use for all registries (recommended if you only use ACR)

```json
{
  "credsStore": "acr"
}
```

#### Option B: Use for specific registries

```json
{
  "credHelpers": {
    "myregistry.azurecr.io": "acr",
    "anotherregistry.azurecr.io": "acr"
  }
}
```

## Usage

Once configured, Docker will automatically use this helper when accessing ACR registries:

```bash
# Pull images
docker pull myregistry.azurecr.io/myimage:latest

# Push images
docker tag myimage:latest myregistry.azurecr.io/myimage:latest
docker push myregistry.azurecr.io/myimage:latest

# Run containers
docker run myregistry.azurecr.io/myimage:latest
```

## How It Works

1. Docker detects you're accessing an ACR registry (e.g., `myregistry.azurecr.io`)
2. Docker calls `docker-credential-acr get` with the server URL
3. The helper validates the URL is an ACR registry (`*.azurecr.io`)
4. The helper authenticates to Azure using `DefaultAzureCredential`
5. The helper requests an Azure access token with scope `https://containerregistry.azure.net/.default`
6. The helper extracts the tenant ID:
   - First, parses the Azure access token (JWT) and extracts the `tid` claim
   - If not found, falls back to the `AZURE_TENANT_ID` environment variable
7. The helper exchanges the Azure token for an ACR refresh token via `POST /oauth2/exchange`
8. The helper returns credentials to Docker:
   - Username: `00000000-0000-0000-0000-000000000000` (null GUID)
   - Password: ACR refresh token
9. Docker uses these credentials to authenticate with the registry

## Troubleshooting

### "Unable to determine tenant ID: not found in access token and AZURE_TENANT_ID environment variable is not set"

This error occurs when:
1. Your Azure access token doesn't include the `tid` claim (rare), AND
2. The `AZURE_TENANT_ID` environment variable is not set

**Solution**: Set the environment variable explicitly:
```bash
export AZURE_TENANT_ID="your-tenant-id"
```

To find your tenant ID:
```bash
az account show --query tenantId -o tsv
```

To make it permanent, add it to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):
```bash
echo 'export AZURE_TENANT_ID="your-tenant-id"' >> ~/.bashrc
source ~/.bashrc
```

### "Azure authentication failed"

**Solution**: Ensure you're authenticated with Azure:

```bash
# Login with Azure CLI
az login --tenant $AZURE_TENANT_ID

# Verify authentication
az account show
```

For managed identities or service principals, verify:
- Environment variables are set correctly
- The identity has proper permissions on the ACR

### "not an ACR registry: URL must end with .azurecr.io"

**Solution**: This helper only works with Azure Container Registry (*.azurecr.io). For other registries, use different credential helpers or `docker login`.

### "ACR token exchange failed"

**Possible causes:**
1. **Incorrect AZURE_TENANT_ID**: Verify it matches your ACR's tenant
2. **Insufficient permissions**: Ensure your identity has `AcrPull` or `AcrPush` role
3. **Network issues**: Check connectivity to `*.azurecr.io`

**Solutions:**
```bash
# Verify tenant ID
az account show --query tenantId

# Check ACR permissions
az role assignment list --scope /subscriptions/YOUR_SUB_ID/resourceGroups/YOUR_RG/providers/Microsoft.ContainerRegistry/registries/YOUR_REGISTRY

# Test ACR connectivity
curl https://myregistry.azurecr.io/v2/
```

### Test the Helper Manually

You can test the credential helper directly:

```bash
# Set required environment variable
export AZURE_TENANT_ID="your-tenant-id"

# Test with your registry
echo '{"ServerURL":"myregistry.azurecr.io"}' | docker-credential-acr get
```

Expected output:
```json
{
  "Username": "00000000-0000-0000-0000-000000000000",
  "Secret": "eyJhbGc..."
}
```

## Required Azure Permissions

The Azure identity used by this helper must have one of these roles on the ACR:

- **AcrPull**: For pulling images (read-only)
- **AcrPush**: For pushing and pulling images
- **AcrDelete**: For deleting images (includes pull and push)
- **Owner** or **Contributor**: Full access

Assign roles via Azure CLI:
```bash
az role assignment create \
  --assignee YOUR_USER_OR_SP_ID \
  --role AcrPull \
  --scope /subscriptions/YOUR_SUB_ID/resourceGroups/YOUR_RG/providers/Microsoft.ContainerRegistry/registries/YOUR_REGISTRY
```

## Limitations

1. **Get operation only**: This helper only implements credential retrieval (`Get`). It does not store credentials (`Add`, `Delete`, `List` not implemented).

2. **ACR registries only**: Only works with `*.azurecr.io` registries. Custom DNS names or private endpoints are not supported.

3. **Tenant ID requirement**: The tenant ID must be available either in the Azure access token's `tid` claim (automatic) or via the `AZURE_TENANT_ID` environment variable (manual).

4. **No token caching**: Each Docker operation triggers a new token exchange. For high-frequency operations, this may add latency.

5. **No per-registry configuration**: All settings are global via environment variables.

## Security Considerations

- The helper never logs or stores tokens
- All communication with ACR uses HTTPS
- The helper is stateless (no credential persistence)
- Only requests the minimum required Azure scope (`https://containerregistry.azure.net/.default`)
- Follows Docker's credential helper security model

## Development

### Building from Source

```bash
# Build static binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o docker-credential-acr

# Verify it's statically linked
ldd docker-credential-acr  # Should output: "not a dynamic executable"
```

### Building Container

```bash
# Build container image
docker build -t docker-credential-acr:dev .

# Test the container
echo '{"ServerURL":"myregistry.azurecr.io"}' | docker run --rm -i docker-credential-acr:dev get
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with race detection
go test -v -race ./...

# Generate coverage report
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
gofmt -s -w .

# Run linters
golangci-lint run

# Security scanning
gosec ./...
govulncheck ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) - Azure authentication
- [Docker Credential Helpers](https://github.com/docker/docker-credential-helpers) - Credential helper protocol
- [Azure Container Registry Documentation](https://docs.microsoft.com/azure/container-registry/) - ACR OAuth2 flow
