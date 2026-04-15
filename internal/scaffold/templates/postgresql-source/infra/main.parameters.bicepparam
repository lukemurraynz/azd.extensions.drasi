using 'main.bicep'

param environmentName = readEnvironmentVariable('AZURE_ENV_NAME', 'dev')
// NOTE: location defaults to resourceGroup().location in main.bicep
// azd automatically sets AZURE_LOCATION from azd config or user input
// If you need to override, set AZURE_LOCATION in your .env file
