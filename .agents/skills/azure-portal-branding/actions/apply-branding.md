# Apply Branding Patterns

## Purpose

Create or refactor frontend UI to match Azure-portal-native interaction patterns using Fluent UI React v9: `FluentProvider` theme setup, `NavDrawer` shell, `Toolbar` command bars, `OverlayDrawer`/`InlineDrawer` detail blades, `DataGrid` resource lists, and accessible filter surfaces.

---

## Flow

### Step 1: Define Scope and Baseline

Identify which views and components are in scope:

- `FluentProvider` setup (theme module)
- App shell/layout (header, `NavDrawer`, content area)
- Theme/token source module
- At least one list/detail workflow

Map project files to these roles:

- **Theme provider / token entrypoint** (e.g., `src/theme.ts`)
- **Shared layout shell** (e.g., `src/components/Layout.tsx`)
- **Primary list/detail page** (e.g., `src/pages/ResourcesPage.tsx`)

**Success Criteria:**

- [ ] In-scope pages/components listed
- [ ] Token source and shell source identified
- [ ] One representative list/detail flow selected
- [ ] Only `@fluentui/react-components` (v9) imports confirmed

---

### Step 2: Establish Token Contract and FluentProvider

Apply `../standards/design-tokens.md`.

Implement or align:

1. **FluentProvider** wrapping entire app at root with theme prop
2. **Theme switching** — support `webLightTheme`, `webDarkTheme`, `teamsHighContrastTheme`
3. **Brand tokens** — `tokens.colorBrandBackground`, hover/pressed/selected states
4. **Neutral tokens** — `tokens.colorNeutralBackground1/2/3`, foreground hierarchy
5. **Status tokens** — `tokens.colorStatusSuccess*`, `Warning*`, `Danger*` (not custom palettes)
6. **Typography tokens** — `tokens.fontSizeBase*`, `fontWeightSemibold`, `fontFamilyBase`
7. **Spacing/radius/shadow tokens** — `tokens.spacingHorizontal*`, `borderRadiusMedium`, `shadow*`
8. **Motion tokens** — `tokens.durationNormal`, `curveEasyEase` with `prefers-reduced-motion` respect

Rules:

- ZERO hex/rgb literals in component files — `tokens.*` exclusively
- Use `makeStyles` from `@fluentui/react-components` for all custom styling
- Custom brand ramp (if needed) via `createLightTheme`/`createDarkTheme` with `BrandVariants`

**Success Criteria:**

- [ ] `FluentProvider` at app root with theme switching
- [ ] Token categories exist and are named consistently
- [ ] All components consume `tokens.*`, not raw values
- [ ] Theme switching verified: light → dark → high-contrast

---

### Step 3: Standardize Shell, Navigation, and Command Bar

Apply `../standards/portal-patterns.md` (shell + navigation + command bar sections).

Implement:

- **App shell**: `FluentProvider` → App header → `NavDrawer` (left rail) → Content area
- **Left navigation**: `NavDrawer` with `NavItem`, `NavCategory`, `NavSubItem`
- **Breadcrumb**: `Breadcrumb` → `BreadcrumbItem` → `BreadcrumbButton` reflecting resource hierarchy
- **Command bar**: `Toolbar` with `ToolbarButton` for primary actions, `Menu` for overflow
- **Primary action**: `appearance="primary"` at leftmost position in `Toolbar`
- **Icon-only buttons**: MUST have `aria-label`

Avoid:

- Floating action buttons for core admin commands
- Multiple disconnected primary-action areas on same page
- Custom navigation components when `NavDrawer` fits the need

**Success Criteria:**

- [ ] Shell consistent across pages with `NavDrawer` and `Breadcrumb`
- [ ] Commands grouped in one `Toolbar` per page
- [ ] `aria-label` on `NavDrawer`, `Breadcrumb`, `Toolbar`, all icon-only buttons

---

### Step 4: Implement Drawer (Blade) Behavior

Apply `../standards/portal-patterns.md` (drawers section).

Implement:

- **`OverlayDrawer`** for detail/edit flows — right-anchored (`position="end"`), `size="medium"` or `size="large"`
- **`InlineDrawer`** for side-by-side views where parent interaction should remain active
- **Responsive switch**: `InlineDrawer` → `OverlayDrawer` at narrow viewports (< 768px)
- **Structure**: `DrawerHeader` → `DrawerHeaderTitle` (with close button) → `DrawerBody`
- **Tabbed content**: `TabList` + `Tab` inside `DrawerBody` for Overview/Properties/Monitoring views
- URL or route state represents open drawer target
- Browser Back closes drawer before leaving parent view
- Parent list context remains intact (filters/sort/search/selection/scroll)

Accessibility:

- Drawer receives focus on open (Fluent v9 default)
- Close button has `aria-label="Close"`
- `Escape` closes drawer (or prompts if unsaved changes)
- Focus returns to triggering element on close
- `OverlayDrawer` traps focus automatically

**Success Criteria:**

- [ ] Drawer uses Fluent v9 `OverlayDrawer`/`InlineDrawer` (not custom modals)
- [ ] Parent context preserved on open/close
- [ ] Keyboard and focus behavior works correctly
- [ ] Responsive breakpoint tested

---

### Step 5: Standardize Filters and DataGrid

Apply `../standards/portal-patterns.md` (filters + DataGrid sections).

Filter bar:

- `SearchBox` with 300ms debounce + `AbortController` for in-flight cancellation
- `Dropdown`/`Combobox` for categorical filters (Status, Region, Type)
- `TagGroup` with dismissible `Tag` elements for active filters
- Filter state synced with URL query params
- "Clear filters" action visible
- Result count announced via `aria-live="polite"`

DataGrid:

- Use `DataGrid` with `createTableColumn`, `DataGridHeader`, `DataGridBody`, `DataGridRow`, `DataGridCell`
- `sortable` prop enabled, columns defined with `compare` functions
- `selectionMode="multiselect"` with `aria-label` on selection cells
- Row-level actions in dedicated column with `aria-label` on icon buttons
- Bulk actions in command bar, enabled/disabled by selection count
- Loading: `Spinner` or skeleton rows
- Empty: illustration + message + primary create action
- Error: message + retry action

Performance:

- Server-side paging or virtualization for > 100 rows
- Page size selector: 20, 50, 100

**Success Criteria:**

- [ ] Filter/search/sort model centralized with URL sync
- [ ] `DataGrid` with sorting, selection, arrow key navigation
- [ ] Loading/empty/error states explicit
- [ ] Row-level and bulk actions consistent with `aria-label`

---

### Step 6: Validate Accessibility and Theme Support

Accessibility validation:

- [ ] Manual keyboard navigation tested end-to-end (shell → nav → breadcrumb → toolbar → content → drawer)
- [ ] Screen reader tested on at least one primary flow (NVDA or VoiceOver)
- [ ] All ARIA landmarks present: `nav`, `main`, `search`, `toolbar`
- [ ] Color contrast verified in all three themes
- [ ] Windows High Contrast Mode tested
- [ ] Touch targets ≥ 44x44px
- [ ] `prefers-reduced-motion: reduce` disables animations

Theme validation:

- [ ] Light → dark → high-contrast → light switch tested
- [ ] No hardcoded colors break on theme switch
- [ ] User theme preference persists across sessions

---

### Step 7: Final Validation

Run `../standards/checklist.md` and capture unresolved gaps.

Validate with project commands:

- `npm run lint` (including `eslint-plugin-jsx-a11y`)
- `npm run test`
- `npm run build`

**Success Criteria:**

- [ ] Full checklist completed
- [ ] Validation commands pass (or failures documented)
- [ ] Remaining tradeoffs captured for follow-up
- [ ] Zero v8 imports remaining
