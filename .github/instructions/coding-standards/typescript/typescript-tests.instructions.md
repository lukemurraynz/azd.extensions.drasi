---
applyTo: "**/*.test.ts,**/*.test.tsx,**/*.spec.ts,**/*.spec.tsx,**/__tests__/**/*.ts,**/__tests__/**/*.tsx"
description: "TypeScript/React testing best practices for Vitest, React Testing Library, Playwright, accessibility testing, and contract testing"
---

# TypeScript/React Test Instructions

All conventions from [typescript.instructions.md](typescript.instructions.md) apply. This file adds testing-specific guidance for TypeScript/React projects using Vitest, React Testing Library, and Playwright.

**IMPORTANT**: Use the `iseplaybook` MCP server for ISE testing best practices. Use `context7` MCP server (`/websites/vitest_dev`) for Vitest API verification. Use `microsoft.learn.mcp` for Playwright and accessibility testing patterns.

## Test Pyramid and Scope

| Level       | Scope                                       | Speed    | Owns                                           |
| ----------- | ------------------------------------------- | -------- | ---------------------------------------------- |
| Unit        | Pure functions, hooks, utilities, no DOM    | < 10 ms  | Business logic, transformations, validation    |
| Component   | Single component with React Testing Library | < 100 ms | Rendering, user interaction, state transitions |
| Integration | Multiple components, mocked API (MSW)       | < 1 s    | Data flow, error handling, API contracts       |
| E2E         | Full app through Playwright                 | < 30 s   | Critical user journeys, accessibility          |

Unit and component tests run on every save. Integration tests run in CI. E2E tests run on PR merge or nightly.

## Test Project Structure

```text
src/
  components/
    UserCard/
      UserCard.tsx
      UserCard.test.tsx          # Component test
  hooks/
    useLatestRequest.ts
    useLatestRequest.test.ts     # Hook test
  utils/
    validation.ts
    validation.test.ts           # Unit test
  services/
    api.ts
    api.test.ts                  # Integration test (MSW)
tests/
  e2e/
    create-resource.spec.ts      # E2E with Playwright
  setup.ts                       # Global test setup
vitest.config.ts
playwright.config.ts
```

- Co-locate unit and component tests with source files (`.test.ts` / `.test.tsx`)
- Place E2E tests in a separate `tests/e2e/` directory
- Share test utilities via a `tests/setup.ts` or `tests/helpers/` directory

## Test Naming and Structure

Name tests as behavior statements, not method names:

```ts
// ❌ WRONG: "tests the function"
test('validateEmail', () => { ... });

// ✅ CORRECT: describes behavior
test('rejects email without @ symbol', () => { ... });
test('accepts standard email format', () => { ... });
```

Use `describe` blocks to group related tests. Follow Arrange-Act-Assert with visible separation:

```tsx
describe("UserCard", () => {
  test("renders user name and email", () => {
    // Arrange
    const user = { id: "1", name: "Alice", email: "alice@example.com" };

    // Act
    const { getByText } = render(<UserCard user={user} />);

    // Assert
    expect(getByText("Alice")).toBeInTheDocument();
    expect(getByText("alice@example.com")).toBeInTheDocument();
  });

  test("calls onSelect with user id when clicked", async () => {
    const onSelect = vi.fn();
    const user = { id: "42", name: "Bob", email: "bob@example.com" };

    const { getByRole } = render(<UserCard user={user} onSelect={onSelect} />);
    await userEvent.click(getByRole("button"));

    expect(onSelect).toHaveBeenCalledWith("42");
  });
});
```

## Component Testing with React Testing Library

Test components through their public interface (props, user interaction, rendered output), not implementation details:

```tsx
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

test("form shows validation error on empty submit", async () => {
  const onSubmit = vi.fn();
  render(<ResourceForm onSubmit={onSubmit} />);

  await userEvent.click(screen.getByRole("button", { name: /create/i }));

  expect(screen.getByText(/name is required/i)).toBeInTheDocument();
  expect(onSubmit).not.toHaveBeenCalled();
});
```

Rules:

- Query by role, label, or text (what the user sees), not by `data-testid` or class name
- Use `userEvent` over `fireEvent` for realistic user interaction simulation
- Avoid testing internal state or hooks directly; test through rendered output
- Wrap state updates in `act()` only when React Testing Library helpers don't already handle it

## Mocking with Vitest

### Module Mocking

```ts
// Mock an entire module
vi.mock("../services/api", () => ({
  fetchResources: vi.fn(),
}));

// Mock with partial implementation
vi.mock("../services/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../services/api")>();
  return {
    ...actual,
    fetchResources: vi.fn(),
  };
});
```

### Timer Mocking

```ts
beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.restoreAllMocks();
});

test('debounced search fires after 300ms', async () => {
  const onSearch = vi.fn();
  render(<SearchInput onSearch={onSearch} />);

  await userEvent.type(screen.getByRole('searchbox'), 'test');
  expect(onSearch).not.toHaveBeenCalled();

  vi.advanceTimersByTime(300);
  expect(onSearch).toHaveBeenCalledWith('test');
});
```

### Mock Functions

```ts
const handler = vi.fn<(id: string) => Promise<void>>();
handler.mockResolvedValueOnce(undefined);
handler.mockRejectedValueOnce(new Error("Not found"));
```

Rules:

- Always restore mocks in `afterEach` to prevent test pollution
- Prefer `vi.fn()` over manual stubs for call tracking and assertions
- Mock at module boundaries (services, API clients), not internal functions

## API Contract Testing with MSW

Use Mock Service Worker for realistic API mocking that validates request/response contracts:

```ts
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import type { ProblemDetails } from '../types';

const server = setupServer(
  http.get('/api/resources', () => {
    return HttpResponse.json({
      value: [{ id: '1', name: 'Resource 1' }],
    });
  }),

  http.post('/api/resources', async ({ request }) => {
    const body = await request.json();
    if (!body.name) {
      return HttpResponse.json(
        {
          type: 'https://tools.ietf.org/html/rfc9110#section-15.5.1',
          title: 'Bad Request',
          status: 400,
          detail: 'Name is required',
        } satisfies ProblemDetails,
        { status: 400, headers: { 'x-error-code': 'VALIDATION_ERROR' } }
      );
    }
    return HttpResponse.json({ id: '2', name: body.name }, { status: 201 });
  })
);

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('renders resources from API', async () => {
  render(<ResourceList />);
  await expect(screen.findByText('Resource 1')).resolves.toBeInTheDocument();
});

test('shows error message on validation failure', async () => {
  render(<CreateResourceForm />);
  await userEvent.click(screen.getByRole('button', { name: /create/i }));
  await expect(screen.findByText(/name is required/i)).resolves.toBeInTheDocument();
});
```

Rules:

- Use `onUnhandledRequest: 'error'` to catch unmatched API calls
- Test both success and error response paths
- Validate error responses return RFC 9457 Problem Details shape
- Use `satisfies` to type-check mock response bodies against shared types

## Accessibility Testing

### Automated Checks with jest-axe

```tsx
import { axe, toHaveNoViolations } from "jest-axe";

expect.extend(toHaveNoViolations);

test("ResourcePanel has no accessibility violations", async () => {
  const { container } = render(<ResourcePanel resource={mockResource} />);
  const results = await axe(container);
  expect(results).toHaveNoViolations();
});
```

### Keyboard Navigation with Playwright

```ts
test("panel keyboard navigation", async ({ page }) => {
  await page.goto("/dashboard");

  // Open panel with keyboard
  await page.keyboard.press("Tab");
  await page.keyboard.press("Enter");

  // Focus should be trapped in panel
  const focusedElement = await page.evaluate(
    () => document.activeElement?.tagName,
  );
  expect(focusedElement).toBe("BUTTON");

  // Esc closes panel, focus restored
  await page.keyboard.press("Escape");
  const restoredFocus = await page.evaluate(() => document.activeElement?.id);
  expect(restoredFocus).toBe("open-panel-button");
});
```

### Contrast Validation

```tsx
test("error message has sufficient contrast", () => {
  const ratio = getContrastRatio(
    tokens.colorPaletteRedForeground1,
    tokens.colorNeutralBackground1,
  );
  expect(ratio).toBeGreaterThanOrEqual(4.5); // WCAG AA
});
```

Rules:

- Run `axe` checks on every interactive component
- Test keyboard navigation for panels, dialogs, and drawers
- Verify focus trap on open and focus restoration on close
- Validate contrast against WCAG AA (4.5:1 normal text, 3:1 large text)

## Extension Testing

For portal/platform architectures with independent extensions:

### Extension Unit Tests

```tsx
test("validateResourceName enforces rules", () => {
  expect(validateResourceName("ab")).toContain("at least 3 characters");
  expect(validateResourceName("1abc")).toContain("start with a letter");
  expect(validateResourceName("abc")).toHaveLength(0);
});
```

### Extension Integration Tests (Shell Mocks)

```tsx
test("extension handles missing notification service", () => {
  const mockShell: Partial<ShellServices> = {
    telemetry: mockTelemetryService,
    navigation: mockNavigationService,
    // notifications intentionally omitted
  };

  const { getByRole } = render(<MyExtension services={mockShell} />);
  fireEvent.click(getByRole("button", { name: "Submit" }));

  // Extension should not crash; verify fallback behavior
  expect(mockTelemetryService.trackEvent).toHaveBeenCalledWith(
    "ActionCompleted",
  );
});
```

Rules:

- Extension tests must not depend on shell services or other extensions
- Mock shell services via dependency injection
- Test capability discovery: verify graceful degradation when optional capabilities are missing

## E2E Testing with Playwright

```ts
import { test, expect } from "@playwright/test";

test("create resource workflow", async ({ page }) => {
  await page.goto("/resources");

  await page.click('[aria-label="Create resource"]');
  await page.fill('[name="resourceName"]', "Test Resource");
  await page.fill('[name="description"]', "Test description");
  await page.click('button:has-text("Create")');

  await expect(page.locator("text=Resource created")).toBeVisible();
  await expect(page.locator("text=Test Resource")).toBeVisible();
});
```

Rules:

- Test critical user journeys, not every UI path
- Use `aria-label` and role selectors over CSS selectors
- Keep E2E tests fast (< 30s each) and independent
- Run E2E in CI after deploy to a test environment

## Async Testing Patterns

```tsx
// ❌ WRONG: race conditions, no cleanup
test("loads data", () => {
  render(<DataList />);
  expect(screen.getByText("Item 1")).toBeInTheDocument(); // May not be rendered yet
});

// ✅ CORRECT: wait for async rendering
test("loads data", async () => {
  render(<DataList />);
  await expect(screen.findByText("Item 1")).resolves.toBeInTheDocument();
});

// ✅ CORRECT: test loading and error states
test("shows loading then error on failure", async () => {
  server.use(http.get("/api/items", () => HttpResponse.error()));

  render(<DataList />);
  expect(screen.getByText(/loading/i)).toBeInTheDocument();
  await expect(
    screen.findByText(/unable to load/i),
  ).resolves.toBeInTheDocument();
});
```

## Vitest Configuration

```ts
// vitest.config.ts
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./tests/setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      thresholds: { lines: 80, functions: 80, branches: 80 },
    },
    restoreMocks: true,
  },
});
```

```ts
// tests/setup.ts
import "@testing-library/jest-dom/vitest";
```

## Test Anti-Patterns

| Anti-Pattern                   | Why It's Harmful                                | Fix                                                |
| ------------------------------ | ----------------------------------------------- | -------------------------------------------------- |
| Testing implementation details | Breaks on refactors, not on bugs                | Test through rendered output and user events       |
| Snapshot overuse               | Large snapshots are never reviewed              | Use targeted assertions; snapshot small components |
| `any` in test code             | Hides type errors that could reveal real bugs   | Use proper types; `satisfies` for mock data        |
| No `afterEach` cleanup         | Tests pollute each other                        | `vi.restoreAllMocks()` in `afterEach`              |
| `data-testid` for everything   | Creates a parallel API that drifts from a11y    | Query by role, label, or text                      |
| `setTimeout` in tests          | Flaky, slow, timing-dependent                   | Use `vi.useFakeTimers()` or `findBy*` queries      |
| Testing third-party libraries  | Not your code to test                           | Test your integration, not the library             |
| Large `beforeEach` blocks      | Hides test dependencies, slows comprehension    | Arrange in each test; extract helpers for reuse    |
| Ignoring `act()` warnings      | Indicates state updates outside React's control | Wrap updates or use RTL helpers that handle it     |

## Final Self-Check (Before Proposing Test Changes)

✅ Tests describe behavior, not method names
✅ Queries use role, label, or text (not `data-testid`)
✅ `userEvent` used over `fireEvent`
✅ Mocks restored in `afterEach`
✅ MSW configured with `onUnhandledRequest: 'error'`
✅ Error responses assert RFC 9457 Problem Details shape
✅ Accessibility checks run on interactive components
✅ Keyboard navigation tested for panels/dialogs
✅ No `any` in test code
✅ Async tests use `findBy*` or `waitFor`, not `setTimeout`
✅ E2E tests use aria/role selectors
✅ Coverage thresholds configured in vitest.config.ts
✅ Extension tests verify graceful degradation for missing capabilities
