# Azure Deployment Guide for Attestations Demo

This guide explains how to deploy the GitHub Attestations Demo Go application to Azure using Azure Container Apps and PostgreSQL Flexible Server.

## Prerequisites

- Azure CLI installed and logged in
- Docker installed (for building container images)
- An active Azure subscription
- Proper permissions to create resources in Azure

## Architecture

The deployment includes:
- **Azure Container Apps**: Hosts the Go application with auto-scaling
- **Azure Container Registry**: Stores the container images
- **Azure Database for PostgreSQL Flexible Server**: Managed PostgreSQL database
- **Log Analytics Workspace**: Application logging and monitoring

## Deployment Steps

### 1. Set Environment Variables

```bash
# Set your Azure subscription
export AZURE_SUBSCRIPTION_ID="your-subscription-id"
export RESOURCE_GROUP="attestations-demo-rg"
export LOCATION="eastus"
export DB_ADMIN_PASSWORD="YourSecurePassword123!"
```

### 2. Create Resource Group

```bash
az group create \
  --name $RESOURCE_GROUP \
  --location $LOCATION
```

### 3. Deploy Infrastructure

Deploy using Bicep template:

```bash
az deployment group create \
  --resource-group $RESOURCE_GROUP \
  --template-file azure/main.bicep \
  --parameters \
    location=$LOCATION \
    environmentName="attestations-demo" \
    dbAdminUsername="pgadmin" \
    dbAdminPassword=$DB_ADMIN_PASSWORD \
    databaseName="attestations"
```

### 4. Get Registry Credentials

```bash
# Get the registry name
REGISTRY_NAME=$(az deployment group show \
  --resource-group $RESOURCE_GROUP \
  --name main \
  --query properties.outputs.containerRegistryLoginServer.value \
  --output tsv | cut -d. -f1)

# Enable admin user
az acr update --name $REGISTRY_NAME --admin-enabled true

# Get credentials
REGISTRY_LOGIN_SERVER=$(az acr show --name $REGISTRY_NAME --query loginServer --output tsv)
REGISTRY_USERNAME=$(az acr credential show --name $REGISTRY_NAME --query username --output tsv)
REGISTRY_PASSWORD=$(az acr credential show --name $REGISTRY_NAME --query passwords[0].value --output tsv)
```

### 5. Build and Push Container Image

```bash
# Login to ACR
az acr login --name $REGISTRY_NAME

# Build and tag the image
docker build -f Dockerfile.go -t $REGISTRY_LOGIN_SERVER/attestations-demo:latest .

# Push the image
docker push $REGISTRY_LOGIN_SERVER/attestations-demo:latest
```

### 6. Update Container App

After pushing the image, restart the container app:

```bash
az containerapp update \
  --name attestations-demo-app \
  --resource-group $RESOURCE_GROUP
```

### 7. Get Application URL

```bash
APP_URL=$(az deployment group show \
  --resource-group $RESOURCE_GROUP \
  --name main \
  --query properties.outputs.containerAppFQDN.value \
  --output tsv)

echo "Application URL: https://$APP_URL"
```

### 8. Verify Deployment

Test the health endpoint:

```bash
curl https://$APP_URL/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-30T16:00:00Z",
  "version": "1.0.0",
  "database": "connected"
}
```

## Database Configuration

### Connection Details

The application automatically connects to PostgreSQL using these environment variables:
- `DB_HOST`: PostgreSQL server hostname
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_SSLMODE`: SSL mode (set to "require" for Azure)

### Manual Database Access

Connect to PostgreSQL using:

```bash
POSTGRES_HOST=$(az deployment group show \
  --resource-group $RESOURCE_GROUP \
  --name main \
  --query properties.outputs.postgresServerFQDN.value \
  --output tsv)

psql "host=$POSTGRES_HOST port=5432 dbname=attestations user=pgadmin sslmode=require"
```

### Database Schema

The application automatically creates the following schema on startup:

```sql
CREATE TABLE IF NOT EXISTS attestations (
    id SERIAL PRIMARY KEY,
    artifact_name VARCHAR(255) NOT NULL,
    artifact_digest VARCHAR(255) NOT NULL,
    provenance_uri TEXT,
    sbom_uri TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_artifact_digest ON attestations(artifact_digest);
CREATE INDEX IF NOT EXISTS idx_created_at ON attestations(created_at);
```

## Scaling Configuration

The Container App is configured with:
- **Min replicas**: 1
- **Max replicas**: 3
- **CPU**: 0.5 cores per instance
- **Memory**: 1 GB per instance
- **Scaling trigger**: 100 concurrent requests

Modify scaling in `azure/main.bicep` if needed.

## Monitoring

### View Logs

```bash
az containerapp logs show \
  --name attestations-demo-app \
  --resource-group $RESOURCE_GROUP \
  --follow
```

### View Metrics

Access metrics in Azure Portal:
1. Navigate to the Container App
2. Select "Metrics" from the left menu
3. View CPU, Memory, and HTTP request metrics

## Cost Optimization

For development/testing:
- Use Burstable tier for PostgreSQL (B1ms)
- Set min replicas to 0 for Container Apps (will scale to 0 when idle)
- Use Basic tier for Container Registry

For production:
- Consider General Purpose tier for PostgreSQL
- Enable geo-redundant backup
- Use Standard tier for Container Registry
- Configure alerts and monitoring

## Security Best Practices

1. **Use Azure Key Vault** for secrets:
   ```bash
   az keyvault create --name mykeyvault --resource-group $RESOURCE_GROUP
   az keyvault secret set --vault-name mykeyvault --name db-password --value $DB_ADMIN_PASSWORD
   ```

2. **Enable managed identity** for Container App to access Key Vault

3. **Configure network isolation**:
   - Use Private Endpoints for PostgreSQL
   - Restrict Container App ingress to specific IPs if needed

4. **Enable Azure Defender** for container security scanning

## Cleanup

To delete all resources:

```bash
az group delete --name $RESOURCE_GROUP --yes --no-wait
```

## Troubleshooting

### Container won't start
- Check logs: `az containerapp logs show`
- Verify environment variables are set correctly
- Ensure container image was pushed successfully

### Database connection fails
- Verify firewall rules allow Azure services
- Check SSL mode is set to "require"
- Verify credentials are correct

### Application not accessible
- Check ingress configuration
- Verify container is running: `az containerapp revision list`
- Check if health probes are passing

## Additional Resources

- [Azure Container Apps Documentation](https://learn.microsoft.com/azure/container-apps/)
- [Azure PostgreSQL Flexible Server](https://learn.microsoft.com/azure/postgresql/flexible-server/)
- [Azure Bicep Documentation](https://learn.microsoft.com/azure/azure-resource-manager/bicep/)
