// drasi-infra.bicep — Drasi infrastructure module for Cosmos DB change-feed workload
// Add Azure resources required by your Drasi workload here (Cosmos DB, Key Vault, etc.).

@description('The Azure region for all resources.')
param location string

@description('Environment name suffix applied to all resource names.')
param environmentName string

@description('Tags to apply to all resources.')
param tags object
