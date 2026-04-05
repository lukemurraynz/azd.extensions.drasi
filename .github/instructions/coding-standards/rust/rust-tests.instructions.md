---
applyTo: "**/*.rs"
description: "Required instructions for Rust test code research, planning, implementation, editing, or creating - Brought to you by microsoft/hve-core"
---

# Rust Test Instructions

Conventions for Rust test code. All conventions from [rust.instructions.md](rust.instructions.md) apply, including naming, error handling, and module structure.

**IMPORTANT**: Use the `iseplaybook` MCP server for ISE testing best practices. Use `context7` MCP server to verify test crate APIs (mockall, wiremock, rstest, testcontainers) against installed versions.

## Test Module Placement

Place unit tests in `#[cfg(test)] mod tests` within the source file they exercise:

<!-- <example-tests> -->

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn given_valid_input_parse_returns_config() {
        let json = r#"{"endpoint": "https://example.com"}"#;
        let config: AppConfig = serde_json::from_str(json).unwrap();
        assert_eq!(config.polling_interval_secs, 10);
    }

    #[tokio::test]
    async fn when_endpoint_available_fetch_returns_data() {
        let service = PollingService::new(AppConfig::from_env());
        let result = service.fetch().await;
        assert!(result.is_ok(), "fetch should succeed when endpoint is available");
    }
}
```

<!-- </example-tests> -->

## Test Naming

Test method format: `given_context_when_action_then_expected` or descriptive snake_case that reads as a behavior statement.

```text
given_valid_input_parse_returns_config
when_endpoint_unavailable_send_returns_error
parses_empty_payload_as_default
```

Prefer one assertion per test. Related assertions validating the same behavior are acceptable.

## Mocking and Testing Libraries

| Library             | Usage                                                   |
| ------------------- | ------------------------------------------------------- |
| `mockall`           | Preferred for trait-based mocking                       |
| `wiremock`          | HTTP server mocking in async tests                      |
| `mockito`           | Lightweight HTTP mocking for synchronous or async tests |
| `rstest`            | Parameterized and fixture-based tests                   |
| `testcontainers`    | Container-based integration tests (databases, queues)   |
| `serial_test`       | Force sequential execution for tests with shared state  |
| `pretty_assertions` | Readable assertion diffs with colored output            |

Use `mockall` to generate mock implementations from traits via `#[automock]`:

```rust
use mockall::automock;

// Application types — defined in your crate (see rust.instructions.md)
pub struct Item {
    pub id: String,
}

// Uses the module-scoped Result alias from rust.instructions.md
#[automock]
pub trait Repository: Send + Sync {
    fn find_by_id(&self, id: &str) -> Result<Option<Item>>;
}

pub struct ItemService {
    repo: Box<dyn Repository>,
}

impl ItemService {
    pub fn new(repo: Box<dyn Repository>) -> Self {
        Self { repo }
    }

    pub fn get(&self, id: &str) -> Result<Option<Item>> {
        self.repo.find_by_id(id)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use mockall::predicate::*;

    #[test]
    fn given_existing_item_service_returns_it() {
        let mut mock = MockRepository::new();
        mock.expect_find_by_id()
            .with(eq("42"))
            .returning(|_| Ok(Some(Item { id: "42".into() })));

        let service = ItemService::new(Box::new(mock));
        let result = service.get("42").unwrap();
        assert_eq!(result.unwrap().id, "42");
    }
}
```

Use `wiremock` to mock HTTP servers in async tests:

```rust
use wiremock::{MockServer, Mock, ResponseTemplate};
use wiremock::matchers::method;

#[tokio::test]
async fn when_api_returns_ok_fetch_succeeds() {
    let mock_server = MockServer::start().await;

    Mock::given(method("GET"))
        .respond_with(ResponseTemplate::new(200).set_body_string(r#"{"id": "1"}"#))
        .mount(&mock_server)
        .await;

    let client = reqwest::Client::new();
    let response = client.get(mock_server.uri()).send().await.unwrap();
    assert_eq!(response.status(), 200);
}
```

Add test dependencies to `[dev-dependencies]` in `Cargo.toml`:

```toml
[dev-dependencies]
mockall = "0.13"
reqwest = { version = "0.13", features = ["json"] }
tokio = { version = "1", features = ["macros", "rt"] }
wiremock = "0.6"
```

## Test Data Patterns

Use builder functions or fixture helpers for test data rather than repeating inline construction:

```rust
#[cfg(test)]
mod tests {
    use super::*;

    fn sample_config() -> AppConfig {
        AppConfig {
            endpoint: "https://example.com".into(),
            polling_interval_secs: 10,
        }
    }

    #[test]
    fn given_custom_interval_config_uses_override() {
        let config = AppConfig {
            polling_interval_secs: 30,
            ..sample_config()
        };
        assert_eq!(config.polling_interval_secs, 30);
    }
}
```

Inline construction is acceptable for simple one-field tests where a builder adds no clarity.

## Integration Tests

Place integration tests in the `tests/` directory at the crate root. Each file in `tests/` compiles as a separate crate with access to the library's public API only:

```rust
// tests/polling_integration.rs
use my_service::AppConfig;

#[tokio::test]
async fn given_valid_config_service_starts() {
    let config = AppConfig {
        endpoint: "https://example.com".into(),
        polling_interval_secs: 1,
    };
    assert!(!config.endpoint.is_empty());
}
```

## Parameterized Tests

Use `rstest` for fixture injection and parameterized test cases:

```rust
use rstest::rstest;

#[rstest]
#[case("user@example.com", true)]
#[case("not-an-email", false)]
#[case("", false)]
fn given_input_validate_email_returns_expected(#[case] input: &str, #[case] expected: bool) {
    assert_eq!(validate_email(input), expected);
}
```

`rstest` fixtures compose with `#[tokio::test]` via `#[rstest] #[tokio::test]` stacking.

## Container-Based Integration Tests

Use `testcontainers` for tests that need real infrastructure (databases, message queues). Gate behind a feature or build flag so `cargo test` stays fast by default:

```rust
//! tests/db_integration.rs
//!
//! Run with: `cargo test --test db_integration` (or via CI job)

use testcontainers::{runners::AsyncRunner, GenericImage};

#[tokio::test]
async fn given_postgres_available_repo_creates_record() {
    let container = GenericImage::new("postgres", "16-alpine")
        .with_env_var("POSTGRES_PASSWORD", "test")
        .start()
        .await
        .expect("postgres container should start");

    let port = container.get_host_port_ipv4(5432).await.unwrap();
    let conn_str = format!("postgres://postgres:test@127.0.0.1:{port}/postgres");

    // Use conn_str to exercise repository layer against a real database
    assert!(!conn_str.is_empty());
}
```

## Shared Test Utilities

In workspace projects with many crates, extract common test helpers (fixtures, builders, mock implementations) into a dedicated `shared-tests` crate:

```toml
# Cargo.toml (workspace root)
[workspace]
members = ["core", "api", "shared-tests"]

# shared-tests/Cargo.toml — never published, test-only
[package]
name = "shared-tests"
publish = false

[dependencies]
mockall = "0.13"
serde_json = "1"
```

Consume via `dev-dependencies` in other crates: `shared-tests = { path = "../shared-tests" }`.

## Test Conventions

- Use `#[tokio::test]` for async tests.
- Prefer assertion messages that explain intent: `assert!(result.is_ok(), "should parse valid JSON")`.
- Use builder functions or fixture helpers for test data rather than repeating inline construction.
- Place integration tests in the `tests/` directory at the crate root.

## Complete Example

Types referenced below (`AppConfig`, `ServiceError`, `Result` alias) are defined in [rust.instructions.md](rust.instructions.md).

```rust
#[cfg(test)]
mod tests {
    use super::*;

    // Fixture helper — see Test Data Patterns
    fn sample_config() -> AppConfig {
        AppConfig {
            endpoint: "https://example.com".into(),
            polling_interval_secs: 10,
        }
    }

    #[test]
    fn given_defaults_config_has_ten_second_interval() {
        let config = sample_config();
        assert_eq!(config.polling_interval_secs, 10);
    }

    #[test]
    fn service_error_not_found_formats_message() {
        let err = ServiceError::not_found("item 42");
        assert_eq!(err.to_string(), "Not found: item 42");
    }

    #[tokio::test]
    async fn when_fetch_fails_error_contains_status() {
        let config = sample_config();
        let service = PollingService::new(config);
        let result = service.fetch().await;
        assert!(result.is_err(), "fetch should fail with unreachable endpoint");
    }
}
```
