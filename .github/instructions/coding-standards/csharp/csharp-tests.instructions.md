---
applyTo: "**/*.Tests/**/*.cs,**/tests/**/*.cs,**/*.Tests.csproj,**/tests/**/*.csproj"
description: "C# and .NET testing best practices for MSTest v4, integration testing, and deterministic test design"
---

# C# Test Instructions

All conventions from [csharp.instructions.md](csharp.instructions.md) apply. This file adds testing-specific guidance for .NET 10 / MSTest v4 projects.

**IMPORTANT**: Use the `iseplaybook` MCP server for ISE testing best practices. Use `context7` MCP server (`/microsoft/testfx`) for MSTest API verification. Use `microsoft.learn.mcp` for ASP.NET Core integration testing patterns.

## Test Pyramid and Scope

Follow the ISE test pyramid: mostly unit tests, fewer integration tests, fewest E2E tests.

| Level       | Scope                                                   | Speed   | Owns                                    |
| ----------- | ------------------------------------------------------- | ------- | --------------------------------------- |
| Unit        | Single class/method, no I/O                             | < 10 ms | Domain logic, calculations, validations |
| Integration | Multiple components, real or containerized dependencies | < 5 s   | Data access, HTTP pipelines, middleware |
| E2E         | Full system through public API or UI                    | < 30 s  | Critical user journeys, smoke tests     |

Unit tests run in the developer inner loop (every save/build). Integration and E2E tests run in CI. Gate PRs on unit + integration; run E2E on merge to main.

## Test Project Structure

```text
src/
  MyService.Domain/
  MyService.Application/
  MyService.Infrastructure/
  MyService.Api/
tests/
  MyService.Domain.Tests/
  MyService.Application.Tests/
  MyService.Infrastructure.Tests/     # integration tests
  MyService.Api.Tests/                # integration tests (WebApplicationFactory)
```

- Name test projects `[ProjectName].Tests`
- One test project per production project
- Domain and Application test projects should have zero infrastructure dependencies

## Test Class and Method Conventions

- Mark test classes `sealed` (no accidental inheritance, enables JIT optimizations)
- Name tests: `MethodName_Scenario_ExpectedBehavior` (e.g., `CalculateTotal_WithDiscount_ReturnsReducedPrice`)
- Follow Arrange-Act-Assert (AAA) with visible separation
- Prefer one assertion per test; related assertions validating the same behavior are acceptable

```csharp
[TestClass]
public sealed class OrderTests
{
    [TestMethod]
    public void AddItem_WithValidProduct_IncreasesTotal()
    {
        // Arrange
        var order = new Order();
        var product = new Product("Widget", price: 10.00m);

        // Act
        order.AddItem(product, quantity: 3);

        // Assert
        Assert.AreEqual(30.00m, order.Total);
    }
}
```

## Assertion Best Practices

Use specialized assertions. The `MSTEST0037` analyzer enforces this.

| Instead of                           | Use                                     |
| ------------------------------------ | --------------------------------------- |
| `Assert.IsTrue(x != null)`           | `Assert.IsNotNull(x)`                   |
| `Assert.IsTrue(a == b)`              | `Assert.AreEqual(expected, actual)`     |
| `Assert.IsTrue(list.Contains(item))` | `CollectionAssert.Contains(list, item)` |
| `Assert.IsTrue(list.Count == 0)`     | `Assert.AreEqual(0, list.Count)`        |
| `[ExpectedException]`                | `Assert.ThrowsExactly<T>(() => ...)`    |

Put `expected` first in `Assert.AreEqual`; failure messages depend on parameter order.

Use `Assert.ThrowsExactly<T>` for exact type matching or `Assert.Throws<T>` when derived exceptions are acceptable. `[ExpectedException]` is removed in MSTest v4.

## Test Lifecycle and Initialization

```csharp
[TestClass]
public sealed class ServiceTests
{
    private readonly IClock _clock;
    private readonly MyService _sut;

    // Constructor injection for TestContext (MSTest 3.6+)
    public ServiceTests(TestContext testContext)
    {
        _clock = new FakeClock(new DateTimeOffset(2025, 1, 15, 0, 0, 0, TimeSpan.Zero));
        _sut = new MyService(_clock);
        TestContext = testContext;
    }

    public TestContext TestContext { get; }

    [TestMethod]
    [Timeout(5000)]
    public async Task ProcessAsync_WithValidInput_Completes()
    {
        // Pass CancellationToken from TestContext; CancellationToken.None defeats timeout enforcement
        var result = await _sut.ProcessAsync("input", TestContext.CancellationToken);
        Assert.IsTrue(result.IsSuccess);
    }
}
```

Rules:

- Initialize sync fields in the constructor (enables `readonly`, satisfies nullable analyzers)
- Use `[TestInitialize]` only for async setup
- Inject `TestContext` via constructor (MSTest 3.6+), not the legacy nullable property pattern
- Pass `TestContext.CancellationToken` to async calls in tests with `[Timeout]`

## Determinism

Favor in-memory test doubles for non-deterministic dependencies:

| Dependency   | Abstraction       | Test Double                                          |
| ------------ | ----------------- | ---------------------------------------------------- |
| Current time | `IClock`          | `FakeClock` with fixed `DateTimeOffset`              |
| GUIDs / IDs  | `IIdGenerator`    | `SequentialIdGenerator` returning predictable values |
| Randomness   | `IRandomProvider` | Seeded implementation                                |

Never use `DateTime.Now`, `DateTime.UtcNow`, `Guid.NewGuid()`, or `Random.Shared` directly in code under test. Inject abstractions at seams.

## Data-Driven Tests

Use `[DataRow]` for simple inline cases:

```csharp
[TestMethod]
[DataRow("user@example.com", true)]
[DataRow("not-an-email", false)]
[DataRow("", false)]
public void IsValidEmail_WithInput_ReturnsExpected(string input, bool expected)
{
    Assert.AreEqual(expected, EmailValidator.IsValid(input));
}
```

Use `[DynamicData]` with `ValueTuple` return types (MSTest 3.7+) for complex or computed test data:

```csharp
[TestMethod]
[DynamicData(nameof(DiscountCases), DynamicDataSourceType.Method,
    DynamicDataDisplayName = nameof(GetDiscountDisplayName))]
public void CalculateDiscount_WithTier_ReturnsExpected(
    CustomerTier tier, decimal orderTotal, decimal expectedDiscount)
{
    var discount = PricingEngine.CalculateDiscount(tier, orderTotal);
    Assert.AreEqual(expectedDiscount, discount);
}

private static IEnumerable<(CustomerTier, decimal, decimal)> DiscountCases()
{
    yield return (CustomerTier.Silver, 100m, 5m);
    yield return (CustomerTier.Gold, 100m, 10m);
    yield return (CustomerTier.Platinum, 100m, 15m);
}

private static string GetDiscountDisplayName(MethodInfo _, object[] data)
    => $"{data[0]} tier on {data[1]:C} order";
```

Guidance:

- `[DataRow]` for 1-5 inline primitive cases
- `[DynamicData]` for complex types, computed data, or cases generated from external sources
- Always provide `DisplayName` or `DynamicDataDisplayName` so test explorer shows meaningful case names
- Avoid `IEnumerable<object[]>` (loses type safety); prefer `ValueTuple` return from `DynamicData` methods

## Integration Testing with WebApplicationFactory

Use `WebApplicationFactory<TEntryPoint>` for integration tests that exercise the full HTTP pipeline (routing, middleware, filters, serialization):

```csharp
[TestClass]
public sealed class OrdersApiTests
{
    private static WebApplicationFactory<Program> _factory = null!;
    private HttpClient _client = null!;

    [ClassInitialize]
    public static void ClassInit(TestContext _)
    {
        _factory = new WebApplicationFactory<Program>()
            .WithWebHostBuilder(builder =>
            {
                builder.ConfigureTestServices(services =>
                {
                    // Replace real dependencies with test doubles
                    services.RemoveAll<IOrderRepository>();
                    services.AddSingleton<IOrderRepository, FakeOrderRepository>();
                });
            });
    }

    [TestInitialize]
    public void TestInit() => _client = _factory.CreateClient();

    [TestCleanup]
    public void TestCleanup() => _client.Dispose();

    [ClassCleanup]
    public static void ClassCleanup() => _factory.Dispose();

    [TestMethod]
    public async Task GetOrder_WithValidId_ReturnsOk()
    {
        var response = await _client.GetAsync("/api/orders/1");

        Assert.AreEqual(HttpStatusCode.OK, response.StatusCode);
        var order = await response.Content.ReadFromJsonAsync<OrderDto>();
        Assert.IsNotNull(order);
        Assert.AreEqual("1", order.Id);
    }

    [TestMethod]
    public async Task CreateOrder_WithInvalidBody_ReturnsProblemDetails()
    {
        var response = await _client.PostAsJsonAsync("/api/orders", new { });

        Assert.AreEqual(HttpStatusCode.BadRequest, response.StatusCode);
        var problem = await response.Content.ReadFromJsonAsync<ProblemDetails>();
        Assert.IsNotNull(problem);
        Assert.AreEqual("https://tools.ietf.org/html/rfc9110#section-15.5.1", problem.Type);
    }
}
```

Rules:

- Use `ConfigureTestServices` to replace dependencies (runs after `Program.cs` DI setup, so replacements win)
- Test against the HTTP contract (status codes, response shapes, headers), not internal implementation
- Verify error responses return RFC 9457 Problem Details
- Share `WebApplicationFactory` across tests in the same class via `[ClassInitialize]`

### Mock Authentication in Integration Tests

When testing authenticated endpoints, register a fake authentication handler:

```csharp
builder.ConfigureTestServices(services =>
{
    services.AddAuthentication("Test")
        .AddScheme<AuthenticationSchemeOptions, FakeAuthHandler>("Test", _ => { });
});
```

```csharp
internal sealed class FakeAuthHandler(
    IOptionsMonitor<AuthenticationSchemeOptions> options,
    ILoggerFactory logger,
    UrlEncoder encoder)
    : AuthenticationHandler<AuthenticationSchemeOptions>(options, logger, encoder)
{
    protected override Task<AuthenticateResult> HandleAuthenticateAsync()
    {
        var claims = new[] { new Claim(ClaimTypes.Name, "test-user") };
        var identity = new ClaimsIdentity(claims, "Test");
        var principal = new ClaimsPrincipal(identity);
        var ticket = new AuthenticationTicket(principal, "Test");
        return Task.FromResult(AuthenticateResult.Success(ticket));
    }
}
```

### Test Database Strategy

For integration tests that need a relational database:

- **Prefer SQLite in-memory** (`Data Source=:memory:`) for fast, deterministic tests that validate EF Core queries and migrations
- **Avoid EF Core InMemory provider** for integration tests (it does not enforce constraints, relationships, or SQL behavior)
- **Use Testcontainers** when you need the exact production database engine (PostgreSQL, SQL Server) for engine-specific features

```csharp
builder.ConfigureTestServices(services =>
{
    services.RemoveAll<DbContextOptions<AppDbContext>>();
    services.AddDbContext<AppDbContext>(options =>
        options.UseSqlite("Data Source=:memory:"));
});
```

## Concurrency and ETag Tests

Assert strict status codes for optimistic concurrency scenarios:

```csharp
[TestMethod]
public async Task UpdateOrder_WithStaleETag_Returns412()
{
    var getResponse = await _client.GetAsync("/api/orders/1");
    var etag = getResponse.Headers.ETag;

    // Simulate a conflicting update by another client
    await _client.PutAsJsonAsync("/api/orders/1", new UpdateOrderDto { Total = 200m });

    // Attempt update with the now-stale ETag
    var request = new HttpRequestMessage(HttpMethod.Put, "/api/orders/1")
    {
        Content = JsonContent.Create(new UpdateOrderDto { Total = 300m })
    };
    request.Headers.IfMatch.Add(etag!);

    var response = await _client.SendAsync(request);
    Assert.AreEqual(HttpStatusCode.PreconditionFailed, response.StatusCode);
}
```

- Assert `412 Precondition Failed` on ETag mismatches; do not accept permissive multi-status responses
- Test the happy path (matching ETag returns `200`) and conflict path (stale ETag returns `412`)

## Benchmarking

Use BenchmarkDotNet for hot-path performance validation:

```csharp
[MemoryDiagnoser]
public class SerializationBenchmarks
{
    private readonly Order _order = OrderFixtures.CreateLargeOrder();

    [Benchmark(Baseline = true)]
    public string SerializeWithSystemTextJson()
        => JsonSerializer.Serialize(_order);
}
```

- Run benchmarks in Release mode, never in test suites
- Track allocations with `[MemoryDiagnoser]`
- Profile Domain + critical Application hot paths

## Test DTO Signature Stability

Test DTO shapes using named parameters to catch breaking changes at compile time:

```csharp
[TestMethod]
public void OrderDto_SignatureStable()
{
    // Named parameters break compilation if properties are renamed or reordered
    var dto = new OrderDto(
        Id: "1",
        CustomerName: "Test",
        Total: 100m,
        Status: OrderStatus.Pending);

    Assert.AreEqual("1", dto.Id);
}
```

## Test Anti-Patterns

| Anti-Pattern            | Why It's Harmful                                | Fix                                        |
| ----------------------- | ----------------------------------------------- | ------------------------------------------ |
| No assertions           | False confidence; test always passes            | Add meaningful assertions                  |
| Swallowed exceptions    | Hides real failures                             | Let exceptions propagate or assert them    |
| `Thread.Sleep`          | Flaky, slow, non-deterministic                  | Use async waits or test abstractions       |
| `DateTime.Now` directly | Non-deterministic; tests fail across time zones | Inject `IClock`                            |
| `Random` without seed   | Non-reproducible failures                       | Inject `IRandomProvider` with fixed seed   |
| Over-mocking            | More mock setup than test logic                 | Test outcomes, not call sequences          |
| Testing private methods | Couples tests to implementation                 | Test through public API                    |
| Shared mutable state    | Order-dependent test failures                   | Fresh state per test                       |
| Asserting exact logs    | Brittle; log messages change often              | Assert structured log properties if needed |

## Test Platform Configuration

.NET 10 uses Microsoft.Testing.Platform (MTP) as the default test runner. Configure in `global.json` at the repo root:

```json
{
  "sdk": { "version": "10.0.100" },
  "test": { "runner": "Microsoft.Testing.Platform" }
}
```

- MTP v2 on .NET 10 SDK no longer supports VSTest-based `dotnet test` mode
- The legacy `TestingPlatformDotnetTestSupport` MSBuild property is removed
- CI must use `dotnet test` with MTP mode, not `vstest.console`
- `dotnet test` syntax: use `--project`/`--solution` instead of positional arguments
- MSTest.Sdk defaults to MTP mode and does not include `Microsoft.NET.Test.Sdk`
- Do not mix VSTest-based and MTP-based test projects in the same solution

## Final Self-Check (Before Proposing Test Changes)

✅ Test classes sealed
✅ Tests named `MethodName_Scenario_ExpectedBehavior`
✅ Specialized assertions used (not generic `Assert.IsTrue`)
✅ AAA pattern with visible separation
✅ TestContext injected via constructor (not legacy property)
✅ CancellationToken passed in async tests with `[Timeout]`
✅ Time/ID/randomness injected through abstractions
✅ No `Thread.Sleep`, no `DateTime.Now`, no unseeded `Random`
✅ Integration tests use `WebApplicationFactory` with `ConfigureTestServices`
✅ SQLite preferred over EF InMemory for database tests
✅ Error responses assert RFC 9457 Problem Details shape
✅ ETag tests assert strict `412` status codes
✅ `global.json` configures MTP test runner
✅ Data-driven tests use `DisplayName` for case identification
