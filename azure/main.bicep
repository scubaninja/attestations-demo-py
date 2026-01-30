// Azure Bicep template for deploying GitHub Attestations Demo to Azure
// This template creates:
// - Azure Container Apps Environment
// - Azure Container Registry
// - Azure Database for PostgreSQL Flexible Server
// - Azure Container App with the Go application

@description('Location for all resources')
param location string = resourceGroup().location

@description('Environment name for the application')
param environmentName string = 'attestations-demo'

@description('Container Registry name')
param registryName string = 'acr${uniqueString(resourceGroup().id)}'

@description('Database administrator username')
@secure()
param dbAdminUsername string

@description('Database administrator password')
@secure()
param dbAdminPassword string

@description('Database name')
param databaseName string = 'attestations'

@description('Container image tag')
param imageTag string = 'latest'

// Azure Container Registry
resource containerRegistry 'Microsoft.ContainerRegistry/registries@2023-01-01-preview' = {
  name: registryName
  location: location
  sku: {
    name: 'Basic'
  }
  properties: {
    adminUserEnabled: true
  }
}

// Log Analytics Workspace for Container Apps
resource logAnalytics 'Microsoft.OperationalInsights/workspaces@2022-10-01' = {
  name: '${environmentName}-logs'
  location: location
  properties: {
    sku: {
      name: 'PerGB2018'
    }
    retentionInDays: 30
  }
}

// Container Apps Environment
resource containerAppEnv 'Microsoft.App/managedEnvironments@2023-05-01' = {
  name: '${environmentName}-env'
  location: location
  properties: {
    appLogsConfiguration: {
      destination: 'log-analytics'
      logAnalyticsConfiguration: {
        customerId: logAnalytics.properties.customerId
        sharedKey: logAnalytics.listKeys().primarySharedKey
      }
    }
  }
}

// PostgreSQL Flexible Server
resource postgresServer 'Microsoft.DBforPostgreSQL/flexibleServers@2022-12-01' = {
  name: '${environmentName}-postgres'
  location: location
  sku: {
    name: 'Standard_B1ms'
    tier: 'Burstable'
  }
  properties: {
    administratorLogin: dbAdminUsername
    administratorLoginPassword: dbAdminPassword
    version: '15'
    storage: {
      storageSizeGB: 32
    }
    backup: {
      backupRetentionDays: 7
      geoRedundantBackup: 'Disabled'
    }
    highAvailability: {
      mode: 'Disabled'
    }
  }
}

// PostgreSQL Firewall Rule - Allow Azure Services
resource firewallRule 'Microsoft.DBforPostgreSQL/flexibleServers/firewallRules@2022-12-01' = {
  parent: postgresServer
  name: 'AllowAzureServices'
  properties: {
    startIpAddress: '0.0.0.0'
    endIpAddress: '0.0.0.0'
  }
}

// PostgreSQL Database
resource database 'Microsoft.DBforPostgreSQL/flexibleServers/databases@2022-12-01' = {
  parent: postgresServer
  name: databaseName
  properties: {
    charset: 'UTF8'
    collation: 'en_US.utf8'
  }
}

// Container App
resource containerApp 'Microsoft.App/containerApps@2023-05-01' = {
  name: '${environmentName}-app'
  location: location
  properties: {
    managedEnvironmentId: containerAppEnv.id
    configuration: {
      ingress: {
        external: true
        targetPort: 8080
        transport: 'http'
        allowInsecure: false
      }
      registries: [
        {
          server: containerRegistry.properties.loginServer
          username: containerRegistry.listCredentials().username
          passwordSecretRef: 'registry-password'
        }
      ]
      secrets: [
        {
          name: 'registry-password'
          value: containerRegistry.listCredentials().passwords[0].value
        }
        {
          name: 'db-password'
          value: dbAdminPassword
        }
      ]
    }
    template: {
      containers: [
        {
          name: 'attestations-demo'
          image: '${containerRegistry.properties.loginServer}/${environmentName}:${imageTag}'
          resources: {
            cpu: json('0.5')
            memory: '1Gi'
          }
          env: [
            {
              name: 'PORT'
              value: '8080'
            }
            {
              name: 'DB_HOST'
              value: postgresServer.properties.fullyQualifiedDomainName
            }
            {
              name: 'DB_PORT'
              value: '5432'
            }
            {
              name: 'DB_USER'
              value: dbAdminUsername
            }
            {
              name: 'DB_PASSWORD'
              secretRef: 'db-password'
            }
            {
              name: 'DB_NAME'
              value: databaseName
            }
            {
              name: 'DB_SSLMODE'
              value: 'require'
            }
          ]
        }
      ]
      scale: {
        minReplicas: 1
        maxReplicas: 3
        rules: [
          {
            name: 'http-rule'
            http: {
              metadata: {
                concurrentRequests: '100'
              }
            }
          }
        ]
      }
    }
  }
}

// Outputs
output containerAppFQDN string = containerApp.properties.configuration.ingress.fqdn
output containerRegistryLoginServer string = containerRegistry.properties.loginServer
output postgresServerFQDN string = postgresServer.properties.fullyQualifiedDomainName
output databaseName string = database.name
