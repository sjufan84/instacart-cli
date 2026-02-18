package instacart

// Ingredient represents a recipe ingredient.
type Ingredient struct {
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	DisplayText string  `json:"displayText,omitempty"`
}

// Recipe represents a recipe to create on Instacart.
type Recipe struct {
	Title        string       `json:"title"`
	ImageURL     string       `json:"image_url,omitempty"`
	Author       string       `json:"author,omitempty"`
	Servings     int          `json:"servings,omitempty"`
	CookingTime  int          `json:"cooking_time,omitempty"`
	Instructions []string     `json:"instructions,omitempty"`
	Ingredients  []Ingredient `json:"ingredients"`
}

// LineItem represents a shopping list item.
type LineItem struct {
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	DisplayText string  `json:"displayText,omitempty"`
}

// ShoppingList represents a shopping list to create on Instacart.
type ShoppingList struct {
	Title        string     `json:"title"`
	ImageURL     string     `json:"image_url,omitempty"`
	ExpiresIn    int        `json:"expires_in,omitempty"`
	LineItems    []LineItem `json:"lineItems"`
	EnablePantry bool       `json:"enable_pantry,omitempty"`
}
