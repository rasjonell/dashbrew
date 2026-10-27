# Changelog

All notable changes to dashbrew are recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Nothing yet.

## [1.2.0]

### Added

- Example dashboards under `examples/`: `focus`, `trading-desk`, `devops-noc`, `minimal`, with their helper scripts and an `examples/README.md` index.

### Fixed

- Table panes overflowing their border: column widths now account for per-column padding, and long pane titles are truncated with `…` ([#3](https://github.com/rasjonell/dashbrew/pull/3), thanks @QueenDekim).

## [1.1.0]

Big v2 themed feature drop. Old v1.0 configs keep working unchanged; every new field is optional with sensible defaults.

### Added

#### Component interactivity

- Per-component `bindings: [{key, label, action}]`. Pressing the bound key on the focused pane runs a shell action.
- Two action types: `fire_and_forget` (spawn, ignore output) and `replace_pane` (spawn, capture stdout, show until reset).
- `output_as: "text" | "inherit"` on replace_pane actions. `inherit` routes the action's stdout through the component's normal data parser so a chart action can return new numbers, a histogram action can return new bins, etc.
- Global `Ctrl+K` action palette aggregating every binding across every component, fuzzy filterable, run with `Enter`.
- Spinner overlay while a `replace_pane` action is in flight, sticky output afterwards until `R` (refetch + resume schedule) or `Esc` (clear only).
- Centered red-bordered error modal when an action fails (non-zero exit, spawn error, or timeout). Dismiss with `Esc`.
- Per-action `timeout_seconds`, defaulting to 30 seconds.
- Config-load validation that rejects bindings using any reserved global key (`ctrl+c`, `ctrl+k`, `shift+arrows`, `H/J/K/L`, `a/A`, `r/R`, `esc`) or any key that conflicts with a component type's internal keys (e.g. `space` on todo).

#### Theming

- Named theme presets: `default`, `dracula`, `nord`, `gruvbox-dark`, `tokyo-night`, `catppuccin-mocha`, `solarized-light`.
- `style.theme: "auto"` to detect terminal background via OSC 11 and pick the dark or light default automatically.
- Role-based palette: `style.palette.{background, foreground, dim, accent, success, warning, error, border, borderFocused}` layered on top of the active theme.
- Per-component overrides: `component.style.{palette, borderType}` so one pane can break from the dashboard-wide theme (useful for alerts panes, etc.).
- Custom theme files loaded from `$DASHBREW_THEMES_DIR` or `$XDG_CONFIG_HOME/dashbrew/themes` so users can ship their own without recompiling.
- CLI flags: `-t / --theme NAME` to override the active theme at startup, `--list-themes` to print the registry.
- Theme-driven colors wired through every component: chart series/axis/label/caption, table header + selected row, histogram bars, todo done items (strikethrough plus success-coloured `✓`).

#### Layout and visuals

- `style.layout.gap` to put a configurable number of whitespace cells between adjacent panes inside row/column containers.
- `style.focusMode: "highlight"` to dim every non-focused pane so the focused one visually pops.
- Component-level `titleAlign: "left" | "center" | "right"` with adaptive title-cell borders. The title cell only draws the border edges that face inward toward the pane, so a right-aligned title gets left and bottom borders, a centered title gets left, right, and bottom, and a left-aligned title keeps right and bottom.
- Component-level `icon` field that prepends a literal glyph to the title.
- Focused-pane title accent: bold title text in the focused color plus matching color on the title cell's outward-facing borders.
- Single-row help line at the bottom of the terminal, contextual to the current mode (default, add-mode, action-active, action-loading, palette-open, error-modal) and the focused component's bindings. Subtle separator rule above it, faint labels with key chords highlighted.
- Histogram `barStyle` modes: `single` (theme accent), `alternate` (accent / dim), `gradient` (LAB-blended dim to accent by value), and `map` (per-bin colors via `barColors`).

### Fixed

- Text component height overflow: `JoinVertical` was treating an empty footer slot as one row, plus footer reservation was using stale viewport state. Both fixed; the pane now hits its allocated height exactly.
- Histogram height drift that made the right column render one row taller than the left, leaving an empty band below the Posts Table.
- `fire_and_forget` actions killing the spawned shell before it could finish: the runner now hands ownership of the cancel to the wait goroutine instead of deferring it on the closure.
- Pre-existing `msg.Width, msg.Width` typo in the not-yet-initialized resize branch (was supposed to be `msg.Width, msg.Height`).

### Changed

- The replace_pane sticky overlay is now a scrollable `bubbles/viewport` rather than fixed truncated text, so a long action output (`cat README.md`, `tail -n 500 file`) can be paged with arrow keys, page-up/down, home/end, or the mouse wheel.
- Default focused border color in the embedded `default` theme moved from `#444444` (which was barely visible on dark terminals) to `#bd93f9`. Existing configs that set `style.border.focusedColor` still win.

## [1.0.0]

Initial public release.

### Added

- Six component types: `text`, `list`, `todo`, `chart`, `histogram`, `table`.
- Three data sources: `script` (shell command), `api` (HTTP GET with optional JSONPath extraction), and direct file paths for todo.
- Container layout primitives with `row` / `column` direction and `flex` sizing.
- Per-component `refresh_interval` (seconds) and chart-specific `refresh_mode: "append"` for accumulating series.
- Configurable border styling via `style.border.{type, color, focusedColor}`.
- Global keymap: `Shift+arrow` / `H J K L` for navigation between panes, `R` to refresh the focused pane, `A` to enter add mode (todo), `Space` to toggle a todo, `Ctrl+C` to quit.
- Mouse click to focus a pane.
- Table component with per-column `label`, `field` (JSONPath-style key for object rows), and `flex`.

[Unreleased]: https://github.com/rasjonell/dashbrew/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/rasjonell/dashbrew/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/rasjonell/dashbrew/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/rasjonell/dashbrew/releases/tag/v1.0.0
