# Portal UX Patterns Standard

## Purpose

Define Azure-portal-native interaction patterns using Fluent UI React v9 for operational, admin, and resource-management frontends. Every surface should feel like it was built by Azure engineering.

---

## 1) Shell and Navigation

### Layout Structure

- **App Shell**: `FluentProvider` → App header → `NavDrawer` (left rail) → Content area
- Use `NavDrawer`, `NavItem`, `NavCategory`, `NavSubItem` for the left navigation rail
- Header contains: app title/logo, breadcrumb, user identity, settings, notifications

### Breadcrumb Navigation

```tsx
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbButton,
  BreadcrumbDivider,
} from "@fluentui/react-components";

<Breadcrumb aria-label="Breadcrumb">
  <BreadcrumbItem>
    <BreadcrumbButton onClick={() => navigate("/subscriptions")}>
      Subscriptions
    </BreadcrumbButton>
  </BreadcrumbItem>
  <BreadcrumbDivider />
  <BreadcrumbItem>
    <BreadcrumbButton onClick={() => navigate(`/rg/${rgName}`)}>
      {rgName}
    </BreadcrumbButton>
  </BreadcrumbItem>
  <BreadcrumbDivider />
  <BreadcrumbItem>
    <BreadcrumbButton current>{resourceName}</BreadcrumbButton>
  </BreadcrumbItem>
</Breadcrumb>;
```

### Rules

- Navigation must be predictable across all pages — same `NavDrawer` position and items
- Place primary page actions in one `Toolbar` (command bar) — never scatter actions
- Keep command placement consistent across list/detail pages
- Breadcrumb hierarchy reflects resource structure: Subscription → Resource Group → Resource → Sub-resource
- Use `aria-label` on `Breadcrumb` and `NavDrawer` for screen reader landmarks
- For keyboard modality awareness in composite layouts, use `useKeyboardNavAttribute` from `@fluentui/react-tabster`
- For custom focus presentation, use Fluent focus utilities (for example, `createFocusOutlineStyle`) instead of ad-hoc outlines

### Reference Shell Variant (Compact Azure-Style Rail)

For data-dense admin experiences, use a compact icon-first rail with a stable branded top bar:

- 50–56px top bar with product identity and user/session actions
- 56–64px left rail with icon + short label nav items
- Active nav marker with clear selected state and high-contrast-safe visual indicator
- Fixed shell regions (`header`, `aside`, `main`) so keyboard tab order remains predictable
- Include a top-level skip link to jump to `main`

---

## 2) Drawers (Blades / Panels)

Azure portal "blades" map to Fluent UI v9 **Drawers**:

| Scenario                    | Component                        | Behavior                                                     |
| --------------------------- | -------------------------------- | ------------------------------------------------------------ |
| Detail/edit panel (overlay) | `OverlayDrawer`                  | Right-anchored, backdrop, parent visible but not interactive |
| Side-by-side view           | `InlineDrawer`                   | Right-anchored, parent content reflows                       |
| Responsive                  | `InlineDrawer` → `OverlayDrawer` | Switch at breakpoint (e.g., < 768px viewport width)          |

### Structure

```tsx
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
      Resource Detail
    </DrawerHeaderTitle>
  </DrawerHeader>
  <DrawerBody>
    <TabList
      selectedValue={activeTab}
      onTabSelect={(_, data) => setActiveTab(data.value)}
    >
      <Tab value="overview">Overview</Tab>
      <Tab value="properties">Properties</Tab>
      <Tab value="monitoring">Monitoring</Tab>
    </TabList>
    {/* Tab content */}
  </DrawerBody>
</OverlayDrawer>
```

### Context Preservation Rules

- Preserve parent state while drawer is open: filters, search, sort, selection, scroll position
- Model drawer state in route/search params when feasible (deep-linkable and reload-safe)
- Browser Back should close the active drawer before navigating away
- Avoid history spam for frequently changing state (for example, live search/filter typing) by using replace semantics for incremental URL updates

### Multi-Blade Stacking Pattern (Recommended)

When workflows require opening detail from detail, stack non-modal right-anchored blades:

- Use `OverlayDrawer` with `modalType="non-modal"` for stacked blades
- Offset each additional blade horizontally for depth cues (for example, 24–32px)
- Maintain z-index progression per blade level
- Preserve parent blade/list state while child blades are active
- Close only the top blade on `Escape`; keep underlying context intact

### Accessibility Rules

- Drawer receives focus on open (Fluent v9 handles this by default)
- `Escape` closes drawer (or prompts on unsaved changes)
- Focus returns to triggering element after close
- `OverlayDrawer` traps focus automatically
- Use `aria-label` on drawer actions (especially icon-only close button)

---

## 3) Command Bar (Toolbar)

The command bar is the single surface for page-level and contextual actions:

```tsx
import {
  Toolbar,
  ToolbarButton,
  ToolbarDivider,
  Menu,
  MenuTrigger,
  MenuPopover,
  MenuList,
  MenuItem,
} from "@fluentui/react-components";
import {
  AddRegular,
  DeleteRegular,
  ArrowDownloadRegular,
  MoreHorizontalRegular,
} from "@fluentui/react-icons";

<Toolbar aria-label="Resource actions">
  <ToolbarButton appearance="primary" icon={<AddRegular />}>
    Create
  </ToolbarButton>
  <ToolbarButton icon={<DeleteRegular />} disabled={selectedCount === 0}>
    Delete
  </ToolbarButton>
  <ToolbarDivider />
  <ToolbarButton icon={<ArrowDownloadRegular />}>Export</ToolbarButton>
  <Menu>
    <MenuTrigger>
      <ToolbarButton
        aria-label="More actions"
        icon={<MoreHorizontalRegular />}
      />
    </MenuTrigger>
    <MenuPopover>
      <MenuList>
        <MenuItem>Refresh</MenuItem>
        <MenuItem>Settings</MenuItem>
      </MenuList>
    </MenuPopover>
  </Menu>
</Toolbar>;
```

### Rules

- Primary action (Create/Add) uses `appearance="primary"` and sits leftmost
- Destructive actions (Delete) are disabled when no selection, use confirmation `Dialog` before executing
- Overflow actions go into a `Menu` with `MoreHorizontalRegular` icon
- Bulk actions are enabled/disabled based on DataGrid selection count
- Use `aria-label="Resource actions"` on the `Toolbar` element
- Icon-only buttons MUST have `aria-label`

---

## 4) Filters and Search

### Filter Bar Pattern

```tsx
<div
  role="search"
  aria-label="Filter resources"
  style={{ display: "flex", gap: tokens.spacingHorizontalM }}
>
  <SearchBox
    placeholder="Search by name..."
    value={searchText}
    onChange={(_, data) => debouncedSearch(data.value)}
    aria-label="Search resources"
  />
  <Dropdown
    placeholder="Status"
    onOptionSelect={(_, data) => setStatusFilter(data.optionValue)}
  >
    <Option value="Running">Running</Option>
    <Option value="Stopped">Stopped</Option>
    <Option value="Failed">Failed</Option>
  </Dropdown>
  <Dropdown
    placeholder="Region"
    onOptionSelect={(_, data) => setRegionFilter(data.optionValue)}
  >
    {/* Region options */}
  </Dropdown>
  <Button appearance="subtle" onClick={clearFilters}>
    Clear filters
  </Button>
</div>;

{
  /* Active filter tags */
}
<TagGroup
  onDismiss={(_, data) => removeFilter(data.value)}
  aria-label="Active filters"
>
  {activeFilters.map((f) => (
    <Tag key={f.key} value={f.key} dismissible>
      {f.label}
    </Tag>
  ))}
</TagGroup>;

{
  /* Result count announcement */
}
<div aria-live="polite" role="status">
  {`${filteredItems.length} resources found`}
</div>;
```

### Rules

- Keep filter state centralized (single hook or context — never split across unrelated components)
- Debounce text search: 300ms minimum
- Cancel in-flight requests when filter inputs change (`AbortController`)
- Sync filter state with URL query params for shareable deep links
- Provide explicit "Clear filters" action
- Show active filters as dismissible `Tag` elements
- Announce result count changes with `aria-live="polite"`
- `Combobox` in accessible contexts: use `inlinePopup={true}` for VoiceOver compatibility

---

## 5) DataGrid — Resource Lists

### Component Selection Gate

- Use `DataGrid` by default for standard list scenarios (sorting, selection, common table interactions)
- Use low-level `Table` primitives + `useTableFeatures` when non-standard interaction or high customization is required
- If requirements diverge significantly from DataGrid patterns, switch to `Table` intentionally rather than over-customizing DataGrid

### Mandatory Capabilities

Every production resource list MUST implement:

- ✅ Sortable columns (click header to toggle)
- ✅ Multi-select with bulk actions
- ✅ Filter integration (filters above grid)
- ✅ Explicit loading state (`Spinner` overlay or skeleton rows)
- ✅ Empty state (illustration + message + create action)
- ✅ Error state (message + retry action)
- ✅ Arrow key navigation within grid rows/cells

### Implementation with `useTableFeatures`

```tsx
import {
  useTableFeatures,
  useTableSort,
  useTableSelection,
  createTableColumn,
} from "@fluentui/react-components";

const {
  getRows,
  sort: { getSortDirection, toggleColumnSort, sort },
  selection: {
    allRowsSelected,
    someRowsSelected,
    toggleAllRows,
    toggleRow,
    isRowSelected,
  },
} = useTableFeatures({ columns, items }, [
  useTableSort({
    defaultSortState: { sortColumn: "name", sortDirection: "ascending" },
  }),
  useTableSelection({ selectionMode: "multiselect" }),
]);

const rows = sort(getRows());
```

### Row-Level Actions

- Use a dedicated action column (rightmost) with icon buttons or a `Menu`
- Row actions: View, Edit, Delete at minimum
- Icon-only action buttons MUST have `aria-label` (e.g., `aria-label="Delete resource-name"`)

### Bulk Actions

- Surface in command bar when `selectedCount > 0`
- Show selection count: `"{selectedCount} selected"` in toolbar or status bar
- Bulk delete requires confirmation `Dialog`

### Performance Rules

- For large datasets (> 100 rows): implement server-side paging or virtualization
- Show page size selector: 20, 50, 100 rows
- Avoid client-side sorting/filtering of datasets > 1,000 rows

### Migration Rule

- If a page currently uses cards, preserve behavior first; migrate incrementally to grid patterns without breaking workflows

---

## 6) Forms and Create Flows

### Field Pattern

```tsx
import {
  Field,
  Input,
  Select,
  Combobox,
  Dialog,
  DialogTrigger,
  DialogSurface,
  DialogTitle,
  DialogBody,
  DialogActions,
  Button,
} from "@fluentui/react-components";

<Field
  label="Resource name"
  required
  validationMessage={nameError}
  validationState={nameError ? "error" : "none"}
>
  <Input
    value={name}
    onChange={(_, data) => setName(data.value)}
    ref={nameInputRef}
  />
</Field>;
```

### Rules

- Use `Field` wrapper for all form controls (provides label, validation message, required indicator)
- Focus first input on form/dialog mount via `useRef` + `useEffect`
- Destructive confirmations use `Dialog` with explicit cancel/confirm
- Multi-step create flows: use stepped layout with progress indicator
- Validation: inline per-field + summary at submit, never only at submit

---

## 7) Notifications and Feedback

| Type          | Component                        | Use When                                                                    |
| ------------- | -------------------------------- | --------------------------------------------------------------------------- |
| Inline status | `MessageBar`                     | Page-level or section-level persistent messages                             |
| Transient     | `Toast` via `useToastController` | Action completion/failure (auto-dismiss 5s for success, persist for errors) |
| In-progress   | `Spinner` or `ProgressBar`       | Long-running operations                                                     |
| Empty state   | Custom composition               | No data matches filters or first-time experience                            |

### Rules

- Success toasts: auto-dismiss after 5 seconds
- Error toasts: persist until user dismisses
- Error messages: user-readable text + optional diagnostic detail for support
- Use `aria-live="polite"` on `MessageBar` for dynamic status updates
- Loading: disable action buttons and show `Spinner` inline or as overlay
- Empty states: include illustration/icon, descriptive text, and primary "Create" action

---

## 8) Accordion and Property Panes

For resource detail views with collapsible sections (like Azure resource property blades):

```tsx
import {
  Accordion,
  AccordionItem,
  AccordionHeader,
  AccordionPanel,
} from "@fluentui/react-components";

<Accordion multiple collapsible defaultOpenItems={["essentials"]}>
  <AccordionItem value="essentials">
    <AccordionHeader>Essentials</AccordionHeader>
    <AccordionPanel>{/* Key-value property grid */}</AccordionPanel>
  </AccordionItem>
  <AccordionItem value="configuration">
    <AccordionHeader>Configuration</AccordionHeader>
    <AccordionPanel>{/* ... */}</AccordionPanel>
  </AccordionItem>
</Accordion>;
```

---

## Good vs Bad

✅ Good:

- List page with `Toolbar` (command bar), filter bar with `SearchBox` + `Dropdown`, `DataGrid` with sorting/selection, route-driven `OverlayDrawer` detail panel
- `DataGrid` with consistent row actions, `aria-label` on all icon buttons, keyboard-navigable
- `FluentProvider` at app root with theme switching support
- Active filters shown as dismissible `Tag` elements with result count announced via `aria-live`

❌ Bad:

- Multiple scattered primary-action buttons outside a command bar
- Opening detail views that reset list filters/selection/scroll
- Raw CSS color values instead of `tokens.*`
- Icon-only buttons without `aria-label`
- Using Fluent UI v8 components alongside v9 (no mixing)
- Color as the sole indicator of status without icon/text pairing
