# Review Branding Compliance

## Purpose

Audit an existing frontend for Azure-portal-native branding, Fluent UI React v9 compliance, accessibility conformance (WCAG 2.2 AA), and theme support — then produce prioritized remediation guidance.

---

## Flow

### Step 1: Inventory Entry Points

Catalog:

- `FluentProvider` setup and theme module
- Shared layout/shell component (header, `NavDrawer`, content area)
- Top 2-3 highest-traffic list/detail pages
- Import analysis: identify any v8 (`@fluentui/react`) imports

Map project files to these roles:

- **Theme provider / token entrypoint** (e.g., `src/theme.ts`)
- **Shared layout shell** (e.g., `src/components/Layout.tsx`)
- **Primary list/detail page** (e.g., `src/pages/ResourcesPage.tsx`)

**Success Criteria:**

- [ ] Token, shell, and workflow entry points identified
- [ ] Review scope explicitly bounded
- [ ] v8 vs v9 import inventory captured

---

### Step 2: Audit Token Discipline and Theme Support

Apply `../standards/design-tokens.md`.

Check:

- `FluentProvider` wraps entire app at root with theme prop
- Theme switching supports `webLightTheme`, `webDarkTheme`, `teamsHighContrastTheme`
- Zero hex/rgb color literals in component files — all use `tokens.*`
- Typography uses `tokens.fontSizeBase*`, `tokens.fontWeightSemibold` etc.
- Spacing uses `tokens.spacingHorizontal*`, `tokens.spacingVertical*` — no raw px values
- Status styling uses `tokens.colorStatusSuccess*`, `Warning*`, `Danger*`
- Styling uses `makeStyles` from `@fluentui/react-components`
- Custom brand ramp (if used) passes WCAG AA contrast

**Success Criteria:**

- [ ] Token drift hotspots documented (grep for hex/rgb literals)
- [ ] v8 token usage flagged for migration
- [ ] Theme switching tested across all three themes
- [ ] High-impact token fixes identified

---

### Step 3: Audit Shell, Navigation, Commands, and Drawers

Apply `../standards/portal-patterns.md` (shell + navigation + command bar + drawers sections).

Check:

- `NavDrawer` used for left rail navigation with `aria-label`
- `Breadcrumb` reflects resource hierarchy with `aria-label="Breadcrumb"`
- Single `Toolbar` (command bar) per page with primary action at leftmost
- Overflow actions in `Menu` with `MoreHorizontalRegular`
- `OverlayDrawer`/`InlineDrawer` for detail views (not custom modals)
- Drawer close button has `aria-label="Close"`
- Parent context preserved when drawer opens/closes
- Drawer state is route-aware
- Responsive drawer switching at breakpoints

**Success Criteria:**

- [ ] Navigation/command inconsistencies documented
- [ ] Drawer behavior risks and fixes documented
- [ ] Missing `aria-label` instances cataloged

---

### Step 4: Audit Filters and DataGrid

Apply `../standards/portal-patterns.md` (filters + DataGrid sections).

Check:

- `SearchBox` with debounce (300ms) and `AbortController` for in-flight cancellation
- `Dropdown`/`Combobox` for categorical filters
- Active filters shown as dismissible `Tag` elements
- Filter state synced with URL query params
- Result count announced via `aria-live="polite"`
- `DataGrid` with `createTableColumn`, `sortable`, `selectionMode="multiselect"`
- Selection cells have `aria-label`
- Row-level action buttons have `aria-label`
- Loading/empty/error states explicit
- Server-side paging for datasets > 100 rows
- Arrow key navigation functional within grid

**Success Criteria:**

- [ ] Filter/data-flow issues documented
- [ ] Missing accessibility attributes cataloged
- [ ] Performance and UX quick wins identified

---

### Step 5: Audit Accessibility (WCAG 2.2 AA)

Keyboard:

- [ ] Tab order logical: shell → nav → breadcrumb → toolbar → content → drawer
- [ ] `Enter`/`Space` activates buttons/selects rows
- [ ] `Escape` closes drawers/dialogs/menus
- [ ] Arrow keys navigate within DataGrid, menus, tabs
- [ ] No keyboard traps
- [ ] Focus indicators visible on all interactive elements

Screen Readers:

- [ ] Test with NVDA (Windows) or VoiceOver (macOS)
- [ ] ARIA landmarks present: `nav`, `main`, `search`, `toolbar`
- [ ] `aria-label` on: icon-only buttons, DataGrid, Breadcrumb, Toolbar, NavDrawer
- [ ] `aria-live` regions for dynamic content
- [ ] `Combobox` uses `inlinePopup={true}` where VoiceOver needed

Visual:

- [ ] Color never sole indicator of state
- [ ] Contrast ratios meet AA: 4.5:1 normal text, 3:1 large text/UI components
- [ ] Touch targets ≥ 44x44px
- [ ] Viewport support ≥ 320px width
- [ ] `prefers-reduced-motion: reduce` disables animations

High Contrast:

- [ ] Windows High Contrast Mode tested
- [ ] `forced-colors: active` handled where custom styling exists
- [ ] No information lost in high-contrast mode

Responsive:

- [ ] Drawer switches from inline to overlay at narrow viewports
- [ ] DataGrid hides lower-priority columns at narrow widths
- [ ] Shell navigation collapses appropriately

**Success Criteria:**

- [ ] Accessibility risks prioritized by severity
- [ ] Responsive regressions captured

---

### Step 6: Produce Findings Report

Report with four buckets:

1. **Critical (must-fix)**: breaks navigation, accessibility, core data workflows, or v8/v9 mixing
2. **High**: causes inconsistency, theme breakage, or recurring implementation drift
3. **Medium**: quality improvements with moderate impact (token drift, missing ARIA labels)
4. **Low**: polish and non-blocking refinements

Include:

- File paths with line references
- Category: Token | Navigation | DataGrid | Accessibility | Theme | Performance
- Concrete change recommendations with v9 component/API to use
- Expected impact
- v8 → v9 migration items flagged separately

**Success Criteria:**

- [ ] Findings prioritized by severity
- [ ] Recommendations are actionable and file-specific
- [ ] v8 migration items have clear v9 replacement path
