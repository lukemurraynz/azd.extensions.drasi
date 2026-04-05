---
applyTo: "**/*.tftest.hcl"
description: "Terraform native testing framework best practices for mock providers, plan/apply assertions, integration tests, and CI pipeline integration"
---

# Terraform Test Instructions

All conventions from [terraform.instructions.md](terraform.instructions.md) apply, including security defaults, module structure, validation pipeline, and AVM patterns.

**IMPORTANT**: Use the `iseplaybook` MCP server for ISE IaC testing best practices. Use `context7` MCP server (`/websites/developer_hashicorp_terraform`) for Terraform test framework documentation. Use `microsoft.learn.mcp` for Azure-specific validation patterns.

## Test Pyramid for Infrastructure

| Level       | Scope                                      | Speed   | Owns                                         |
| ----------- | ------------------------------------------ | ------- | -------------------------------------------- |
| Unit        | Single module with mock providers (`plan`) | < 5 s   | Naming, variable validation, resource config |
| Integration | Real providers with `apply` + `destroy`    | < 5 min | Provider behavior, actual resource creation  |
| Validation  | Check blocks on live infrastructure        | < 30 s  | Runtime health, endpoint reachability        |

Unit tests run on every PR. Integration tests run in a dedicated CI job with credentials. Check blocks run on every plan/apply.

## Test File Organization

```text
modules/
  my-module/
    main.tf
    variables.tf
    outputs.tf
    versions.tf
    README.md
    tests/
      unit/
        naming.tftest.hcl         # unit: naming conventions
        validation.tftest.hcl     # unit: variable validation
        defaults.tftest.hcl       # unit: default values
      integration/
        deploy.tftest.hcl         # integration: real providers
    examples/
      basic/
        main.tf
```

- Place test files in `tests/` within the module directory
- Separate unit tests (mock providers) from integration tests (real providers)
- Every custom module must include at least one test file
- Run all tests with `terraform test` from the module root

## Unit Tests with Mock Providers

Use `mock_provider` to test module logic without credentials or real infrastructure:

```hcl
# tests/unit/naming.tftest.hcl

mock_provider "azurerm" {}

run "validates_resource_group_name" {
  command = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
  }

  assert {
    condition     = azurerm_resource_group.main.name == "rg-myapp-dev"
    error_message = "Resource group name did not match expected pattern: rg-myapp-dev"
  }
}

run "validates_storage_account_name" {
  command = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
  }

  assert {
    condition     = length(azurerm_storage_account.main.name) <= 24
    error_message = "Storage account name exceeds 24 character limit"
  }

  assert {
    condition     = can(regex("^[a-z0-9]+$", azurerm_storage_account.main.name))
    error_message = "Storage account name must be lowercase alphanumeric only"
  }
}
```

Rules:

- Use `command = plan` for unit tests (no infrastructure created)
- Test naming conventions, variable validation, and conditional logic
- Include descriptive `error_message` that shows both expected and actual values
- Mock every provider the module uses

## Variable Validation Tests

Test that variable validation rules reject invalid inputs:

```hcl
# tests/unit/validation.tftest.hcl

mock_provider "azurerm" {}

run "rejects_empty_project_name" {
  command = plan

  variables {
    project     = ""
    environment = "dev"
    location    = "uksouth"
  }

  expect_failures = [
    var.project,
  ]
}

run "rejects_invalid_environment" {
  command = plan

  variables {
    project     = "myapp"
    environment = "staging"  # not in allowed list
    location    = "uksouth"
  }

  expect_failures = [
    var.environment,
  ]
}

run "accepts_valid_configuration" {
  command = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
  }

  # No expect_failures — this must succeed
  assert {
    condition     = azurerm_resource_group.main.location == "uksouth"
    error_message = "Location should be uksouth"
  }
}
```

Rules:

- Use `expect_failures` to test that validation rules reject bad input
- Test both positive (valid input accepted) and negative (invalid input rejected) cases
- Reference the specific variable (`var.project`) in `expect_failures`

## Default Value Tests

Verify optional variables default correctly:

```hcl
# tests/unit/defaults.tftest.hcl

mock_provider "azurerm" {}

run "uses_default_sku_when_not_specified" {
  command = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
    # sku intentionally omitted — should use default
  }

  assert {
    condition     = azurerm_service_plan.main.sku_name == "B1"
    error_message = "Default SKU should be B1"
  }
}
```

## Security Default Tests

Verify that security defaults are enforced by the module:

```hcl
# tests/unit/security.tftest.hcl

mock_provider "azurerm" {}

run "enforces_tls_1_2" {
  command = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
  }

  assert {
    condition     = azurerm_storage_account.main.min_tls_version == "TLS1_2"
    error_message = "TLS 1.2 must be the minimum version"
  }

  assert {
    condition     = azurerm_storage_account.main.https_traffic_only_enabled == true
    error_message = "HTTPS-only traffic must be enabled"
  }
}

run "disables_public_access_by_default" {
  command = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
  }

  assert {
    condition     = azurerm_storage_account.main.public_network_access_enabled == false
    error_message = "Public network access must be disabled by default"
  }
}
```

## Integration Tests (Real Providers)

Use real providers for tests that validate actual Azure behavior:

```hcl
# tests/integration/deploy.tftest.hcl

provider "azurerm" {
  features {}
}

variables {
  project     = "tftest"
  environment = "ci"
  location    = "uksouth"
}

run "deploys_resource_group" {
  command = apply

  assert {
    condition     = azurerm_resource_group.main.name == "rg-tftest-ci"
    error_message = "Resource group was not created with expected name"
  }
}

run "deploys_storage_account" {
  command = apply

  assert {
    condition     = azurerm_storage_account.main.account_tier == "Standard"
    error_message = "Storage account tier should be Standard"
  }
}
```

Rules:

- Run integration tests in a dedicated CI job with Azure credentials
- Use short-lived resource names (include `ci` or timestamp) to avoid conflicts
- Terraform automatically runs `destroy` after tests complete
- Gate production infrastructure changes on integration test pass

## Parallel Test Execution (Terraform 1.x)

Run independent test blocks in parallel for faster CI:

```hcl
# tests/unit/parallel.tftest.hcl

test {
  parallel = true
}

mock_provider "azurerm" {}

run "validates_naming" {
  state_key = "naming"
  command   = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
  }

  assert {
    condition     = azurerm_resource_group.main.name == "rg-myapp-dev"
    error_message = "Naming mismatch"
  }
}

run "validates_tags" {
  state_key = "tags"
  command   = plan

  variables {
    project     = "myapp"
    environment = "dev"
    location    = "uksouth"
  }

  assert {
    condition     = azurerm_resource_group.main.tags["environment"] == "dev"
    error_message = "Environment tag missing"
  }
}
```

Rules:

- Set `parallel = true` at the `test` block level for global opt-in
- Each parallel `run` block needs a unique `state_key`
- Sequential dependencies between `run` blocks (referencing `run.<name>.value`) are respected automatically
- Use `parallel = false` on a specific `run` to force sequential execution

## Setup and Helper Modules

Use helper modules for test setup (e.g., creating prerequisites):

```hcl
# tests/integration/full_stack.tftest.hcl

provider "azurerm" {
  features {}
}

run "setup_network" {
  module {
    source = "./tests/setup/network"
  }

  variables {
    location = "uksouth"
    prefix   = "tftest"
  }
}

run "deploy_app" {
  command = apply

  variables {
    project     = "tftest"
    environment = "ci"
    location    = "uksouth"
    subnet_id   = run.setup_network.subnet_id
  }

  assert {
    condition     = azurerm_linux_web_app.main.id != ""
    error_message = "Web app was not created"
  }
}
```

## CI Pipeline Integration

```yaml
# Minimum Terraform test pipeline
steps:
  - run: terraform fmt -check -recursive
  - run: terraform validate
  - run: terraform test # unit tests (mock providers)
  - run: terraform test -filter=tests/integration/ # integration (separate job, needs creds)
  - run: terraform plan -out=tfplan
  - run: terraform show -json tfplan > tfplan.json
```

- Run `terraform test` between `validate` and `plan` in CI
- Filter integration tests into a separate job with Azure credentials
- Use `-filter` to run specific test directories
- Gate production applies on unit + integration test pass

## Test Anti-Patterns

| Anti-Pattern                       | Why It's Harmful                                 | Fix                                                  |
| ---------------------------------- | ------------------------------------------------ | ---------------------------------------------------- |
| No mock providers                  | Unit tests require real credentials              | Use `mock_provider` for all unit tests               |
| Missing `error_message`            | Failures show only `condition = false`           | Add descriptive messages showing expected values     |
| Only testing happy path            | Invalid inputs not caught until apply            | Use `expect_failures` for validation rules           |
| No security assertions             | Module may ship without TLS, encryption defaults | Test security defaults explicitly                    |
| Mixing unit and integration tests  | `terraform test` runs everything, CI is slow     | Separate into `tests/unit/` and `tests/integration/` |
| Hard-coded resource names in tests | Tests conflict in shared subscriptions           | Use unique prefixes (`tftest-ci-{suffix}`)           |

## Final Self-Check (Before Proposing Test Changes)

✅ Every custom module has at least one `.tftest.hcl` file
✅ Unit tests use `mock_provider` and `command = plan`
✅ Variable validation tested with `expect_failures`
✅ Security defaults asserted (TLS, HTTPS, public access)
✅ Integration tests gated in separate CI job with credentials
✅ Descriptive `error_message` on every assertion
✅ `terraform test` included in CI between validate and plan
✅ Both positive and negative test cases covered
✅ Test file organization follows `tests/unit/` and `tests/integration/` pattern
