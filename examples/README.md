# Examples

Four ready-to-run dashboards, each leaning into a different visual identity. Run any of them from the repo root:

```bash
dashbrew -c examples/<name>.json
```

The dashboards reference helper scripts under [`examples/scripts/`](./scripts) using relative paths, so the working directory needs to be the repo root.

## Dependencies

- All examples assume a POSIX shell, `awk`, and `head`/`tail` (any modern macOS or Linux box has these).
- `trading-desk.json` additionally needs `curl` and `jq` for the live coinpaprika tickers. Install with `brew install jq` on macOS or your distro's package manager.

## minimal.json

The smallest valid dashboard: one component, one shell source, almost no styling. Use this as the starting point when you are learning the JSON shape.

```bash
dashbrew -c examples/minimal.json
```

![minimal](./gifs/minimal.gif)

## hackerman.json

Cyberpunk-flavoured ops board with a `MATRIX` ASCII banner, live network-throughput chart, rolling threat feed, fake netstat-style session list, and a gradient-bar histogram of attack vectors. Custom green-on-near-black palette overrides the default, with cyan focused borders and thick lipgloss frames.

Run:

```bash
dashbrew -c examples/hackerman.json
```

Showcases:

- top-level `style.palette` override producing a fully custom color story
- `chart` in `append` mode for a continuously-pulsing throughput trace
- `histogram` with `barStyle: gradient` for value-driven bar shading
- mixed refresh intervals so the panes shimmer at different cadences

![hackerman](./gifs/hackerman.gif)

## trading-desk.json

Bloomberg-amber crypto cockpit on the `gruvbox-dark` theme. Four live coinpaprika tickers (BTC, ETH, SOL, BNB) with formatted price, 24-hour percent change, and volume; a live BTC USD price chart in `append` mode; a per-symbol coloured market-cap histogram (`barStyle: map`); and a fortune-cookie wisdom feed in the corner.

Run:

```bash
dashbrew -c examples/trading-desk.json
```

Showcases:

- multiple `api`-sourced panes
- `script`-sourced panes piping `curl` through `jq`
- `chart` in `append` mode wired to a live API endpoint
- `histogram` with `barStyle: map` and a per-bin `barColors` palette
- a named theme (`gruvbox-dark`) shared across the whole dashboard

![trading-desk](./gifs/trading-desk.gif)

## devops-noc.json

Production-NOC board on the `tokyo-night` theme. Top row of four service-health tiles (api / db / cache / queue) each with its own coloured border, a full-width p99 latency chart, a status-code histogram, and a live log tail across the bottom. Each tile uses `style.palette` to claim a distinct accent so you can tell them apart instantly.

Run:

```bash
dashbrew -c examples/devops-noc.json
```

Showcases:

- per-component `style.palette` and `style.borderType` overrides on top of a shared theme
- `style.focusMode: "highlight"` to dim the non-focused panes
- `chart` + `histogram` + `list` + `text` in one cohesive layout
- realistic mock data so the panels never look empty

![devops-noc](./gifs/devops-noc.gif)

## GIF assets

GIFs live under `examples/gifs/`. Filenames mirror the JSON basename (`hackerman.gif`, `trading-desk.gif`, etc.) so the embeds above keep working once the assets land. Drop one in, push, and the README plus this page update automatically.
