---
applyTo: "**/*.ps1,**/*.psm1"
description: "PowerShell cmdlet and scripting best practices based on Microsoft guidelines"
---

# PowerShell Cmdlet Development Guidelines

This guide provides PowerShell-specific instructions to help GitHub Copilot generate idiomatic, safe, and maintainable scripts. It aligns with Microsoft's PowerShell cmdlet development guidelines.

**IMPORTANT**: Use the `iseplaybook` MCP server for automation best practices and the `microsoft.learn.mcp` MCP server for Azure/PowerShell guidance when version- or platform-dependent behavior is involved.

## Safe Automation Defaults

- Support **dry-run** mode and require an explicit `-Apply`/`-Write`/`-Update` switch for mutations.
- Calculate and display changes before executing them.
- Avoid prompts for secrets; accept via environment or managed identity.
- Never assume default subscriptions or tenants; require explicit parameters or context selection.

### Execution Safety

- Do not mutate without an explicit intent flag (`-Apply`, `-Write`, `-Update`).
- Compute and display a change plan before execution.
- Provide a read-only or `-WhatIf`/`-DryRun` mode in scripts that can mutate state.

### Change Workflow

- Discover → Measure → Recommend → Compare (Current vs Optimal) → Flag `changeRequired` → Apply only if explicitly requested.

### Data-Driven Rules

- Put business rules (pricing, thresholds, tiers, policy mappings) in hashtables/objects near the top of the script.
- Avoid hard-coded if/else decision trees; make calculations explainable/inspectable.

## Naming Conventions

### Cmdlet Names

- **Verb-Noun Format:**
  - Use approved PowerShell verbs (check with `Get-Verb`)
  - Use singular nouns
  - PascalCase for both verb and noun
  - Avoid special characters and spaces

### Parameter Names

- Use PascalCase
- Choose clear, descriptive names
- Use singular form unless always multiple
- Follow PowerShell standard names (e.g., `Path`, `Name`, `Force`)

### Variable Names

- Use PascalCase for public variables
- Use camelCase for private variables
- Avoid abbreviations
- Use meaningful names

### Alias Avoidance

- Use full cmdlet names in scripts
- Avoid aliases (e.g., use `Get-ChildItem` instead of `gci`, `dir`, or `ls`)
- Use `Where-Object` instead of `?` or `where`
- Use `ForEach-Object` instead of `%`
- Document any custom aliases
- Use full parameter names

**Note:** Aliases are acceptable for interactive shell use but should be avoided in production scripts.

### Naming Example

```powershell
function Get-UserProfile {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Username,

        [Parameter()]
        [ValidateSet('Basic', 'Detailed')]
        [string]$ProfileType = 'Basic'
    )

    process {
        # Logic here
    }
}
```

## Parameter Design

### Standard Parameters

- Use common parameter names (`Path`, `Name`, `Force`, `Verbose`, `WhatIf`)
- Follow built-in cmdlet conventions
- Use parameter aliases sparingly for specialized terms
- Document parameter purpose clearly

### Parameter Names and Types

- Use singular form unless always multiple
- Choose clear, descriptive names
- Follow PowerShell conventions
- Use PascalCase formatting
- Use common .NET types
- Implement proper validation attributes

### Validation

- Use `ValidateSet` for limited options
- Use `ValidateNotNullOrEmpty()` for required strings
- Use `ValidateRange()` for numeric constraints
- Use `ValidateScript()` for custom validation
- Enable tab completion where possible

### Switch Parameters

- Use `[switch]` for boolean flags
- Avoid `$true`/`$false` parameters
- Default to `$false` when omitted
- Use clear action names (e.g., `-Force`, `-PassThru`, `-Recurse`)

### Parameter Design Example

```powershell
function Set-ResourceConfiguration {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [ValidateNotNullOrEmpty()]
        [string]$Name,

        [Parameter()]
        [ValidateSet('Dev', 'Test', 'Prod')]
        [string]$Environment = 'Dev',

        [Parameter()]
        [switch]$Force,

        [Parameter()]
        [ValidateNotNullOrEmpty()]
        [string[]]$Tags
    )

    process {
        # Logic here
    }
}
```

## Pipeline and Output

### Pipeline Input

- Use `ValueFromPipeline` for direct object input
- Use `ValueFromPipelineByPropertyName` for property mapping
- Implement `Begin`/`Process`/`End` blocks for pipeline handling
- Document pipeline input requirements in help

### Output Objects

- Return rich objects, not formatted text
- Use `PSCustomObject` for structured data
- Avoid `Write-Host` for data output
- Enable downstream cmdlet processing
- Use `Write-Output` explicitly when needed
- Emit structured objects suitable for CSV/JSON export; avoid `Write-Host` as the only output
- Include FinOps outputs when applicable (cost deltas, savings potential, SKU recommendations)

### Pipeline Streaming

- Output one object at a time in the `process` block
- Use `process` block for streaming operations
- Avoid collecting large arrays in memory
- Enable immediate downstream processing

### PassThru Pattern

- Default to no output for action cmdlets (Set-, New-, Remove-)
- Implement `-PassThru` switch for object return
- Return modified/created object when `-PassThru` is specified
- Use `Write-Verbose` or `Write-Warning` for status updates

### Pipeline Example

```powershell
function Update-ResourceStatus {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, ValueFromPipeline, ValueFromPipelineByPropertyName)]
        [string]$Name,

        [Parameter(Mandatory)]
        [ValidateSet('Active', 'Inactive', 'Maintenance')]
        [string]$Status,

        [Parameter()]
        [switch]$PassThru
    )

    begin {
        Write-Verbose 'Starting resource status update process'
        $timestamp = Get-Date
    }

    process {
        # Process each resource individually
        Write-Verbose "Processing resource: $Name"

        $resource = [PSCustomObject]@{
            Name        = $Name
            Status      = $Status
            LastUpdated = $timestamp
            UpdatedBy   = $env:USERNAME
        }

        # Only output if PassThru is specified
        if ($PassThru.IsPresent) {
            Write-Output $resource
        }
    }

    end {
        Write-Verbose 'Resource status update process completed'
    }
}
```

## Script Structure

### Comment-Based Help

Include comment-based help for all public-facing functions and cmdlets:

```powershell
<#
.SYNOPSIS
    Brief description of what the script does.

.DESCRIPTION
    Detailed description of the script's purpose, behavior, and usage.

.PARAMETER Name
    Description of the Name parameter.

.PARAMETER Force
    Skip confirmation prompts.

.EXAMPLE
    .\Script.ps1 -Name "Example"
    Runs the script with the specified name.

.EXAMPLE
    .\Script.ps1 -Name "Example" -Force
    Runs the script without confirmation prompts.

.OUTPUTS
    System.Management.Automation.PSCustomObject
    Returns an object with resource status information.

.NOTES
    Author: Your Name
    Version: 1.0.0
    Requires: PowerShell 5.1 or later
#>
```

### Parameter Definition

Use `[CmdletBinding()]` and proper parameter attributes:

```powershell
function Get-ResourceData {
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter(Mandatory = $true, Position = 0, ValueFromPipeline)]
        [ValidateNotNullOrEmpty()]
        [string]$Name,

        [Parameter(Mandatory = $false)]
        [ValidateRange(1, 100)]
        [int]$Count = 10,

        [Parameter()]
        [switch]$Force
    )

    begin {
        Write-Verbose "Starting resource data retrieval"
    }

    process {
        # Process each item from pipeline
    }

    end {
        Write-Verbose "Completed resource data retrieval"
    }
}
```

## Error Handling and Safety

### ShouldProcess Implementation

- Use `[CmdletBinding(SupportsShouldProcess = $true)]` for cmdlets that make changes
- Set appropriate `ConfirmImpact` level (Low, Medium, High)
- Call `$PSCmdlet.ShouldProcess()` before system changes
- Use `ShouldContinue()` for additional confirmations when needed
- Respect `-WhatIf` and `-Confirm` parameters

### Message Streams

- `Write-Verbose` for operational details (shown with `-Verbose`)
- `Write-Warning` for warning conditions
- `Write-Error` for non-terminating errors
- `throw` or `$PSCmdlet.ThrowTerminatingError()` for terminating errors
- Avoid `Write-Host` except for user interface text
- `Write-Debug` for debugging information (shown with `-Debug`)

### Error Handling Pattern

- Use `try`/`catch` blocks for error management
- Set appropriate `ErrorAction` preferences
- Return meaningful error messages with context
- Use `ErrorVariable` when needed
- Include proper terminating vs non-terminating error handling
- In advanced functions with `[CmdletBinding()]`, prefer `$PSCmdlet.WriteError()` over `Write-Error`
- In advanced functions with `[CmdletBinding()]`, prefer `$PSCmdlet.ThrowTerminatingError()` over `throw`
- Construct proper `ErrorRecord` objects with category, target, and exception details
- Wrap external calls in `try`/`catch` and continue processing other subscriptions/resources
- Collect failures in a final summary object for pipeline/CI reporting
- Empty `catch {}` blocks are not allowed; handle, log safely, or rethrow.

### Non-Interactive Design

- Accept input via parameters
- Avoid `Read-Host` in scripts
- Support automation scenarios
- Document all required inputs
- Use `-Force` switch to bypass confirmations in automation
- Never prompt for secrets; use managed identity or environment variables

### Fail Fast

Set strict error handling at the start:

```powershell
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
```

### Error Handling Example

```powershell
function Remove-UserAccount {
    [CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory, ValueFromPipeline)]
        [ValidateNotNullOrEmpty()]
        [string]$Username,

        [Parameter()]
        [switch]$Force
    )

    begin {
        Write-Verbose 'Starting user account removal process'
        $ErrorActionPreference = 'Stop'
    }

    process {
        try {
            # Validation
            if (-not (Test-UserExists -Username $Username)) {
                $errorRecord = [System.Management.Automation.ErrorRecord]::new(
                    [System.Exception]::new("User account '$Username' not found"),
                    'UserNotFound',
                    [System.Management.Automation.ErrorCategory]::ObjectNotFound,
                    $Username
                )
                $PSCmdlet.WriteError($errorRecord)
                return
            }

            # Confirmation
            $shouldProcessMessage = "Remove user account '$Username'"
            if ($Force -or $PSCmdlet.ShouldProcess($Username, $shouldProcessMessage)) {
                Write-Verbose "Removing user account: $Username"

                # Main operation
                Remove-ADUser -Identity $Username -ErrorAction Stop
                Write-Warning "User account '$Username' has been removed"
            }
        } catch [Microsoft.ActiveDirectory.Management.ADException] {
            $errorRecord = [System.Management.Automation.ErrorRecord]::new(
                $_.Exception,
                'ActiveDirectoryError',
                [System.Management.Automation.ErrorCategory]::NotSpecified,
                $Username
            )
            $PSCmdlet.ThrowTerminatingError($errorRecord)
        } catch {
            $errorRecord = [System.Management.Automation.ErrorRecord]::new(
                $_.Exception,
                'UnexpectedError',
                [System.Management.Automation.ErrorCategory]::NotSpecified,
                $Username
            )
            $PSCmdlet.ThrowTerminatingError($errorRecord)
        }
    }

    end {
        Write-Verbose 'User account removal process completed'
    }
}
```

### Simple Try/Catch Example

```powershell
try {
    $result = Invoke-RestMethod -Uri $uri -Method Get -ErrorAction Stop
    Write-Verbose "Request successful: $($result.message)"
} catch {
    Write-Error "Failed to make request: $_"
    throw
} finally {
    # Cleanup code
}
```

## Output and Logging

### Use Appropriate Output Cmdlets

```powershell
# Return data objects (preferred)
Write-Output $resultObject

# User-facing status messages (use sparingly)
Write-Host "Processing item: $name" -ForegroundColor Green

# Detailed debugging information
Write-Verbose "Connecting to $endpoint"

# Debug information
Write-Debug "Variable value: $debugInfo"

# Warnings
Write-Warning "Configuration file not found, using defaults"

# Non-terminating errors
Write-Error "Operation failed: $reason"

# Progress for long operations
Write-Progress -Activity "Processing" -Status "Item $i of $total" -PercentComplete (($i / $total) * 100)
```

### Sensitive/PII Logging Guardrails

- Do not log raw emails, user principal names, object IDs, access tokens, or connection strings.
- Mask identifiers when logging is required for troubleshooting (for example first 3/last 3 characters).
- Use structured objects for machine-readable diagnostics and redact sensitive fields before output.

### Timestamps for Long Operations

```powershell
function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Verbose "[$timestamp] $Message"
}
```

## Azure CLI Integration

### Check Prerequisites

```powershell
# Check for Azure CLI
if (-not (Get-Command az -ErrorAction SilentlyContinue)) {
    Write-Error "Azure CLI is not installed. Please install it from https://aka.ms/install-azure-cli"
    exit 1
}

# Verify login
$account = az account show --output json 2>$null | ConvertFrom-Json
if (-not $account) {
    Write-Error "Not logged in to Azure. Run 'az login' first."
    exit 1
}

Write-Host "Using subscription: $($account.name)"
```

### Parse JSON Output

```powershell
$resources = az resource list --output json | ConvertFrom-Json

if ($LASTEXITCODE -ne 0) {
    Write-Error "Azure CLI command failed"
    exit 1
}

foreach ($resource in $resources) {
    Write-Host "Resource: $($resource.name) ($($resource.type))"
}
```

## File and Path Handling

### Cross-Platform Paths

```powershell
# Use Join-Path for path construction
$configPath = Join-Path -Path $PSScriptRoot -ChildPath "config" | Join-Path -ChildPath "settings.json"

# Use $PSScriptRoot for script-relative paths
$scriptDir = $PSScriptRoot
```

### Check File Existence

```powershell
if (-not (Test-Path -Path $filePath -PathType Leaf)) {
    Write-Error "File does not exist: $filePath"
    exit 1
}
```

## Functions

### Define Reusable Functions

```powershell
function Get-ConfigValue {
    <#
    .SYNOPSIS
        Gets a configuration value from environment or file.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Name,

        [Parameter()]
        [string]$Default = ""
    )

    $envValue = [Environment]::GetEnvironmentVariable($Name)
    if ($envValue) {
        return $envValue
    }

    return $Default
}
```

## azd Integration

### Get Environment Values

```powershell
$envName = azd env get-values --output json | ConvertFrom-Json
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to get azd environment values"
    exit 1
}

$apiEndpoint = $envName.API_ENDPOINT
```

## Documentation and Style

### Comment-Based Help Requirements

Include comment-based help for any public-facing function or cmdlet. Inside the function, add a `<# ... #>` help comment with at least:

- `.SYNOPSIS` Brief description
- `.DESCRIPTION` Detailed explanation
- `.PARAMETER` descriptions for each parameter
- `.EXAMPLE` sections with practical usage
- `.OUTPUTS` Type of output returned
- `.NOTES` Additional information (author, version, requirements)

### Consistent Formatting

- Follow consistent PowerShell style
- Use proper indentation (4 spaces recommended)
- Opening braces on same line as statement
- Closing braces on new line aligned with statement
- Use line breaks after pipeline operators for readability
- PascalCase for function and parameter names
- Avoid unnecessary whitespace

### Pipeline Support

- Implement `Begin`/`Process`/`End` blocks for pipeline functions
- Use `ValueFromPipeline` where appropriate
- Support pipeline input by property name when logical
- Return proper objects, not formatted text

## Actionable Patterns

### Pattern 1: ErrorActionPreference (Error Handling)

**❌ WRONG: Using 'Continue' allows errors to pass silently**

```powershell
$ErrorActionPreference = 'Continue'  # ⚠️ Script continues after errors!
$result = Invoke-RestMethod -Uri $uri  # Error hidden in stream
# Script continues with invalid $result
```

**✅ CORRECT: Use 'Stop' to fail fast on errors**

```powershell
$ErrorActionPreference = 'Stop'  # ✅ Fails on first error
Set-StrictMode -Version Latest   # ✅ Strict variable checking

try {
    $result = Invoke-RestMethod -Uri $uri
    # Process $result safely
} catch {
    Write-Error "API request failed: $_"
    throw
}
```

**Rule:** Set `$ErrorActionPreference = 'Stop'` at script start. Wrap non-terminating cmdlets in `try/catch` for proper error handling.

---

### Pattern 2: Parameter Validation (Avoid Runtime Failures)

**❌ WRONG: Manual validation inside function (late failure)**

```powershell
function Set-ResourceConfig {
    param([string]$Environment)

    if ($Environment -notin @('Dev', 'Test', 'Prod')) {  # ⚠️ Fails at runtime
        throw "Invalid environment: $Environment"
    }
}
```

**✅ CORRECT: Declarative validation attributes (early failure)**

```powershell
function Set-ResourceConfig {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [ValidateSet('Dev', 'Test', 'Prod')]  # ✅ Fails before function execution
        [string]$Environment
    )
}
```

**❌ WRONG: Accepting null/empty values quietly**

```powershell
param([string]$Name)  # ⚠️ $Name can be null or empty string
```

**✅ CORRECT: Enforce non-null/non-empty with validation**

```powershell
param(
    [Parameter(Mandatory)]
    [ValidateNotNullOrEmpty()]  # ✅ Ensures value provided
    [string]$Name
)
```

**Rule:** Use `[ValidateSet()]`, `[ValidateNotNullOrEmpty()]`, `[ValidateRange()]`, `[ValidateScript()]` attributes instead of manual checks.

---

### Pattern 3: Pipeline Support (Process Block)

**❌ WRONG: Collecting all pipeline input in memory (high memory usage)**

```powershell
function Update-UserStatus {
    param([Parameter(ValueFromPipeline)]
        [string]$Username)

    # ⚠️ Waits for all input, then processes (high memory)
    $allUsers = @()
    $input | ForEach-Object { $allUsers += $_ }

    foreach ($user in $allUsers) {
        # Process user
    }
}
```

**✅ CORRECT: Stream processing with process block (memory efficient)**

```powershell
function Update-UserStatus {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, ValueFromPipeline)]
        [string]$Username
    )

    process {  # ✅ Processes each item immediately
        Write-Verbose "Processing user: $Username"
        # Process $Username
    }
}
```

**Rule:** Use `begin/process/end` blocks for pipeline functions. Never accumulate `$input` manually in memory.

---

### Pattern 4: ShouldProcess (Confirmation for Destructive Operations)

See also the comprehensive [Error Handling Example](#error-handling-example) for ShouldProcess combined with ErrorRecord construction and typed catch blocks.

**❌ WRONG: No confirmation for destructive operations**

```powershell
function Remove-ResourceItem {
    param([string]$ResourceId)

    Remove-AzResource -ResourceId $ResourceId  # ⚠️ No confirmation!
}
```

**✅ CORRECT: Implement ShouldProcess for safety**

```powershell
function Remove-ResourceItem {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory)]
        [string]$ResourceId
    )

    if ($PSCmdlet.ShouldProcess($ResourceId, 'Remove resource')) {
        Remove-AzResource -ResourceId $ResourceId -Force -ErrorAction Stop
    }
}
```

**Rule:** Use `[CmdletBinding(SupportsShouldProcess)]` for cmdlets that mutate state. Set `ConfirmImpact` based on operation severity (Low/Medium/High).

---

### Pattern 5: Approved Verbs (Discoverability & Consistency)

**❌ WRONG: Using unapproved verbs (breaks conventions)**

```powershell
function Validate-UserInput { }  # ⚠️ 'Validate' not approved
function Display-Report { }      # ⚠️ 'Display' not approved
function Fetch-Data { }          # ⚠️ 'Fetch' not approved
```

**✅ CORRECT: Use approved PowerShell verbs**

```powershell
function Test-UserInput { }   # ✅ 'Test' is approved
function Show-Report { }      # ✅ 'Show' is approved
function Get-Data { }         # ✅ 'Get' is approved
```

**Verification:**

```powershell
Get-Verb | Where-Object Verb -eq 'Validate'  # Returns nothing (unapproved)
Get-Verb | Where-Object Verb -eq 'Test'      # Returns approved verb
```

**Rule:** Use `Get-Verb` to find approved verbs. Common: Get, Set, New, Remove, Test, Invoke, Add, Update, Write.

---

### Pattern 6: Write-Host vs Write-Output (Data vs UI)

**❌ WRONG: Using Write-Host for data output (not pipeable)**

```powershell
function Get-UsersReport {
    foreach ($user in $users) {
        Write-Host "$($user.Name): $($user.Email)"  # ⚠️ Not pipeable!
    }
}
```

**✅ CORRECT: Return structured objects (pipeable)**

```powershell
function Get-UsersReport {
    [CmdletBinding()]
    param()

    foreach ($user in $users) {
        [PSCustomObject]@{  # ✅ Structured output
            Name  = $user.Name
            Email = $user.Email
        }
    }
}

# Enable downstream processing:
Get-UsersReport | Where-Object Name -like '*Admin*' | Export-Csv report.csv
```

**Rule:** Use `Write-Output` (or implicit return) for data. Reserve `Write-Host` for UI text only. Always return objects, not formatted strings.

---

### Pattern 7: Alias Avoidance in Scripts (Readability & Portability)

**❌ WRONG: Using aliases in production scripts (unclear)**

```powershell
$files = gci -r | ? { $_.Length -gt 1MB } | % { $_.FullName }  # ⚠️ Unreadable
```

**✅ CORRECT: Full cmdlet names (clear & portable)**

```powershell
$files = Get-ChildItem -Recurse |
    Where-Object { $_.Length -gt 1MB } |
    ForEach-Object { $_.FullName }  # ✅ Self-documenting
```

**Rule:** Avoid aliases (`gci`, `?`, `%`, `dir`, `ls`, `where`, `select`) in scripts. Use full cmdlet names for clarity.

---

### Pattern 8: ErrorRecord Construction (Advanced Error Handling)

**❌ WRONG: Using throw with string (loses context)**

```powershell
function Get-UserData {
    param([string]$Username)

    if (-not (Test-Path $userFile)) {
        throw "User file not found"  # ⚠️ No error category, target, or ID
    }
}
```

**✅ CORRECT: Construct proper ErrorRecord (rich error details)**

```powershell
function Get-UserData {
    [CmdletBinding()]
    param([string]$Username)

    if (-not (Test-Path $userFile)) {
        $errorRecord = [System.Management.Automation.ErrorRecord]::new(
            [System.Exception]::new("User file not found for '$Username'"),
            'UserFileNotFound',                                  # ErrorId
            [System.Management.Automation.ErrorCategory]::ObjectNotFound,
            $Username                                             # TargetObject
        )
        $PSCmdlet.ThrowTerminatingError($errorRecord)  # ✅ Rich error details
    }
}
```

**Rule:** In `[CmdletBinding()]` functions, use `$PSCmdlet.WriteError()` for non-terminating errors and `$PSCmdlet.ThrowTerminatingError()` for terminating errors with proper ErrorRecord objects.

---

## Best Practices

1. **Idempotency**: Scripts should be safe to run multiple times without side effects
2. **Validation**: Validate inputs using parameter validation attributes (not manual checks)
3. **Cleanup**: Use `finally` blocks or trap statements for cleanup operations
4. **Exit Codes**: Return meaningful exit codes (0 for success, non-zero for failure)
5. **Confirmation**: Use `ShouldProcess` for destructive operations (with proper `ConfirmImpact`)
6. **Error Handling**: Set `$ErrorActionPreference = 'Stop'` and implement comprehensive try/catch blocks
7. **Pipeline Support**: Use `begin/process/end` blocks; avoid accumulating $input in memory
8. **Output Design**: Return structured objects (PSCustomObject), not formatted text; use `-PassThru` for action cmdlets
9. **Verbose Support**: Provide verbose output for operational details (`Write-Verbose`)
10. **Non-Interactive**: Design for automation; avoid `Read-Host`; use parameters and environment variables
11. **Graceful Termination**: Avoid `exit` inside reusable functions; return errors/throw and let script entrypoint decide final exit code

## Control Plane vs Data Plane

- Prefer declarative reconciliation (ARM/Bicep/Terraform, policy) over imperative step sequencing for control-plane changes when feasible.
- Use Az cmdlets for data plane discovery (metrics, inventory, resource properties).
- Use ARM REST via `Invoke-RestMethod` when SDKs/cmdlets lag or are incomplete.
- Acquire tokens from the current Az context; never use manual auth in scripts.

## CI/CD Integration

- Must run non-interactively in GitHub Actions and Azure DevOps.
- Emit outputs via `$GITHUB_OUTPUT` (GitHub) or `##vso[task.setvariable]` (ADO).
- Do not require interactive login in pipelines.

## ARM Batch Support

- Provide helper functions like `New-ArmBatchRequest` and `Invoke-ArmBatch` for bulk ARM calls.
- Batch size must be configurable (default 10–20).
- Validate each inner `httpStatusCode` and capture failures per item.
- No destructive operations without an explicit intent flag (`-Apply`/`-Update`/`-Write`).

## Full Example: End-to-End Cmdlet Pattern

```powershell
function New-Resource {
    <#
    .SYNOPSIS
        Creates a new resource with specified configuration.

    .DESCRIPTION
        The New-Resource cmdlet creates a new resource in the specified environment.
        It supports pipeline input and implements proper confirmation for safe operation.

    .PARAMETER Name
        The name of the resource to create. This parameter is mandatory and accepts
        pipeline input.

    .PARAMETER Environment
        The target environment for the resource. Valid values are 'Development' and
        'Production'. Defaults to 'Development'.

    .EXAMPLE
        New-Resource -Name "MyResource" -Environment Production
        Creates a new resource named "MyResource" in the Production environment.

    .EXAMPLE
        "Resource1", "Resource2" | New-Resource -Environment Development
        Creates multiple resources from pipeline input.

    .OUTPUTS
        System.Management.Automation.PSCustomObject
        Returns an object containing the resource name, environment, and creation timestamp.

    .NOTES
        Author: Your Name
        Version: 1.0.0
        Requires: PowerShell 5.1 or later
    #>
    [CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Name,

        [Parameter()]
        [ValidateSet('Development', 'Production')]
        [string]$Environment = 'Development'
    )

    begin {
        Write-Verbose 'Starting resource creation process'
        $ErrorActionPreference = 'Stop'
    }

    process {
        try {
            if ($PSCmdlet.ShouldProcess($Name, "Create new resource in $Environment")) {
                Write-Verbose "Creating resource: $Name in $Environment"

                # Resource creation logic here
                $resource = [PSCustomObject]@{
                    PSTypeName  = 'Custom.Resource'
                    Name        = $Name
                    Environment = $Environment
                    Created     = Get-Date
                    CreatedBy   = $env:USERNAME
                }

                Write-Verbose "Resource '$Name' created successfully"
                Write-Output $resource
            }
        } catch {
            $errorRecord = [System.Management.Automation.ErrorRecord]::new(
                $_.Exception,
                'ResourceCreationFailed',
                [System.Management.Automation.ErrorCategory]::NotSpecified,
                $Name
            )
            $PSCmdlet.ThrowTerminatingError($errorRecord)
        }
    }

    end {
        Write-Verbose 'Completed resource creation process'
    }
}
```

## Testing

### Pester Tests

```powershell
Describe "Get-ConfigValue" {
    It "Returns environment variable when set" {
        $env:TEST_VALUE = "expected"
        Get-ConfigValue -Name "TEST_VALUE" | Should -Be "expected"
    }

    It "Returns default when environment variable not set" {
        Remove-Item Env:TEST_VALUE -ErrorAction SilentlyContinue
        Get-ConfigValue -Name "TEST_VALUE" -Default "default" | Should -Be "default"
    }

    It "Accepts pipeline input" {
        "TestValue" | Get-ConfigValue -Default "default" | Should -Be "default"
    }

    It "Supports WhatIf" {
        { New-Resource -Name "Test" -WhatIf } | Should -Not -Throw
    }
}
```

## References

- [PowerShell Documentation](https://learn.microsoft.com/powershell/)
- [PowerShell Best Practices](https://poshcode.gitbook.io/powershell-practice-and-style/)
- [Azure PowerShell Documentation](https://learn.microsoft.com/powershell/azure/)
- [ISE Shell Checklist](https://microsoft.github.io/code-with-engineering-playbook/code-reviews/recipes/bash/)
