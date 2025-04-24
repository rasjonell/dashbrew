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

![screenshot](./screen.png)

---

## ✨ Features

* **Configurable Layout:** Define complex dashboard layouts (rows, columns, flexible sizing)
* **Multiple Data Sources:**
    * `script`: Execute local shell commands/scripts and display their output.
    * `api`: Fetch data from HTTP APIs and display the response body.
    * `todo`: For ToDo lists, read and write a custom todo list text file.
* **Component Types**
    * _Text_: Scrollable, auto-wrapped text output.
    * _List_: Display a list of items from scripts or APIs.
    * _Todo_: Interactive todo list with add/remove and toggle done state support.
* **Terminal UI:** Built with the delightful [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.
* **Navigation:** Easily move focus between components using arrow keys, `hjkl`, or your mouse.
* **Scrolling:** Scroll through component content that doesn't fit using the mouse wheel or `h`/`j`.
* **Text Wrapping:** Long lines within components are automatically wrapped to fit the component's width.
* **Auto-Refresh:** Configure components to automatically refresh their data at specified intervals.
* **Visual Focus Indicator:** Clearly see which component is currently selected.
* **Mouse Support:** Click to focus, scroll with the wheel.

---

## 🚀 Installation & Usage (will be changed)

1. Clone the repository:

```bash
git clone https://github.com/rasjonell/dashbrew.git # Replace with actual repo URL
cd dashbrew
```

2. Build the binary:

```bash
go build -o dashbrew ./cmd/dashbrew
```

3. Create your `dashboard.json`

4. Run:

```bash
./dashbrew -c your_dashboard.json
```

## ⚙️ Configuration (`dashboard.json`)

Dashbrew uses a `json` file in to define the layout and components. Find an example [here](./dashboard.json)

**Structure:**

```jsonc
{
  "style": {
    "borderType": "rounded",
    "borderColor": "#ffffff",
    "focusedBorderColor": "#00ff00"
  },
  "layout": {
    "type": "container", // "container" or "component"
    "direction": "row",  // "row" or "column" (for containers)
    "flex": 1,           // Optional: Relative size factor (default: 1)
    "children": [ ], // Array of child layout nodes (for containers)
    "component": { } // Component definition (for components)
  }
}
```

### Style Configuration:

*All of the style configuration is optional, defaults will be used when not provided.*

- `borderType`: `"rounded"` | `"thicc"` | `"double"` | `"hidden"` | `"normal"` | `"md"` | `"ascii"` | `"block"`
- `borderColor`: any valid hex color (`#ffffff`, `#696969`)
- `focusedBorderColor`: any valid hex color (`#00ff00`, `#694200`)

### Layout Nodes:

- `type`: Can be "container" or "component".
- `direction`: (Only for container) How children are arranged ("row" or "column").
- `flex`: (Optional) An integer determining how space is distributed among siblings. A component with flex: 2 will try to be twice as large (in the container's direction) as a sibling with flex: 1. Defaults to 1.
- `children`: (Only for container) An array of nested layout nodes.
- `component`: (Only for component) Defines the actual widget to display.

### Component Definition:

**Example Structure:**
```jsonc
{
  "id": "unique-component-id"
  "type": "text",
  "title": "My Component Title",
  "data": {
    "source": "script",
    "command": "date +%Y-%m-%d", // Required if source is "script"
    "url": "https://api.example.com/status", // Required if source is "api"
    "refresh_interval": 5 // Optional: Refresh data every 5 seconds. Omit or 0 for no auto-refresh.
  }
}
```

- `id`: (Optional) A unique ID. If omitted, an internal ID is generated.
- `type`: The type of widget. Currently, `text`, `list`, and `todo` components are supported.
- `title`: The title displayed in the component's header.
- `data`: Defines where the component gets its content.
    - `source`: "script", "api", or todo list file path for components with type `todo` (example bellow).
    - `command`: (For script source) The command to execute.
    - `url`: (For api source) The URL to fetch via HTTP GET.
    - `refresh_interval`: (Optional) Time in seconds between data refreshes.

### Text Component

Displays the output of a script or API as scrollable, wrapped text.

```jsonc
{
  "type": "component",
  "component": {
    "type": "text",
    "title": "System Info",
    "data": {
      "source": "script", // "script" or "api"
      "command": "uname -a", // Required if source is "script"
      "url": "https://api.example.com/status", // Required if source is "api"
      "refresh_interval": 10 // Optional: seconds between refreshes
    }
  }
}
```

### List Component

Displays a list of items, with filtering and selection.

- If `source` is `"script"`, each line of output is an item.
- If `source` is `"api"`, the API must return a JSON array of strings.

```jsonc
{
  "type": "component",
  "component": {
    "type": "list",
    "title": "Recent Logs",
    "data": {
      "source": "script", // "script" or "api"
      "command": "cat /var/log/syslog | tail -n 10", // For script
      "url": "https://api.example.com/items", // For api, must return ["item1", "item2", ...]
      "refresh_interval": 5 // Optional
    }
  }
}
```

### Todo Component

Displays and manages a todo list stored in a local file.
You can add, toggle, and delete items interactively.

```jsonc
{
  "type": "component",
  "component": {
    "type": "todo",
    "title": "My Todos",
    "data": {
      "source": "file",
      "path": "./todos.txt"
    }
  }
}
```

#### Todo File Format

Each line represents a todo item:
  - `-` means to do
  - `+` means done

Example:
```
 + learn HTMX
 + rise & grind
 - profit
```

## ⌨️ Keybindings

- Navigation:
    - `shift + ↑` / `shifht + K` - Move Up
    - `shift + ↓` / `shift + J` - Move Down
    - `shift + ←` / `shift + H` - Move Left
    - `shift + →` / `shift + L` - Move Right
    - `Left Click` - Focus Component Under Cursor
- Focused Componet Actions:
    - Mouse Wheel - Scroll
    - `PgUp` / `b` / `u` - Scroll up.
    - `PgDown` / `Space` / `d` - Scroll down.
    - `r` - Refresh Data Source 
- Todo List:
    - `a` - Add a new todo item (type, then `Enter` to save, `Esc` to cancel)
    - `Space` - Toggle done/undone for selected item
    - `d` / `Delete` / `Backspace` - Delete selected item

- Quit: `Ctrl+C`

### 💡 Future Ideas

- More component types (charts, gauges, lists).
- More data sources
- Data parsing/transformation (e.g., extracting values from JSON).
- Customizable themes/colors.
