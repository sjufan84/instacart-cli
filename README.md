# instacart-cli 🛒

A Go CLI and reusable library for creating recipes and shopping lists on Instacart.

## Install

```bash
go install github.com/sjufan84/instacart-cli@latest
```

## Usage

### Create a recipe
```bash
instacart recipe "Chicken Tacos" \
  --servings 4 \
  --ingredient "chicken breast,1.5,lb" \
  --ingredient "taco shells,8,piece" \
  --instruction "Season and grill chicken" \
  --instruction "Assemble and serve"
```

### Create a shopping list
```bash
instacart list "Weekly Groceries" \
  --item "milk,1,gallon" \
  --item "eggs,12" \
  --item "bread,1,loaf"
```

### From JSON files
```bash
instacart recipe --file recipe.json
instacart list --file groceries.json
cat recipe.json | instacart recipe --file -
```

## As a Go Library

```go
import "github.com/sjufan84/instacart-cli/pkg/instacart"

client := instacart.NewClient(instacart.DefaultEndpoint)
url, err := client.CreateShoppingList(ctx, instacart.ShoppingList{
    Title:     "Weekly Groceries",
    LineItems: []instacart.LineItem{{Name: "milk", Quantity: 1, Unit: "gallon"}},
})
```

## Documentation

See [SPEC.md](SPEC.md) for the full project specification.

## License

MIT
