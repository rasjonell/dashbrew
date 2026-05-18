# Examples

Four ready-to-run dashboards, each leaning into a different visual identity. Run any of them from the repo root:

```bash
dashbrew -c examples/<name>.json
```

The dashboards reference helper scripts under [`examples/scripts/`](./scripts) using relative paths, so the working directory needs to be the repo root.

## Dependencies

All examples assume a POSIX shell, `awk`, and the usual coreutils. Nothing else. The trading-desk and focus dashboards are fully offline; they simulate data locally so they never hit a network API or a rate limit.

## minimal.json

The smallest valid dashboard: one component, one shell source, almost no styling. Use this as the starting point when you are learning the JSON shape.

```bash
dashbrew -c examples/minimal.json
```

![minimal](./gifs/minimal.gif)

## focus.json

A cozy personal cockpit on the `catppuccin-mocha` theme: a giant ASCII clock at the top, a 25/5 pomodoro focus chart in append mode, a todo list pulled from `./todo.txt`, and an hourly progress histogram with gradient bars.

Run:

```bash
dashbrew -c examples/focus.json
```

Showcases:

- a tall hero `text` pane rendered from a custom shell script that draws block-character digits
- `chart` in `append` mode wired to a 25/5 work-break synthetic signal
- `todo` reading the repo's `todo.txt`
- `histogram` with `barStyle: gradient`
- `style.focusMode: "highlight"` so the focused pane visibly pops

![focus](./gifs/focus.gif)

## trading-desk.json

A `gruvbox-dark` crypto cockpit. The top row shows the `TRADES` ASCII banner next to a live BTC USD chart. Below it sit four ticker tiles (BTC / ETH / SOL / BNB) showing price, 24-hour change and volume. The bottom is a per-symbol coloured market-cap histogram.

Prices are simulated by `examples/scripts/prices.sh` using a per-symbol random walk with mild mean reversion (state lives under `/tmp/dashbrew-sim/`). No network, no API limits.

Run:

```bash
dashbrew -c examples/trading-desk.json
```

Showcases:

- many `script`-sourced panes wired to a shared simulator
- `chart` in `append` mode pulling from the simulator on a 1-second cadence
- `histogram` with `barStyle: map` and a per-bin `barColors` palette
- a named theme (`gruvbox-dark`) shared across the whole dashboard

![trading-desk](./gifs/trading-desk.gif)

## devops-noc.json

A production-NOC board on the `tokyo-night` theme. Top row of four service-health tiles (api / db / cache / queue), each with its own coloured border and a `d` binding that drills down into a detail page. A full-width p99 latency chart with an `s` binding for a 60-sample snapshot. A status-code histogram and live log tail across the bottom; the log pane has `e` and `w` bindings for errors-only and warnings-only filtered views.

Run:

```bash
dashbrew -c examples/devops-noc.json
```

Bindings cheat sheet:

| pane            | key | what it does                                |
| --------------- | --- | ------------------------------------------- |
| api tile        | `d` | replace pane with API service detail        |
| db tile         | `d` | replace pane with DB service detail         |
| cache tile      | `d` | replace pane with cache service detail      |
| queue tile      | `d` | replace pane with queue service detail      |
| p99 latency     | `s` | snapshot 60 samples via `output_as: inherit` |
| live log        | `e` | dense error-only log view                   |
| live log        | `w` | dense warning-only log view                 |

Press `Ctrl+K` to see every binding in the global palette.

Showcases:

- per-component `style.palette` overrides on top of a shared theme
- `style.focusMode: "highlight"` for visual focus
- `replace_pane` actions in both `text` and `inherit` output modes
- a mix of `chart`, `histogram`, `text`, and per-tile colour coding

![devops-noc](./gifs/devops-noc.gif)

## GIF assets

GIFs live under `examples/gifs/`. Filenames mirror the JSON basename (`focus.gif`, `trading-desk.gif`, etc.) so the embeds above keep working once the assets land. Drop one in, push, and the README plus this page update automatically.
