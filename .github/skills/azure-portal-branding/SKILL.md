---
name: azure-portal-branding
description: >-
  Build Azure-native frontend experiences using Fluent UI React v9 with portal
  patterns and accessibility-first design (WCAG 2.2 AA).
  USE FOR: create portal-style UI, define Fluent UI v9 design tokens, implement
  blade/drawer navigation, build DataGrid with filtering/sorting/selection, build
  command bar with contextual actions, implement breadcrumb navigation, audit
  branding compliance, build accessible forms, implement dark/light/high-contrast theming.
license: MIT
---

# Azure Portal Branding

> **MANDATORY**: All frontend code generated under this skill MUST use Fluent UI React v9 (`@fluentui/react-components`) exclusively. Do NOT use Fluent UI v8 (`@fluentui/react`) for new components. The goal is a frontend that looks and behaves like it was built by Azure engineering — native Fluent capability, Azure portal interaction patterns, and accessibility baked in from day one.

> [!IMPORTANT]
> Enforce Fluent v9-only imports in CI. Add to your lint configuration or pre-commit hook:
>
> ```bash
> # Detect Fluent v8 imports (fail build)
> grep -r "@fluentui/react\"" src/ --include="*.ts" --include="*.tsx" | grep -v "react-components" && echo "ERROR: Fluent UI v8 detected — use @fluentui/react-components (v9)" && exit 1
> ```

## Overview

This skill defines how to build **next-generation Azure-native frontend experiences** — admin UIs, resource management dashboards, and operational portals that feel indistinguishable from first-party Azure portal surfaces.

It enforces:

- **Fluent UI React v9** as the sole component library (no v8, no custom design systems)
- **Azure portal interaction paradigms**: blades/drawers, resource grids, command bars, breadcrumb navigation, filter surfaces
- **Accessibility-first development** (WCAG 2.2 AA minimum, targeting AAA where feasible)
- **Theme-aware, token-driven styling** with full dark mode and high-contrast support
- **Data-dense, performance-optimized layouts** for operational/monitoring use cases
- **Production baseline primitives**: global error boundaries, toast notifications, and keyboard skip links

---

## Capabilities

| Capability                     | Action                                        | Description                                                                                                          |
| ------------------------------ | --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| **Apply Branding Patterns**    | [apply-branding](actions/apply-branding.md)   | Create or refactor a frontend to use shared design tokens, shell, blade navigation, filters, and data-grid patterns. |
| **Review Branding Compliance** | [review-branding](actions/review-branding.md) | Audit an existing frontend for token drift, inconsistent UX patterns, and accessibility/performance gaps.            |

---

## Standards

| Standard               | File                                               | Description                                                                                     |
| ---------------------- | -------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| **Design Tokens**      | [design-tokens.md](standards/design-tokens.md)     | Token architecture for color, typography, spacing, radius, shadows, and semantic status states. |
| **Portal UX Patterns** | [portal-patterns.md](standards/portal-patterns.md) | Blade/drawer behavior, command bar structure, filters, DataGrid conventions, and navigation.    |
| **Checklist**          | [checklist.md](standards/checklist.md)             | Go/no-go validation before shipping branded UI changes.                                         |

---

## Fluent UI v9 Component Mapping (Azure Portal Equivalents)

Map Azure portal surfaces to their Fluent UI v9 component implementations:

| Azure Portal Surface  | Fluent UI v9 Component                                                      | Notes                                                                                    |
| --------------------- | --------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Resource list / grid  | `DataGrid`, `DataGridHeader`, `DataGridBody`, `DataGridRow`, `DataGridCell` | Use `useTableFeatures` with `useTableSort`, `useTableSelection` for sorting/multi-select |
| Detail blade / panel  | `OverlayDrawer` (overlay) or `InlineDrawer` (side-by-side)                  | Right-anchored, `size="medium"` or `size="large"`; preserve parent context               |
| Command bar           | `Toolbar` with `ToolbarButton`, `ToolbarDivider`, `Menu`                    | Primary actions left, overflow via `MenuTrigger` + `MenuPopover`                         |
| Breadcrumb navigation | `Breadcrumb`, `BreadcrumbItem`, `BreadcrumbButton`                          | Resource hierarchy: Subscription → Resource Group → Resource → Sub-resource              |
| Filter bar            | `SearchBox` + `Dropdown` / `Combobox` + `TagGroup` for active filters       | Debounce search (300ms), cancel in-flight, sync with URL                                 |
| Property pane         | `Accordion` with `AccordionItem` sections                                   | Collapsible property groups like Azure resource blades                                   |
| Notifications         | `MessageBar` (inline) + `Toast` / `Toaster` (transient)                     | Use `intent`: `success`, `warning`, `error`, `info`                                      |
| Forms / create flows  | `Field` + `Input` / `Select` / `Combobox` + `Dialog` for confirmation       | Focus management via `useRef`, validation with `validationMessage`                       |
| Status indicators     | `Badge`, `PresenceBadge`, `Spinner`, `ProgressBar`                          | Map resource states: Running, Stopped, Failed, Creating                                  |
| Navigation menu       | `NavDrawer`, `NavItem`, `NavCategory`, `NavSubItem`                         | Left rail navigation matching Azure portal service menu                                  |
| Tab views             | `TabList`, `Tab`                                                            | Resource detail sub-views (Overview, Properties, Monitoring, etc.)                       |

---

## Accessibility-First Development (WCAG 2.2 AA Baseline)

Every component and page MUST meet these requirements:

### Keyboard Navigation

- **Arrow key navigation** within grids and lists: use `useArrowNavigationGroup({ axis: 'grid' })` on `DataGrid`
- **Tab stops** are logical: shell → nav → breadcrumb → command bar → content → blade
- **Escape** closes active drawers/dialogs, restores focus to trigger element
- **Enter/Space** activates buttons, selects rows, opens menus

### Focus Management

- Drawers/dialogs receive focus on open (`autoFocus` or programmatic `ref.focus()`)
- Focus returns to trigger element after drawer/dialog close
- Focus trapping within modal dialogs (built into Fluent `Dialog`)
- Visible focus indicators on all interactive elements (Fluent v9 default styles)

### Screen Reader Support

- Use `aria-label` on icon-only buttons and interactive elements
- Use `aria-live="polite"` regions for dynamic content updates (filter results count, toast notifications)
- DataGrid: use `role="grid"` with `aria-label` describing the data set
- Breadcrumb: use `nav` landmark with `aria-label="Breadcrumb"`
- Command bar: use `role="toolbar"` with `aria-label="Actions"`
- Status changes announced: use `role="status"` or `aria-live` for badge/progress changes

### Color and Contrast

- NEVER use color as the sole indicator of state — always pair with icon/text/pattern
- Use Fluent design tokens (not hardcoded colors) to guarantee contrast ratios
- Support `webLightTheme` and `webDarkTheme` as the baseline runtime themes
- Prefer Windows High Contrast support via `forced-colors: active`; only add explicit high-contrast themes when product requirements mandate them

### Responsive Design

- Minimum viewport: 320px width (WCAG 1.4.10 Reflow)
- Touch targets: minimum 44x44px
- Drawer switches from `InlineDrawer` to `OverlayDrawer` at breakpoint (e.g., < 768px)
- DataGrid column visibility adapts: hide lower-priority columns at narrow widths

---

## Theme Architecture

```typescript
// App root — single FluentProvider wrapping entire app
import {
  FluentProvider,
  webLightTheme,
  webDarkTheme,
  tokens
} from "@fluentui/react-components";

// Theme selection based on user preference + system setting
<FluentProvider theme={resolvedTheme}>
  <App />
</FluentProvider>
```

### Token Usage Rules

1. **Always use `tokens.*`** from `@fluentui/react-components` for colors, spacing, typography — never raw CSS values
2. **Semantic tokens over primitives**: `tokens.colorStatusDangerBackground1` not `tokens.colorPaletteRedBackground1`
3. **Custom tokens** for app-specific branding extend the Fluent theme, they don't replace it
4. **Status colors** map to Fluent intents: `success`, `warning`, `error`, `info` — no custom status palettes

### Production Baseline Requirements

Every branded frontend should include these primitives before feature work is considered complete:

- **Global app shell** with stable top bar + left rail navigation
- **Global error boundary** for runtime failure isolation
- **Global toast provider** for transient success/error feedback
- **Skip link** (`Skip to main content`) for keyboard accessibility
- **Runtime theme switching** (light/dark/high-contrast) with persisted user preference and system fallback
- **Runtime theme switching** (light/dark) with persisted user preference and system fallback
- **Windows High Contrast compatibility** validated via `forced-colors: active` behavior

### Theme Migration Guidance (v8 to v9)

- Preferred end state is **pure Fluent v9 tokens/themes** from `@fluentui/react-components`
- Transitional projects may map legacy v8 Azure palette values into v9 `BrandVariants` during migration
- Migration bridges are temporary: remove v8 theme dependencies when equivalent v9 token coverage is in place
- New UI components must never consume v8 component APIs

---

## DataGrid Pattern (Resource List)

The core interaction pattern for any resource management UI:

```typescript
import {
  DataGrid, DataGridHeader, DataGridBody, DataGridRow,
  DataGridHeaderCell, DataGridCell, TableColumnDefinition,
  createTableColumn, TableCellLayout, useArrowNavigationGroup,
  SearchBox, Toolbar, ToolbarButton, Menu, MenuTrigger,
  MenuPopover, MenuList, MenuItem, Badge
} from "@fluentui/react-components";

// 1. Define strongly-typed columns with sort comparators
const columns: TableColumnDefinition<Resource>[] = [
  createTableColumn<Resource>({
    columnId: "name",
    compare: (a, b) => a.name.localeCompare(b.name),
    renderHeaderCell: () => "Name",
    renderCell: (item) => (
      <TableCellLayout media={<ResourceIcon type={item.type} />}>
        {item.name}
      </TableCellLayout>
    ),
  }),
  createTableColumn<Resource>({
    columnId: "status",
    compare: (a, b) => a.status.localeCompare(b.status),
    renderHeaderCell: () => "Status",
    renderCell: (item) => (
      <Badge appearance="filled" color={statusColor(item.status)}>
        {item.status}
      </Badge>
    ),
  }),
  // ... additional columns
];

// 2. Render with selection + sorting + keyboard navigation
<DataGrid
  items={filteredItems}
  columns={columns}
  sortable
  selectionMode="multiselect"
  getRowId={(item) => item.id}
  aria-label="Resources"
>
  <DataGridHeader>
    <DataGridRow selectionCell={{ "aria-label": "Select all rows" }}>
      {({ renderHeaderCell }) => (
        <DataGridHeaderCell>{renderHeaderCell()}</DataGridHeaderCell>
      )}
    </DataGridRow>
  </DataGridHeader>
  <DataGridBody<Resource>>
    {({ item, rowId }) => (
      <DataGridRow<Resource> key={rowId} selectionCell={{ "aria-label": "Select row" }}>
        {({ renderCell }) => (
          <DataGridCell>{renderCell(item)}</DataGridCell>
        )}
      </DataGridRow>
    )}
  </DataGridBody>
</DataGrid>
```

---

## Drawer/Blade Pattern (Detail View)

```typescript
import { OverlayDrawer, DrawerHeader, DrawerHeaderTitle, DrawerBody, Button } from "@fluentui/react-components";
import { DismissRegular } from "@fluentui/react-icons";

// Right-anchored detail drawer — preserves parent list context
<OverlayDrawer
  open={isOpen}
  onOpenChange={(_, data) => setIsOpen(data.open)}
  position="end"
  size="medium"
>
  <DrawerHeader>
    <DrawerHeaderTitle
      action={
        <Button
          appearance="subtle"
          aria-label="Close"
          icon={<DismissRegular />}
          onClick={() => setIsOpen(false)}
        />
      }
    >
      {selectedResource.name}
    </DrawerHeaderTitle>
  </DrawerHeader>
  <DrawerBody>
    <TabList selectedValue={activeTab} onTabSelect={(_, data) => setActiveTab(data.value)}>
      <Tab value="overview">Overview</Tab>
      <Tab value="properties">Properties</Tab>
      <Tab value="monitoring">Monitoring</Tab>
    </TabList>
    {/* Tab content renders here */}
  </DrawerBody>
</OverlayDrawer>
```

---

## Principles

1. **Native Azure Feel**: Every surface should look like it belongs in the Azure portal — same patterns, same component library, same interaction models.
2. **Semantic Tokens First**: Keep brand decisions in theme/token files; consume `tokens.*` everywhere, never raw CSS.
3. **Accessibility Is Not Optional**: WCAG 2.2 AA is the floor, not the ceiling. Keyboard, screen reader, high contrast, and reflow support are required from the first commit.
4. **Portal Predictability**: Stable shell + command surfaces + list/detail flows; users should always know where actions live.
5. **State Preservation**: Opening/closing drawers must preserve list context (filters, sort, search, selection, scroll position).
6. **Data-Dense Clarity**: Prefer structured grids/lists with explicit loading/empty/error states and row-level action consistency.
7. **Theme-Aware Always**: Components render correctly in light, dark, and high-contrast modes without additional code.
8. **Portable by Default**: Use generic placeholders and architecture-level guidance so patterns transfer to future projects.

---

## Usage

1. Choose the action:
   - New implementation or refactor: `actions/apply-branding.md`
   - Audit only: `actions/review-branding.md`
2. Load standards from `standards/` before making code changes.
3. Apply patterns with minimal disruption (surgical edits, no broad rewrites unless requested).
4. Validate with `standards/checklist.md` before finalizing.

---

## References

- [Fluent UI React v9 Components](https://react.fluentui.dev/) — Primary component reference
- [Fluent UI Icon Browser (flicon.io)](https://flicon.io/) — Searchable reference for `@fluentui/react-icons` icon names
- [Fluent 2 Design System](https://fluent2.microsoft.design/) — Design principles, tokens, accessibility
- [Fluent UI React v9 Storybook](https://master--628d031b55e942004ac95df1.chromatic.com/) — Interactive component examples
- [Azure Portal Design Patterns](https://learn.microsoft.com/azure/azure-portal/) — Portal UX conventions
- [WCAG 2.2 Guidelines](https://www.w3.org/WAI/WCAG22/quickref/) — Accessibility standards reference
- [TypeScript Instructions (Repo)](../../instructions/typescript.instructions.md)

---

## Related Skills

- [TypeScript React Patterns](../typescript-react-patterns/SKILL.md) — React component and state patterns
- [Azure Defaults](../azure-defaults/SKILL.md) — Azure resource configuration standards
- [SPA Endpoint Configuration](../spa-endpoint-configuration/SKILL.md) — Frontend ↔ API connectivity

---

## Currency and Verification

- **Date checked:** 2026-03-31
- **Compatibility:** Fluent UI React v9 (`@fluentui/react-components`), React 19.x, TypeScript 5.x, WCAG 2.2 AA
- **Sources:** [Fluent UI v9](https://react.fluentui.dev/), [Fluent 2 Design System](https://fluent2.microsoft.design/), [WCAG 2.2](https://www.w3.org/TR/WCAG22/)
- **Verification steps:**
  1. Check Fluent UI version: `npm list @fluentui/react-components` (v9.x required, no v8 imports)
  2. Check for v8 imports: `grep -r "@fluentui/react" src/ | grep -v react-components`
  3. Verify theme switching works across light, dark, and high-contrast modes

### Known Pitfalls

| Area                       | Pitfall                                                                                    | Mitigation                                                                                          |
| -------------------------- | ------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| v8 / v9 coexistence        | Mixing `@fluentui/react` (v8) and `@fluentui/react-components` (v9) causes token conflicts | Use v9 exclusively; grep for v8 imports and migrate progressively                                   |
| Nested `FluentProvider`    | Nesting providers fragments token context; portals (Dialog, Toast) may lose theming        | One `FluentProvider` at app root; ensure portals inherit from root provider                         |
| Custom brand ramp contrast | Custom `BrandVariants` can fail WCAG AA contrast at certain ramp steps                     | Validate all 16 brand ramp steps against WCAG AA contrast ratios before shipping                    |
| High Contrast Mode         | `forced-colors: active` overrides Fluent tokens; custom colors disappear                   | Test with Windows High Contrast; avoid relying solely on color for meaning                          |
| `DataGrid` vs `Table`      | Using `Table` primitives when `DataGrid` suffices adds unnecessary complexity              | Default to `DataGrid`; use `Table` + `useTableFeatures` only for non-standard advanced requirements |
| Drawer state persistence   | Opening/closing drawers loses parent list context (filters, scroll, selection)             | Model drawer state in URL params; preserve parent state in React state/context                      |
