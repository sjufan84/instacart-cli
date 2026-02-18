package instacart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const DefaultEndpoint = "https://mcp.dev.instacart.tools/mcp"

type Client struct {
	Endpoint   string
	Token      string // Bearer token for auth
	HTTPClient *http.Client
}

func NewClient(endpoint, token string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return &Client{
		Endpoint: endpoint,
		Token:    token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type rpcRequest struct {
	JSONRPC string    `json:"jsonrpc"`
	Method  string    `json:"method"`
	Params  rpcParams `json:"params"`
	ID      string    `json:"id"`
}

type rpcParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type rpcResponse struct {
	Result map[string]any `json:"result"`
	Error  *rpcError      `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CreateRecipe creates a recipe page on Instacart and returns the shareable URL.
func (c *Client) CreateRecipe(ctx context.Context, recipe Recipe) (string, error) {
	ingredients := make([]map[string]any, 0, len(recipe.Ingredients))
	for _, ing := range recipe.Ingredients {
		m := map[string]any{"name": ing.Name}
		if ing.Quantity > 0 {
			m["quantity"] = ing.Quantity
		}
		if ing.Unit != "" {
			m["unit"] = ing.Unit
		}
		if ing.DisplayText != "" {
			m["displayText"] = ing.DisplayText
		}
		ingredients = append(ingredients, m)
	}

	args := map[string]any{
		"title":       recipe.Title,
		"ingredients": ingredients,
	}
	if recipe.Author != "" {
		args["author"] = recipe.Author
	}
	if recipe.Servings > 0 {
		args["servings"] = recipe.Servings
	}
	if recipe.CookingTime > 0 {
		args["cooking_time"] = recipe.CookingTime
	}
	if recipe.ImageURL != "" {
		args["image_url"] = recipe.ImageURL
	}
	if len(recipe.Instructions) > 0 {
		args["instructions"] = recipe.Instructions
	}

	result, err := c.callTool(ctx, "create-recipe", args)
	if err != nil {
		return "", err
	}
	return extractURL(result), nil
}

// CreateShoppingList creates a shopping list on Instacart and returns the shareable URL.
func (c *Client) CreateShoppingList(ctx context.Context, list ShoppingList) (string, error) {
	lineItems := make([]map[string]any, 0, len(list.LineItems))
	for _, item := range list.LineItems {
		m := map[string]any{"name": item.Name}
		if item.Quantity > 0 {
			m["quantity"] = item.Quantity
		}
		if item.Unit != "" {
			m["unit"] = item.Unit
		}
		if item.DisplayText != "" {
			m["displayText"] = item.DisplayText
		}
		lineItems = append(lineItems, m)
	}

	args := map[string]any{
		"title":     list.Title,
		"lineItems": lineItems,
	}
	if list.ImageURL != "" {
		args["image_url"] = list.ImageURL
	}
	if list.ExpiresIn > 0 {
		args["expires_in"] = list.ExpiresIn
	}
	if list.EnablePantry {
		if args["landingPageConfiguration"] == nil {
			args["landingPageConfiguration"] = map[string]any{}
		}
		args["landingPageConfiguration"].(map[string]any)["enablePantryItems"] = true
	}

	result, err := c.callTool(ctx, "create-shopping-list", args)
	if err != nil {
		return "", err
	}
	return extractURL(result), nil
}

func (c *Client) callTool(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
	payload := rpcRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		ID:      fmt.Sprintf("instacart-cli-%d", time.Now().UnixNano()),
		Params: rpcParams{
			Name:      toolName,
			Arguments: args,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal rpc payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("instacart returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error (%d): %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func extractURL(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	keys := []string{"url", "share_url", "shareUrl", "shopping_list_url", "shoppingListUrl"}
	for _, key := range keys {
		if v, ok := data[key].(string); ok && strings.HasPrefix(v, "http") {
			return v
		}
	}

	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`https?://[^\s\"']+`)
	return re.FindString(string(b))
}
