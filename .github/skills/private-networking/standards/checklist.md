# Private Networking Checklist

## VNet Design

- [ ] VNet address space planned for growth
- [ ] Subnets separated by function (compute, private endpoints, integration)
- [ ] NSGs associated with all subnets
- [ ] No overlapping address spaces with peered VNets

## Private Endpoints

- [ ] Private endpoints created for all PaaS services
- [ ] Private DNS zones created for each service type
- [ ] DNS zones linked to all relevant VNets
- [ ] DNS zone groups configured on private endpoints
- [ ] Public network access disabled on PaaS resources

## VNet Integration

- [ ] App Service/Functions VNet integration configured
- [ ] Integration subnet delegated to `Microsoft.Web/serverFarms`
- [ ] Integration subnet sized appropriately (minimum `/27` (`/26` recommended))
- [ ] Container Apps environment uses infrastructure subnet
- [ ] Container Apps subnet delegated to `Microsoft.App/environments`

## Network Security Groups

- [ ] Deny-all default rule at lowest priority
- [ ] Allow rules use specific CIDR ranges, not `*`
- [ ] NSG flow logs enabled for troubleshooting
- [ ] Service tags used where applicable (`AzureCloud`, `Storage`, etc.)

## Network Security Perimeter

- [ ] NSP created for PaaS-to-PaaS workload groups
- [ ] All supported PaaS resources associated with the NSP
- [ ] Initial deployment in Learning mode
- [ ] Inbound access rules scoped to required IP ranges
- [ ] Outbound access rules scoped to required FQDNs
- [ ] NSP diagnostic logs forwarded to Log Analytics
- [ ] Learning mode violations reviewed before switching to Enforced

## DNS Configuration

- [ ] Private DNS zones use standard naming (`privatelink.<service>.windows.net`)
- [ ] DNS zones are centralised (one zone per service type)
- [ ] VNet links created for all VNets that need resolution
- [ ] Custom DNS servers (if used) forward to Azure DNS (168.63.129.16)

## Infrastructure as Code

- [ ] VNet, subnets, and NSGs deployed via Bicep/Terraform
- [ ] Private endpoints and DNS zones in IaC
- [ ] NSP and associations in IaC
- [ ] Public access disablement configured in resource definitions
