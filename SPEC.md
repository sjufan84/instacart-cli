# Instacart CLI — Go Project Spec

## Overview
A lightweight Go CLI that creates recipes and shopping lists on Instacart via their MCP endpoint. Two commands, zero config needed.

## Why
No standalone Instacart CLI exists. This fills the gap for developers, AI agents, and tools like mealplan that need programmatic Instacart access.

## Tech Stack
- **Language:** Go 1.22+
- **CLI Framework:** Cobra
- **HTTP:** Standard `net/http` (no dependencies beyond Cobra)
- **Protocol:** JSON-RPC 2.0 to Instacart's MCP endpoint

## Project Structure
```
instacart-cli/
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go          # Root command, global flags
│   ├── recipe.go        # create-recipe command
│   └── list.go          # create-shopping-list command
├── pkg/
│   └── instacart/       # Reusable client library (importable by other Go projects)
│       ├── client.go    # MCP HTTP client
│       └── types.go     # Recipe, ShoppingList, Ingredient types
└── README.md
```

**Key design choice:** The `pkg/instacart/` package is the real product. The CLI is just a thin wrapper. This means mealplan (or anyone) can do:
```go
import "github.com/sjufan84/instacart-cli/pkg/instacart"
```

## Data Types

```go
// pkg/instacart/types.go

type Ingredient struct {
    Name        string  `json:"name"`
    Quantity    float64 `json:"quantity,omitempty"`
    Unit        string  `json:"unit,omitempty"`
    DisplayText string  `json:"displayText,omitempty"`
}

type Recipe struct {
    Title        string       `json:"title"`
    ImageURL     string       `json:"image_url,omitempty"`
    Author       string       `json:"author,omitempty"`
    Servings     int          `json:"servings,omitempty"`
    CookingTime  int          `json:"cooking_time,omitempty"`
    Instructions []string     `json:"instructions,omitempty"`
    Ingredients  []Ingredient `json:"ingredients"`
}

type LineItem struct {
    Name        string  `json:"name"`
    Quantity    float64 `json:"quantity,omitempty"`
    Unit        string  `json:"unit,omitempty"`
    DisplayText string  `json:"displayText,omitempty"`
}

type ShoppingList struct {
    Title      string     `json:"title"`
    ImageURL   string     `json:"image_url,omitempty"`
    ExpiresIn  int        `json:"expires_in,omitempty"` // days
    LineItems  []LineItem `json:"lineItems"`
}
```

## MCP Client

```go
// pkg/instacart/client.go

const DefaultEndpoint = "https://mcp.dev.instacart.tools/mcp"

type Client struct {
    Endpoint string
    HTTP     *http.Client
}

// JSON-RPC 2.0 request format
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"` // "2.0"
    Method  string      `json:"method"`  // "tools/call"
    Params  MCPParams   `json:"params"`
    ID      int         `json:"id"`
}

type MCPParams struct {
    Name      string      `json:"name"`      // "create-recipe" or "create-shopping-list"
    Arguments interface{} `json:"arguments"`
}

func (c *Client) CreateRecipe(r Recipe) (string, error)        // returns Instacart URL
func (c *Client) CreateShoppingList(s ShoppingList) (string, error) // returns Instacart URL
```

## CLI Commands

### `instacart recipe`
Create a recipe page on Instacart.

```bash
# Inline ingredients
instacart recipe "Chicken Tacos" \
  --servings 4 \
  --author "Dave" \
  --time 25 \
  --ingredient "chicken breast,1.5,lb" \
  --ingredient "taco shells,8,piece" \
  --ingredient "avocado,2" \
  --instruction "Season and grill chicken" \
  --instruction "Warm taco shells" \
  --instruction "Assemble and serve"

# From JSON file
instacart recipe --file taco-recipe.json

# From stdin (pipe from other tools)
cat recipe.json | instacart recipe --file -
```

**Output:**
```
✅ Recipe "Chicken Tacos" created!
🔗 https://customers.dev.instacart.tools/store/recipes/12345
```

**Flags:**
| Flag | Short | Description |
|---|---|---|
| `--ingredient` | `-i` | Ingredient as "name,qty,unit" (repeatable) |
| `--instruction` | `-s` | Instruction step (repeatable, ordered) |
| `--servings` | `-n` | Number of servings |
| `--author` | `-a` | Recipe author |
| `--time` | `-t` | Cooking time in minutes |
| `--image` | | Image URL (500x500) |
| `--file` | `-f` | Load recipe from JSON file (- for stdin) |
| `--json` | `-j` | Output as JSON instead of human-friendly |
| `--open` | `-o` | Open URL in browser after creation |

### `instacart list`
Create a shopping list on Instacart.

```bash
# Inline items
instacart list "Weekly Groceries" \
  --item "milk,1,gallon" \
  --item "eggs,12" \
  --item "bread,1,loaf" \
  --item "bananas,6"

# From JSON
instacart list --file groceries.json

# With expiration
instacart list "Party Supplies" --expires 7 --item "chips,3,bag"
```

**Output:**
```
✅ Shopping list "Weekly Groceries" created!
🔗 https://customers.dev.instacart.tools/store/shopping_lists/12346
```

**Flags:**
| Flag | Short | Description |
|---|---|---|
| `--item` | `-i` | Line item as "name,qty,unit" (repeatable) |
| `--expires` | `-e` | Expiration in days |
| `--image` | | Image URL (500x500) |
| `--file` | `-f` | Load list from JSON file (- for stdin) |
| `--json` | `-j` | Output as JSON |
| `--open` | `-o` | Open URL in browser |
| `--pantry` | | Enable pantry item detection |

### Global Flags
| Flag | Description |
|---|---|
| `--endpoint` | Override MCP endpoint URL |
| `--json` | JSON output for all commands |
| `--quiet` | Only output the URL (for scripting) |

## JSON File Formats

**Recipe:**
```json
{
  "title": "Chicken Tacos",
  "servings": 4,
  "cooking_time": 25,
  "author": "Dave",
  "ingredients": [
    {"name": "chicken breast", "quantity": 1.5, "unit": "lb"},
    {"name": "taco shells", "quantity": 8, "unit": "piece"}
  ],
  "instructions": [
    "Season and grill chicken.",
    "Warm taco shells.",
    "Assemble and serve."
  ]
}
```

**Shopping list:**
```json
{
  "title": "Weekly Groceries",
  "expires_in": 7,
  "lineItems": [
    {"name": "milk", "quantity": 1, "unit": "gallon"},
    {"name": "eggs", "quantity": 12}
  ]
}
```

## Piping & Scripting Examples

```bash
# Quiet mode for scripting — just the URL
URL=$(instacart list "Groceries" --item "milk,1,gallon" --quiet)

# Pipe from jq or other tools
cat meal-plan.json | jq '.shopping_list' | instacart list --file -

# JSON output for programmatic use
instacart recipe "Tacos" -i "chicken,1,lb" --json | jq '.url'
```

## Integrating with Mealplan CLI

In mealplan's `go.mod`:
```
require github.com/sjufan84/instacart-cli v1.0.0
```

In mealplan's order command:
```go
import "github.com/sjufan84/instacart-cli/pkg/instacart"

client := instacart.NewClient(instacart.DefaultEndpoint)
url, err := client.CreateShoppingList(instacart.ShoppingList{
    Title:     "Week of Feb 17",
    LineItems: convertItems(shoppingList.Items),
})
```

## Implementation Notes

- Parse `--ingredient "name,qty,unit"` by splitting on commas; qty and unit are optional
- For `--file -`, read from `os.Stdin`
- Extract URL from MCP response text (parse "View and share..." line)
- `--open` uses `github.com/pkg/browser` or `os/exec` to call `xdg-open`/`open`/`start`
- Keep the client stateless — no auth, no config file needed
- Error messages should be helpful: show the MCP error if Instacart returns one

## Distribution
- `go install github.com/sjufan84/instacart-cli@latest`
- GitHub releases with pre-built binaries (use GoReleaser)
- ClawHub skill for AI agents
- Future: Homebrew tap
