# Design Tokens Standard

## Purpose

Ensure branding is implemented through Fluent UI React v9 design tokens — a stable, theme-aware token contract rather than ad-hoc per-component styling. All colors, spacing, typography, and motion use `tokens.*` from `@fluentui/react-components`.

---

## FluentProvider Setup

The app root MUST wrap everything in a single `FluentProvider`:

```tsx
import {
  FluentProvider,
  webLightTheme,
  webDarkTheme,
  type Theme,
} from "@fluentui/react-components";

const themes: Record<string, Theme> = {
  light: webLightTheme,
  dark: webDarkTheme,
};

// Resolve from: user preference → system preference → default light
const resolvedTheme =
  themes[userPreference] ?? themes[systemPreference] ?? webLightTheme;

<FluentProvider theme={resolvedTheme}>
  <App />
</FluentProvider>;
```

### Rules

- ONE `FluentProvider` at the app root — do not nest providers unless creating an isolated theme region (e.g., a high-contrast preview)
- Theme selection respects `prefers-color-scheme` media query as fallback
- Persist user theme preference (localStorage or user settings API)
- NEVER reference raw theme objects in components — always use `tokens.*`
- Runtime theme switching (light/dark) is mandatory for production surfaces
- Validate all primary flows after switching themes; no visual regressions are acceptable
- Validate Windows High Contrast behavior using `forced-colors: active`; explicit high-contrast themes are optional and requirement-driven
- Ensure portals (dialogs, toasts, overlays) inherit token variables from the root `FluentProvider`; avoid nested providers that fragment token context

---

## Required Token Layers

### 1) Brand Colors

Use Fluent's built-in brand ramp or create a custom `BrandVariants` for organization branding:

| Token                                  | Usage                             |
| -------------------------------------- | --------------------------------- |
| `tokens.colorBrandBackground`          | Primary action button backgrounds |
| `tokens.colorBrandBackgroundHover`     | Primary hover state               |
| `tokens.colorBrandBackgroundPressed`   | Primary pressed state             |
| `tokens.colorBrandForegroundOnLight`   | Brand text on light surfaces      |
| `tokens.colorNeutralForegroundOnBrand` | Text on brand-colored backgrounds |
| `tokens.colorCompoundBrandForeground1` | Links, accents                    |

### 2) Neutral Colors — Surface Hierarchy

| Token                                 | Usage                                        |
| ------------------------------------- | -------------------------------------------- |
| `tokens.colorNeutralBackground1`      | Page background                              |
| `tokens.colorNeutralBackground2`      | Card/surface background                      |
| `tokens.colorNeutralBackground3`      | Elevated surface (drawer, dialog)            |
| `tokens.colorNeutralForeground1`      | Primary text                                 |
| `tokens.colorNeutralForeground2`      | Secondary text                               |
| `tokens.colorNeutralForeground3`      | Tertiary/disabled text                       |
| `tokens.colorNeutralStroke1`          | Borders and dividers                         |
| `tokens.colorNeutralStrokeAccessible` | Borders that must meet contrast requirements |

### 3) Semantic Status Colors

Map ONLY to Fluent's semantic status tokens — never create custom status palettes:

| Status  | Background                             | Foreground                             | Border                             |
| ------- | -------------------------------------- | -------------------------------------- | ---------------------------------- |
| Success | `tokens.colorStatusSuccessBackground1` | `tokens.colorStatusSuccessForeground1` | `tokens.colorStatusSuccessBorder1` |
| Warning | `tokens.colorStatusWarningBackground1` | `tokens.colorStatusWarningForeground1` | `tokens.colorStatusWarningBorder1` |
| Danger  | `tokens.colorStatusDangerBackground1`  | `tokens.colorStatusDangerForeground1`  | `tokens.colorStatusDangerBorder1`  |
| Info    | —                                      | —                                      | —                                  |

Use `MessageBar` intents (`success`, `warning`, `error`, `info`) and `Badge` colors for status rendering — they automatically use correct semantic tokens.

Exception policy:

- Palette tokens (`tokens.colorPalette*`) are allowed only for data visualization or non-semantic decorative accents
- Business status and health states must use semantic status tokens or Fluent intents

### 4) Typography

| Token                        | Usage                       |
| ---------------------------- | --------------------------- |
| `tokens.fontFamilyBase`      | All body text               |
| `tokens.fontFamilyMonospace` | Code, IDs, technical values |
| `tokens.fontSizeBase200`     | Small/caption text          |
| `tokens.fontSizeBase300`     | Body text (default)         |
| `tokens.fontSizeBase400`     | Subtitle text               |
| `tokens.fontSizeBase500`     | Section headings            |
| `tokens.fontSizeBase600`     | Page title                  |
| `tokens.fontWeightRegular`   | Body weight                 |
| `tokens.fontWeightSemibold`  | Headings, emphasis          |
| `tokens.lineHeightBase300`   | Default line height         |

### 5) Spacing, Radius, and Shadow

| Token                       | Usage                         |
| --------------------------- | ----------------------------- |
| `tokens.spacingHorizontalS` | Tight inline spacing (4px)    |
| `tokens.spacingHorizontalM` | Standard inline spacing (8px) |
| `tokens.spacingHorizontalL` | Loose inline spacing (12px)   |
| `tokens.spacingVerticalM`   | Standard block spacing        |
| `tokens.spacingVerticalXL`  | Section separation            |
| `tokens.borderRadiusMedium` | Cards, buttons, inputs        |
| `tokens.borderRadiusLarge`  | Dialogs, drawers              |
| `tokens.shadow4`            | Elevated cards                |
| `tokens.shadow8`            | Drawers, dialogs              |
| `tokens.shadow16`           | Popovers, menus               |

### 6) Motion

| Token                       | Usage                        |
| --------------------------- | ---------------------------- |
| `tokens.durationNormal`     | Standard transitions (200ms) |
| `tokens.durationSlow`       | Drawer open/close            |
| `tokens.curveEasyEase`      | Default easing               |
| `tokens.curveDecelerateMax` | Enter animations             |
| `tokens.curveAccelerateMax` | Exit animations              |

Use `prefers-reduced-motion: reduce` media query to disable animations for accessibility.

---

## Custom Theme Extension

For organization-specific branding on top of Fluent:

```tsx
import {
  createLightTheme,
  createDarkTheme,
  type BrandVariants,
} from "@fluentui/react-components";

const orgBrand: BrandVariants = {
  10: "#020305",
  20: "#111723",
  /* ... full 16-step ramp */ 160: "#F0F4FF",
};

const orgLightTheme = { ...createLightTheme(orgBrand) };
const orgDarkTheme = { ...createDarkTheme(orgBrand) };
```

### Transitional Bridge from Fluent v8 Azure Themes

For migrations from legacy stacks, a temporary bridge may map v8 Azure palette values into v9 `BrandVariants`.

Guardrails:

- Bridge exists only in the theme module (single file), never in components
- v8 components (`@fluentui/react`) remain disallowed
- New code consumes only v9 `tokens.*`
- Remove the bridge once equivalent v9 brand variants are stabilized

### Rules

- Custom brand ramps MUST maintain WCAG AA contrast ratios across all 16 steps
- Custom themes still use the full Fluent token set — override only brand colors
- Test custom themes in light, dark, AND high-contrast modes

---

## Rules Summary

1. **Single Source of Truth** — All token/theme configuration in one module (e.g., `src/theme.ts`)
2. **No Color Scatter** — Zero hex/rgb literals in component files; use `tokens.*` exclusively
3. **Semantic Usage** — Components consume semantic tokens, not color primitives
4. **State Completeness** — Hover, pressed, selected, disabled states defined via Fluent token variants
5. **Contrast Safety** — Text/background combinations auto-guaranteed by Fluent tokens when used correctly
6. **Theme Portability** — Switching `FluentProvider` theme instantly updates all components (no additional code)

---

## Good vs Bad

✅ Good:

```tsx
// Using tokens for all styling
import { tokens, makeStyles } from "@fluentui/react-components";

const useStyles = makeStyles({
  card: {
    backgroundColor: tokens.colorNeutralBackground2,
    borderRadius: tokens.borderRadiusMedium,
    padding: tokens.spacingHorizontalL,
    boxShadow: tokens.shadow4,
  },
  statusText: {
    color: tokens.colorStatusSuccessForeground1,
    fontSize: tokens.fontSizeBase300,
  },
});
```

❌ Bad:

```tsx
// Inline color literals — breaks theming, accessibility, and maintainability
const card = {
  backgroundColor: "#1A1E3A",
  borderRadius: "8px",
  padding: "12px",
};
const statusText = { color: "#4CAF50" };
```
