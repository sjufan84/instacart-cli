package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sjufan84/instacart-cli/pkg/instacart"
	"github.com/spf13/cobra"
)

var (
	listItems    []string
	listExpires  int
	listImageURL string
	listFile     string
	listJSON     bool
	listOpen     bool
	listQuiet    bool
	listPantry   bool
)

var listCmd = &cobra.Command{
	Use:   "list [title]",
	Short: "Create a shopping list on Instacart",
	Long:  "Create a shopping list on Instacart Marketplace that users can shop from.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringArrayVarP(&listItems, "item", "i", nil, `Line item as "name,qty,unit" (repeatable)`)
	listCmd.Flags().IntVarP(&listExpires, "expires", "e", 0, "Expiration in days")
	listCmd.Flags().StringVar(&listImageURL, "image", "", "Image URL (500x500)")
	listCmd.Flags().StringVarP(&listFile, "file", "f", "", "Load list from JSON file (- for stdin)")
	listCmd.Flags().BoolVarP(&listJSON, "json", "j", false, "Output as JSON")
	listCmd.Flags().BoolVarP(&listOpen, "open", "o", false, "Open URL in browser")
	listCmd.Flags().BoolVarP(&listQuiet, "quiet", "q", false, "Only output the URL")
	listCmd.Flags().BoolVar(&listPantry, "pantry", false, "Enable pantry item detection")
}

func runList(cmd *cobra.Command, args []string) error {
	var list instacart.ShoppingList

	if listFile != "" {
		var data []byte
		var err error
		if listFile == "-" {
			data, err = os.ReadFile("/dev/stdin")
		} else {
			data, err = os.ReadFile(listFile)
		}
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		if err := json.Unmarshal(data, &list); err != nil {
			return fmt.Errorf("parse JSON: %w", err)
		}
	} else {
		if len(args) == 0 {
			return fmt.Errorf("list title is required (or use --file)")
		}
		list.Title = args[0]
		list.ExpiresIn = listExpires
		list.ImageURL = listImageURL
		list.EnablePantry = listPantry

		for _, raw := range listItems {
			item, err := parseLineItem(raw)
			if err != nil {
				return err
			}
			list.LineItems = append(list.LineItems, item)
		}
	}

	if len(list.LineItems) == 0 {
		return fmt.Errorf("at least one item is required")
	}

	client := instacart.NewClient(endpoint, getToken())
	url, err := client.CreateShoppingList(context.Background(), list)
	if err != nil {
		return fmt.Errorf("create shopping list: %w", err)
	}

	return printResult("Shopping list", list.Title, url, listJSON, listQuiet, listOpen)
}

func parseLineItem(raw string) (instacart.LineItem, error) {
	parts := strings.SplitN(raw, ",", 3)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return instacart.LineItem{}, fmt.Errorf("invalid item: %q", raw)
	}

	item := instacart.LineItem{Name: strings.TrimSpace(parts[0])}
	if len(parts) >= 2 {
		q := strings.TrimSpace(parts[1])
		if q != "" {
			var f float64
			fmt.Sscanf(q, "%f", &f)
			item.Quantity = f
		}
	}
	if len(parts) >= 3 {
		item.Unit = strings.TrimSpace(parts[2])
	}
	return item, nil
}
