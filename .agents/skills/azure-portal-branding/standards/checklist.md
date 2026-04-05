# Azure Portal Branding Checklist

## Purpose

Go/no-go checklist for shipping frontend branding changes with native Azure portal-style UX using Fluent UI React v9. Every item is a hard requirement unless explicitly scoped out.

---

## Scope Setup

- [ ] Token source file identified (e.g., `src/theme.ts`)
- [ ] `FluentProvider` wraps the entire app at root level
- [ ] Shared shell/layout file identified (app header, `NavDrawer`, content area)
- [ ] At least one list/detail page included in review
- [ ] Scope boundaries captured (what is intentionally out of scope)
- [ ] Only `@fluentui/react-components` (v9) imports — zero v8 imports
- [ ] Global production primitives present: App shell, Error boundary, Toast provider, Skip link

---

## Token Discipline

- [ ] Brand, neutral, and status token layers exist in theme module
- [ ] Typography tokens (`fontSizeBase*`, `fontWeightSemibold`, etc.) used — no raw `font-size` values
- [ ] Spacing tokens (`spacingHorizontal*`, `spacingVertical*`) used — no raw `px` padding/margin values
- [ ] Interaction states (hover/pressed/disabled/selected) use Fluent token variants
- [ ] Zero hex/rgb color literals in component files — `tokens.*` exclusively
- [ ] Semantic status tokens (`colorStatusSuccess*`, `colorStatusDanger*`, etc.) used for all status rendering
- [ ] Palette tokens (`colorPalette*`) used only for data visualization/decorative accents, not business status states
- [ ] Custom brand ramp (if used) passes WCAG AA contrast across all 16 steps
- [ ] Contrast verified for text/surface pairings in light, dark, AND high-contrast themes

---

## Theme Support

- [ ] App supports `webLightTheme` and `webDarkTheme`
- [ ] Theme switching mechanism exists (user preference + system fallback)
- [ ] User theme preference persists across sessions
- [ ] Runtime theme toggle is available and usable without page reload
- [ ] All components render correctly when theme is switched — no hardcoded colors break
- [ ] `prefers-color-scheme` media query respected as fallback
- [ ] `forced-colors: active` media query handled for Windows High Contrast Mode
- [ ] Explicit high-contrast theme is added only when product requirements require it

---

## Shell and Navigation

- [ ] App shell is stable and consistent across pages
- [ ] Skip link (`Skip to main content`) is present and focuses `main` region
- [ ] `NavDrawer` used for left rail navigation with `aria-label`
- [ ] `Breadcrumb` reflects resource hierarchy with `aria-label="Breadcrumb"`
- [ ] Each page has one `Toolbar` (command bar) for actions
- [ ] Primary action uses `appearance="primary"` at leftmost position
- [ ] Overflow actions use `Menu` with `MoreHorizontalRegular` icon
- [ ] Icon-only buttons have `aria-label`

---

## Drawers (Blades/Panels)

- [ ] `OverlayDrawer` or `InlineDrawer` used for detail/edit flows (not custom modals)
- [ ] Drawer `position="end"` (right-anchored) with appropriate `size`
- [ ] Multi-blade workflows use stacked non-modal drawers with clear depth offsets
- [ ] Opening/closing drawer preserves parent context (filters, sort, search, selection, scroll)
- [ ] Drawer/open-panel state is URL-addressable where feasible (route or search params)
- [ ] Browser Back closes active drawer before navigating away
- [ ] Close button has `aria-label="Close"`
- [ ] `Escape` closes drawer (or prompts on unsaved changes)
- [ ] Focus returns to trigger element after close
- [ ] Responsive: switches from `InlineDrawer` to `OverlayDrawer` at narrow viewports

---

## Filters and Search

- [ ] Filter/search/sort logic centralized in single hook or context
- [ ] `SearchBox` debounced (300ms minimum)
- [ ] In-flight requests cancelled on filter change (`AbortController`)
- [ ] Filter state synced with URL query params for shareability
- [ ] Incremental search/filter URL updates avoid history spam (replace behavior)
- [ ] "Clear filters" action available
- [ ] Active filters shown as dismissible `Tag` elements in `TagGroup`
- [ ] Result count announced via `aria-live="polite"`
- [ ] `Combobox` uses `inlinePopup={true}` where VoiceOver compatibility is required

---

## DataGrid (Resource Lists)

- [ ] `DataGrid` is the default for tabular resource data unless advanced non-standard behavior requires `Table` primitives
- [ ] DataGrid/Table selection is intentional: use `DataGrid` for standard use, `Table` + `useTableFeatures` for non-standard advanced needs
- [ ] Columns defined with `createTableColumn` and `compare` functions
- [ ] `sortable` prop enabled on DataGrid
- [ ] `selectionMode="multiselect"` configured
- [ ] Selection cell has `aria-label` ("Select all rows" / "Select row")
- [ ] Row-level actions in dedicated column with `aria-label` on icon buttons
- [ ] Bulk actions enabled/disabled based on selection count in command bar
- [ ] Loading state: `Spinner` or skeleton rows
- [ ] Empty state: illustration + message + primary create action
- [ ] Error state: message + retry action
- [ ] Arrow key navigation functional within grid
- [ ] `aria-label` on `DataGrid` describing the dataset

---

## Forms and Dialogs

- [ ] All form controls wrapped in `Field` (label, validation, required indicator)
- [ ] First input receives focus on form/dialog mount
- [ ] Destructive actions require `Dialog` confirmation
- [ ] Inline validation per-field + summary at submit
- [ ] `Dialog` components used from Fluent v9 (not custom modals)

---

## Notifications

- [ ] `MessageBar` used for inline/persistent status messages
- [ ] `Toast` via `useToastController` for transient notifications
- [ ] Global toast provider is configured at app root (not per-page)
- [ ] Success toasts auto-dismiss (5s), error toasts persist
- [ ] Error messages are user-readable with optional diagnostic detail
- [ ] `aria-live="polite"` on dynamic status updates

---

## Resilience and Compliance Baseline

- [ ] Global `ErrorBoundary` wraps routed app content
- [ ] Unhandled UI failures render a safe fallback view with recovery path
- [ ] Consent/telemetry notice pattern exists where analytics or tracking is enabled

---

## Accessibility (WCAG 2.2 AA)

### Keyboard

- [ ] Tab order: shell → nav → breadcrumb → toolbar → content → drawer
- [ ] Keyboard-modality detection is implemented where needed (for example, `useKeyboardNavAttribute`)
- [ ] `Enter`/`Space` activates buttons, selects rows, opens menus
- [ ] `Escape` closes drawers, dialogs, menus
- [ ] Arrow keys navigate within DataGrid, menus, tabs
- [ ] No keyboard traps — user can always tab out of any region
- [ ] Focus indicators visible on all interactive elements

### Screen Readers

- [ ] Tested with NVDA (Windows) or VoiceOver (macOS)
- [ ] All ARIA landmarks present: `nav`, `main`, `search`, `toolbar`
- [ ] `aria-label` on: icon-only buttons, DataGrid, Breadcrumb, Toolbar, NavDrawer
- [ ] `aria-live` regions for: result counts, toast notifications, status changes
- [ ] `role="status"` for filter result count announcements
- [ ] Drawer/dialog transitions announced properly

### Visual

- [ ] Color never sole indicator of state — always paired with icon/text
- [ ] Text contrast ratio ≥ 4.5:1 (AA) for normal text, ≥ 3:1 for large text
- [ ] Interactive element contrast ratio ≥ 3:1 against adjacent colors
- [ ] Touch targets minimum 44x44px
- [ ] Minimum viewport support: 320px width (WCAG 1.4.10 Reflow)
- [ ] `prefers-reduced-motion: reduce` disables animations
- [ ] Custom focus styles use Fluent focus utilities (for example, `createFocusOutlineStyle`) rather than ad-hoc outlines

### High Contrast

- [ ] Tested in Windows High Contrast Mode
- [ ] `forced-colors: active` media query handled where custom styling exists
- [ ] No information lost in high-contrast mode

---

## Performance

- [ ] DataGrid rows > 100: server-side paging or virtualization implemented
- [ ] Page size selector available (20, 50, 100)
- [ ] No client-side sorting/filtering of datasets > 1,000 rows
- [ ] Debounced search prevents excessive API calls
- [ ] Large images/assets lazy-loaded

---

## Validation

- [ ] ESLint / TypeScript build passes with zero errors
- [ ] Accessibility linter (`eslint-plugin-jsx-a11y`) passes
- [ ] Manual keyboard navigation tested end-to-end
- [ ] Screen reader tested on at least one primary flow
- [ ] Theme switching tested (light → dark → high-contrast → light)
- [ ] Theme toggle verification captured for at least one list page and one drawer flow
- [ ] Known limitations documented with follow-up actions
- [ ] File-level change summary prepared for review
