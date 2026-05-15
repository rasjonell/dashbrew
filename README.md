```
       /$$                     /$$       /$$                                        
      | $$                    | $$      | $$                                        
  /$$$$$$$  /$$$$$$   /$$$$$$$| $$$$$$$ | $$$$$$$   /$$$$$$   /$$$$$$  /$$  /$$  /$$
 /$$__  $$ |____  $$ /$$_____/| $$__  $$| $$__  $$ /$$__  $$ /$$__  $$| $$ | $$ | $$
| $$  | $$  /$$$$$$$|  $$$$$$ | $$  \ $$| $$  \ $$| $$  \__/| $$$$$$$$| $$ | $$ | $$
| $$  | $$ /$$__  $$ \____  $$| $$  | $$| $$  | $$| $$      | $$_____/| $$ | $$ | $$
|  $$$$$$$|  $$$$$$$ /$$$$$$$/| $$  | $$| $$$$$$$/| $$      |  $$$$$$$|  $$$$$/$$$$/
 \_______/ \_______/|_______/ |__/  |__/|_______/ |__/       \_______/ \_____/\___/ 
```

**Dashbrew** is a terminal dashboard builder that lets you visualize data from scripts and APIs right in your console, using a simple JSON configuration. Stay informed without leaving your terminal!

[![BuyMeACoffee](https://raw.githubusercontent.com/pachadotdev/buymeacoffee-badges/main/bmc-yellow.svg)](https://www.buymeacoffee.com/rasjonell)

![screenshot](./screen.gif)

## 🚀 Installation

### Using Go Install

Pull the latest release:

```bash
go install github.com/rasjonell/dashbrew/cmd/dashbrew@latest
```

Pin a specific version (any tag from the [Releases](https://github.com/rasjonell/dashbrew/releases) page):

```bash
go install github.com/rasjonell/dashbrew/cmd/dashbrew@v1.1.0
go install github.com/rasjonell/dashbrew/cmd/dashbrew@v1.0.0
```

### From Source

```bash
# Clone the repository
git clone https://github.com/rasjonell/dashbrew.git
cd dashbrew

# Build and install
go install ./cmd/dashbrew
```

## Quick Start

1. Create a simple dashboard configuration file(`dashboard.json`):
```jsonc
{
  "style": {
    "border": {
      "type": "thicc",
      "color": "#cccccc",
      "focusedColor": "#474747"
    }
  },
  "layout": {
    "type": "container",
    "direction": "row",
    "children": [
      {
        "type": "component",
        "flex": 1,
        "component": {
          "type": "text",
          "title": "Hello Dashbrew",
          "data": {
            "source": "script",
            "command": "echo 'Welcome to Dashbrew!'"
          }
        }
      }
    ]
  }
}
```

2. run:

```bash
dashbrew -c dashboard.json
```

## Complete Documentation

For comprehensive documentation on all features, please refer to our [GitHub Wiki](https://github.com/rasjonell/dashbrew/wiki):

- [Layout System](https://github.com/rasjonell/dashbrew/wiki/Layout-System)
- [Components](https://github.com/rasjonell/dashbrew/wiki/Components)
- [Styling Options](https://github.com/rasjonell/dashbrew/wiki/Styling-Options)

## Basic Examples

### Display API Data

Show real-time data from an API:

```jsonc
{
  "type": "component",
  "component": {
    "type": "text",
    "title": "🌦️ Weather",
    "data": {
      "source": "api",
      "url": "https://wttr.in/<YOUR_CITY>?format=4",
      "refresh_interval": 60
    }
  }
}
```

### Creating a ToDo List

Create a `todo.txt` file:

```
- finish work
+ laundry
+ dishes
```

Add it to your dashboard:

```jsonc
{
  "type": "component",
  "component": {
    "type": "todo",
    "title": "📋 My Todo List",
    "data": {
      "source": "./todo.txt"
    }
  }
}
```

### Creating a Chart

Visualize Data with charts:

```jsonc
{
  "type": "component",
  "component": {
    "type": "chart",
    "title": "📊 System Metrics",
    "data": {
      "source": "script",
      "command": "echo '10\n25\n15\n30\n45'",
      "refresh_interval": 5,
      "caption": "CPU Usage (%)"
    }
  }
}
```

### Create a Histogram

Show distributions with histograms:

```jsonc
{
  "type": "component",
  "component": {
    "type": "histogram",
    "title": "📊 Age Distribution",
    "data": {
      "source": "script",
      "command": "echo '\"18-24\": 45\n\"25-34\": 78\n\"35-44\": 52\n\"45-54\": 34\n\"55+\": 21'",
      "caption": "Users by Age Group"
    }
  }
}
```

## Basic Navigation

- `Shift+Arrow` or `Shift` + `H/J/K/L`: Move between components
- `A`: Add item (in todo lists)
- `Space`: Toggle item state (in todo lists)
- `R`: Refresh data for the focused component
- `Ctrl+K`: Open the action palette (lists every binding from every component, fuzzy filterable)
- `Ctrl+C`: Quit

## Component Interactivity

Each component can declare a list of **bindings** that map a key chord to a shell action. Pressing the key when that component is focused runs the action; the same actions are also reachable via the global `Ctrl+K` palette.

```jsonc
{
  "type": "component",
  "component": {
    "id": "logs",
    "type": "text",
    "title": "📜 Logs",
    "data": { "source": "script", "command": "tail -n 50 ./app.log", "refresh_interval": 5 },
    "bindings": [
      {
        "key": "f",
        "label": "Show last 500 lines",
        "action": {
          "type": "replace_pane",
          "command": "tail -n 500 ./app.log",
          "output_as": "text"
        }
      },
      {
        "key": "c",
        "label": "Clear log file",
        "action": { "type": "fire_and_forget", "command": ": > ./app.log" }
      }
    ]
  }
}
```

### Action types

- **`fire_and_forget`**: runs the command in the background and discards its output. The pane stays unchanged. Good for "open weather.app", "trigger refresh", "clear logs".
- **`replace_pane`**: runs the command, then replaces the pane's content with its stdout until you reset. Scheduled refresh is paused while the action result is sticky. Press **R** to refetch the original data and resume the schedule, or **Esc** to just clear the action state without refetching.

### Output rendering for `replace_pane`

- `output_as: "text"` (default): stdout is shown as raw scrollable text.
- `output_as: "inherit"`: stdout is parsed by the focused component's existing data parser. So a chart action expects newline-separated numbers, a histogram action expects `key: count` lines or a JSON object, a table action expects JSON rows, etc. Not supported on `todo` (its data path is file-based).

### Reserved keys

These keys are reserved by dashbrew and cannot be used as binding keys:

`ctrl+c`, `ctrl+k`, `shift+up/down/left/right`, `H`, `J`, `K`, `L`, `a`, `A`, `r`, `R`, `esc`

In addition, each component type has its own **internal keys** (e.g. `space` toggles a todo item, viewport-style components consume `up/down/pgup/pgdown/home/end` for scrolling). Bindings on a component cannot use that type's internal keys. Both checks happen at config load time with a clear error.

### `timeout_seconds`

Each action can declare `timeout_seconds` (default 30). The shell process is killed via `context.WithTimeout` and surfaces an error modal on timeout. Long-running stream commands (e.g. `tail -f`) are not supported yet; use a snapshot like `tail -n 500 file.log` instead.

## Theming

Dashbrew ships with a small set of named themes (`default`, `dracula`, `nord`, `gruvbox-dark`, `tokyo-night`, `catppuccin-mocha`, `solarized-light`). Pick one from JSON:

```jsonc
{ "style": { "theme": "dracula" } }
```

Or set `"theme": "auto"` to detect the terminal's background color (OSC 11) and pick a dark or light default.

Override at runtime without editing the config:

```bash
dashbrew -c dashboard.json -t nord
dashbrew --list-themes
```

Custom themes drop into `$XDG_CONFIG_HOME/dashbrew/themes/*.json` (or set `DASHBREW_THEMES_DIR`) and are picked up automatically. Each theme is a `name` plus a `palette` of role colors (`background`, `foreground`, `dim`, `accent`, `success`, `warning`, `error`, `border`, `borderFocused`). The same shape works for `style.palette` overrides at the dashboard or per-component level.

## Versioning & Releases

Dashbrew follows [Semantic Versioning](https://semver.org/). See the [CHANGELOG](./CHANGELOG.md) for what shipped in each release and the [Releases page](https://github.com/rasjonell/dashbrew/releases) for tag-by-tag notes.

`@latest` always points at the newest tag. To pin a release explicitly:

```bash
go install github.com/rasjonell/dashbrew/cmd/dashbrew@v1.1.0
```

## License

[MIT License](./LICENSE)


## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=rasjonell/dashbrew&type=Date)](https://www.star-history.com/#rasjonell/dashbrew&Date)
