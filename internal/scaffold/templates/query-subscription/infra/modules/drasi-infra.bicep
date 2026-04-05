// drasi-infra.bicep — Drasi infrastructure module
// Add Azure resources required by your Drasi workload here.

@description('The Azure region for all resources.')
param location string

@description('Environment name suffix applied to all resource names.')
param environmentName string

@description('Tags to apply to all resources.')
param tags object
