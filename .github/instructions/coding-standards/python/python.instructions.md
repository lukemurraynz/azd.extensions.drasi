---
applyTo: "**/*.py"
description: "Python development best practices following ISE Engineering Playbook guidelines"
---

# Python Code Instructions

Follow ISE Python Code Review Checklist and PEP 8 style guidelines.

**IMPORTANT**: Use the `iseplaybook` MCP server to get the latest Python best practices. Use `context7` MCP server for framework-specific documentation (FastAPI, Django, etc.). Do not assume—verify current guidance.

**Purpose and scope**: applies to all Python-based services and scripts owned by the ISE organization. For framework-specific deviations (FastAPI, Django, Flask), consult the `context7` MCP server for current overrides.

## Async/Await Governance

- Use `async`/`await` for I/O-bound work (HTTP, file, database).
- Avoid mixing threads with asyncio unless bridging legacy code.
- Always propagate cancellation (`asyncio.CancelledError`); do not swallow it.
- Use `asyncio.timeout()` (Python 3.11+) or `asyncio.wait_for()` for timeouts based on project version.
- Use `async with` for async context management (connection pools, sessions, transactions).
- Prefer structured task groups (Python 3.11+):

```python
import asyncio

async with asyncio.TaskGroup() as tg:
    tg.create_task(fetch_user(user_id))
```

- When integrating libraries that mix sync/async, consider `anyio` for compatibility.

## Code Style

- Use 4-space indentation
- Maximum line length: 88-120 characters (Black default: 88)
- Use snake_case for functions and variables
- Use PascalCase for classes
- Use UPPER_SNAKE_CASE for constants

## Type Hints

Use type hints for function signatures and complex variables:

- Prefer `from __future__ import annotations` for Python >= 3.9 targets.
- Use `TypedDict` or `@dataclass` instead of `Dict[str, Any]` for predictable contracts.
- Add `pytest-mypy` (or equivalent) to ensure annotated tests pass type consistency checks.

```python
from __future__ import annotations

from typing import Optional, List, Dict, Any

def get_user(user_id: int) -> Optional[User]:
    """Retrieve a user by ID."""
    return user_repository.find(user_id)

def process_items(items: List[str]) -> Dict[str, Any]:
    """Process a list of items and return results."""
    return {item: len(item) for item in items}
```

## Docstrings

Use Google-style docstrings:

- Add module-level docstrings for public modules (e.g., `main.py`, `config.py`).
- Use example-based docstrings (`>>>`) only for exposed library functions; avoid for internal modules.
- Optional: use `pdoc` or `sphinx-autodoc` for generated documentation aligned to this format.

```python
def calculate_discount(price: float, discount_percent: float) -> float:
    """Calculate the discounted price.

    Args:
        price: The original price.
        discount_percent: The discount percentage (0-100).

    Returns:
        The discounted price.

    Raises:
        ValueError: If discount_percent is not between 0 and 100.

    Example:
        >>> calculate_discount(100.0, 20.0)
        80.0
    """
    if not 0 <= discount_percent <= 100:
        raise ValueError("Discount must be between 0 and 100")
    return price * (1 - discount_percent / 100)
```

## Classes

```python
from dataclasses import dataclass
from typing import Optional

@dataclass
class User:
    """Represents a user in the system."""
    
    id: int
    name: str
    email: str
    is_active: bool = True
    
    def __post_init__(self) -> None:
        """Validate user data after initialization."""
        if not self.email or '@' not in self.email:
            raise ValueError("Invalid email address")
```

## Error Handling

Use specific exception types:

```python
class UserNotFoundError(Exception):
    """Raised when a user is not found."""
    pass

class ValidationError(Exception):
    """Raised when validation fails."""
    pass

def get_user_or_fail(user_id: int) -> User:
    """Get a user or raise an error."""
    user = get_user(user_id)
    if user is None:
        raise UserNotFoundError(f"User {user_id} not found")
    return user

- Prefer exception chaining to preserve root cause (`raise CustomError(...) from e`).
- Add graceful degradation or retry logic for transient errors (async HTTP/DB calls).
```

## Logging

Use the logging module with structured messages and correlation IDs:

```python
import logging

logger = logging.getLogger(__name__)

def process_order(order_id: str, correlation_id: str) -> None:
    """Process an order."""
    logger.info("Processing order", extra={"order_id": order_id, "correlation_id": correlation_id})
    try:
        # processing logic
        logger.info("Order processed successfully", extra={"order_id": order_id, "correlation_id": correlation_id})
    except Exception as e:
        logger.error("Failed to process order", extra={"order_id": order_id, "correlation_id": correlation_id}, exc_info=True)
        raise
```

- Standardize logging context keys (e.g., `component`, `correlation_id`) and centralize them via helper or `logging.Filter`.
- For larger services, prefer OpenTelemetry context propagation for correlation.

## Async/Await

For asynchronous code:

```python
import asyncio
from typing import List

async def fetch_user(user_id: int) -> User:
    """Fetch a user asynchronously."""
    async with httpx.AsyncClient() as client:
        response = await client.get(f"/api/users/{user_id}")
        response.raise_for_status()
        return User(**response.json())

async def fetch_users(user_ids: List[int]) -> List[User]:
    """Fetch multiple users concurrently."""
    tasks = [fetch_user(uid) for uid in user_ids]
    return await asyncio.gather(*tasks)
```

## Testing

### pytest Best Practices

```python
import pytest
from unittest.mock import Mock, patch

class TestUserService:
    """Tests for UserService."""

    @pytest.fixture
    def service(self):
        """Create a test service instance."""
        return UserService(repository=Mock())

    def test_get_user_returns_user(self, service):
        """Test that get_user returns a user when found."""
        # Arrange
        expected_user = User(id=1, name="Test", email="test@example.com")
        service.repository.find.return_value = expected_user

        # Act
        result = service.get_user(1)

        # Assert
        assert result == expected_user
        service.repository.find.assert_called_once_with(1)

    def test_get_user_returns_none_when_not_found(self, service):
        """Test that get_user returns None when user not found."""
        service.repository.find.return_value = None
        assert service.get_user(999) is None
```

- Use `pytest-asyncio` for async test patterns.
- Naming: `test_*.py` and `test_*` functions; avoid `*_test.py`.
- Enforce coverage in CI (`pytest --cov`, baseline >= 90%).
- Consider property-based testing with `hypothesis` for data-processing modules.

## Project Structure

```text
src/
  ├── __init__.py
  ├── main.py           # Application entry point
  ├── config.py         # Configuration management
  ├── models/           # Data models
  ├── schemas/          # Pydantic/OpenAPI schemas
  ├── services/         # Business logic
  ├── repositories/     # Data access
  ├── api/              # API endpoints
  ├── utils/            # Utility functions
  └── cli/              # CLI entry points / scripts
py.typed                # Package typing marker (when publishing)
tests/
  ├── __init__.py
  ├── conftest.py       # pytest fixtures
  ├── unit/             # Unit tests
  └── integration/      # Integration tests
```

## Actionable Patterns

### Pattern 1: Context Managers (with statement)

**❌ WRONG: Manual resource cleanup (easy to forget, error-prone)**
```python
file = open('data.txt', 'r')
try:
    data = file.read()
    # Process data
finally:
    file.close()  # ⚠️ Easy to forget, verbose
```

**✅ CORRECT: Use context managers (automatic cleanup)**
```python
with open('data.txt', 'r') as file:
    data = file.read()
    # Process data
# ✅ File automatically closed, even if exception occurs
```

**❌ WRONG: Not using context manager for database transactions**
```python
import sqlite3

con = sqlite3.connect(":memory:")
try:
    con.execute("INSERT INTO users(name) VALUES(?)", ("Alice",))
    con.commit()  # ⚠️ Manual commit, no automatic rollback on errors
except Exception:
    con.rollback()
finally:
    con.close()
```

**✅ CORRECT: Context manager handles commit/rollback**
```python
import sqlite3

con = sqlite3.connect(":memory:")
try:
    with con:  # ✅ Auto-commits on success, auto-rollbacks on exception
        con.execute("INSERT INTO users(name) VALUES(?)", ("Alice",))
finally:
    con.close()  # ✅ Still need manual close for connection
```

**Rule:** Use `with` statement for file I/O, database connections, locks, and any resource that needs cleanup. Guarantees `__exit__()` is called.

---

### Pattern 2: Exception Handling (Specificity & Chaining)

**❌ WRONG: Bare except catches everything (hides bugs)**
```python
try:
    result = int(user_input)
except:  # ⚠️ Catches KeyboardInterrupt, SystemExit, everything!
    print("Error occurred")
```

**✅ CORRECT: Catch specific exceptions**
```python
try:
    result = int(user_input)
except ValueError:  # ✅ Only catches conversion errors
    print("Invalid number format")
except KeyError as e:  # ✅ Specific error with context
    print(f"Missing key: {e}")
```

**❌ WRONG: Swallowing exceptions (loses root cause)**
```python
try:
    data = fetch_user_data(user_id)
except Exception:
    raise CustomError("Failed to process user")  # ⚠️ Lost original error!
```

**✅ CORRECT: Exception chaining preserves root cause**
```python
try:
    data = fetch_user_data(user_id)
except Exception as e:
    raise CustomError("Failed to process user") from e  # ✅ Preserves traceback
```

**Rule:** Never use bare `except:`. Always catch specific exceptions. Use `raise ... from e` to preserve exception chain.

---

### Pattern 3: Pathlib vs os.path (Modern Path Handling)

**❌ WRONG: Using os.path for file operations (string manipulation)**
```python
import os

config_dir = os.path.join(os.getcwd(), 'config', 'settings')
if os.path.exists(config_dir):
    with open(os.path.join(config_dir, 'app.json'), 'r') as f:
        # ⚠️ Verbose, error-prone string operations
        data = f.read()
```

**✅ CORRECT: Use pathlib for object-oriented path handling**
```python
from pathlib import Path

config_dir = Path.cwd() / 'config' / 'settings'  # ✅ Cleaner
if config_dir.exists():
    data = (config_dir / 'app.json').read_text()  # ✅ Chainable, less error-prone
```

**Rule:** Prefer `pathlib.Path` over `os.path` for all path operations. More readable, type-safe, and chainable.

---

### Pattern 4: Type Hints (Annotations for Mypy)

**❌ WRONG: Missing type hints (runtime errors, poor IDE support)**
```python
def get_user(user_id):  # ⚠️ No type hints
    return db.query(user_id)

result = get_user("123")  # ⚠️ Should be int, caught only at runtime
```

**✅ CORRECT: Add type hints for function signatures**
```python
from typing import Optional

def get_user(user_id: int) -> Optional[User]:  # ✅ Clear contract
    return db.query(user_id)

result = get_user("123")  # ✅ Mypy catches: Argument 1 has incompatible type "str"
```

**❌ WRONG: Using Dict[str, Any] for structured data (loses type safety)**
```python
def create_user(data: Dict[str, Any]) -> User:
    return User(
        id=data["id"],       # ⚠️ No type checking
        name=data["name"],   # ⚠️ Typo not caught
        emial=data["email"]  # ⚠️ Typo not caught!
    )
```

**✅ CORRECT: Use TypedDict or dataclass for structured data**
```python
from typing import TypedDict

class UserData(TypedDict):
    id: int
    name: str
    email: str

def create_user(data: UserData) -> User:  # ✅ Mypy validates keys
    return User(
        id=data["id"],
        name=data["name"],
        emial=data["email"]  # ✅ Mypy error: Key 'emial' not found in TypedDict
    )
```

**Rule:** Use type hints for all public functions. Prefer `TypedDict` or `@dataclass` over `Dict[str, Any]`. Enable mypy in CI.

---

### Pattern 5: Async/Await (I/O-Bound Operations)

**❌ WRONG: Synchronous I/O in async function (blocks event loop)**
```python
async def fetch_users(user_ids: List[int]) -> List[User]:
    users = []
    for uid in user_ids:
        response = requests.get(f"/api/users/{uid}")  # ⚠️ Blocking!
        users.append(response.json())
    return users  # ⚠️ Sequential, no concurrency
```

**✅ CORRECT: Use async libraries with asyncio.gather (concurrent)**
```python
import asyncio
import httpx

async def fetch_user(user_id: int) -> User:
    async with httpx.AsyncClient() as client:
        response = await client.get(f"/api/users/{user_id}")  # ✅ Non-blocking
        response.raise_for_status()
        return User(**response.json())

async def fetch_users(user_ids: List[int]) -> List[User]:
    tasks = [fetch_user(uid) for uid in user_ids]
    return await asyncio.gather(*tasks)  # ✅ Concurrent execution
```

**Rule:** Use `async`/`await` for I/O-bound operations. Use `asyncio.gather()` for concurrency. Never mix `requests` (sync) with async functions.

---

### Pattern 6: Dataclasses vs Manual __init__ (Boilerplate Reduction)

**❌ WRONG: Manual __init__ with boilerplate (verbose, error-prone)**
```python
class User:
    def __init__(self, id: int, name: str, email: str, is_active: bool = True):
        self.id = id
        self.name = name
        self.email = email
        self.is_active = is_active  # ⚠️ Repetitive, easy to mistype
    
    def __repr__(self):
        return f"User(id={self.id}, name={self.name}, email={self.email})"
```

**✅ CORRECT: Use dataclass (auto-generates __init__, __repr__, __eq__)**
```python
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    email: str
    is_active: bool = True  # ✅ Default values supported
    
    def __post_init__(self) -> None:  # ✅ Custom validation
        if '@' not in self.email:
            raise ValueError("Invalid email")
```

**Rule:** Use `@dataclass` for data-holding classes. Auto-generates `__init__`, `__repr__`, `__eq__`. Add `__post_init__()` for validation.

---

### Pattern 7: Logging with Structured Context (Not String Formatting)

**❌ WRONG: String formatting in log messages (breaks structured logging)**
```python
import logging

logger = logging.getLogger(__name__)

def process_order(order_id: str):
    logger.info(f"Processing order {order_id}")  # ⚠️ Hard to parse/filter
    try:
        # Process order
        logger.info(f"Order {order_id} completed successfully")
    except Exception as e:
        logger.error(f"Failed to process order {order_id}: {str(e)}")
```

**✅ CORRECT: Use extra parameter for structured data (JSON-parseable)**
```python
import logging

logger = logging.getLogger(__name__)

def process_order(order_id: str, correlation_id: str):
    logger.info("Processing order", extra={
        "order_id": order_id,
        "correlation_id": correlation_id  # ✅ Structured, filterable
    })
    try:
        # Process order
        logger.info("Order completed", extra={
            "order_id": order_id,
            "correlation_id": correlation_id
        })
    except Exception as e:
        logger.error("Order processing failed", extra={
            "order_id": order_id,
            "correlation_id": correlation_id,
            "error_type": type(e).__name__
        }, exc_info=True)  # ✅ Include full traceback
```

**Rule:** Use `extra={}` parameter for structured logging. Avoid f-strings in log messages. Use `exc_info=True` for exception tracebacks.

---

### Pattern 8: Import Organization (isort & PEP 8)

**❌ WRONG: Random import order (hard to read, conflicts)**
```python
from .models import User
import os
from typing import List
import asyncio
from pathlib import Path
import sys
```

**✅ CORRECT: Group imports (stdlib → 3rd party → local)**
```python
# Standard library imports
import asyncio
import os
import sys
from pathlib import Path
from typing import List

# Third-party imports
import httpx
import pytest

# Local application imports
from .models import User
from .services import UserService
```

**Configuration:**
```toml
# pyproject.toml
[tool.isort]
profile = "black"
line_length = 88
```

**Rule:** Use `isort` with `profile = "black"`. Order: stdlib → third-party → local. Enforce in CI/pre-commit hooks.

---

## Tools and Linting

Use these tools for code quality:

- **Black**: Code formatting (enforces 88-char line length)
- **Ruff**: Fast linting (replaces Flake8, supports 100+ rules)
- **isort**: Import sorting (use `profile = "black"`)
- **mypy**: Type checking (enable `strict = true` for new projects)
- **pytest**: Testing (with `pytest-cov` for coverage >= 90%)
- **bandit**: Static security scanning (detects unsafe patterns)
- **pre-commit**: Local hooks for black/isort/mypy/ruff (runs before commit)

```toml
# pyproject.toml
[tool.black]
line-length = 88

[tool.isort]
profile = "black"

[tool.mypy]
strict = true
disallow_untyped_defs = true
warn_return_any = true

[tool.ruff]
line-length = 88
select = ["E", "F", "I", "UP", "B", "C4", "SIM"]  # ✅ Enable multiple rule sets
ignore = ["B008"]  # FastAPI dependency pattern

[tool.pytest.ini_options]
addopts = "--cov=src --cov-report=term-missing --cov-fail-under=90"
```

- Run lint/type/test tools in CI (GitHub Actions or Azure DevOps templates).
- Fail builds on linting errors, type errors, or coverage below threshold.

## Governance and MCP Integration

- Refresh MCP guidance before merging to `main` to avoid drift.
- For CI, consider version-locking MCP recommendations in a checked-in artifact to keep reviews stable.

## References

- [PEP 8 Style Guide](https://pep8.org/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [ISE Python Checklist](https://microsoft.github.io/code-with-engineering-playbook/code-reviews/recipes/python/)
- [Python Documentation](https://docs.python.org/3/)
