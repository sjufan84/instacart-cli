package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sjufan84/instacart-cli/pkg/instacart"
	"github.com/spf13/cobra"
)

var (
	recipeServings    int
	recipeAuthor      string
	recipeCookingTime int
	recipeImageURL    string
	recipeIngredients []string
	recipeInstructions []string
	recipeFile        string
	recipeJSON        bool
	recipeOpen        bool
	recipeQuiet       bool
)

var recipeCmd = &cobra.Command{
	Use:   "recipe [title]",
	Short: "Create a recipe page on Instacart",
	Long:  "Create a recipe page on Instacart Marketplace with ingredients that users can add to their cart.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRecipe,
}

func init() {
	rootCmd.AddCommand(recipeCmd)
	recipeCmd.Flags().IntVarP(&recipeServings, "servings", "n", 0, "Number of servings")
	recipeCmd.Flags().StringVarP(&recipeAuthor, "author", "a", "", "Recipe author")
	recipeCmd.Flags().IntVarP(&recipeCookingTime, "time", "t", 0, "Cooking time in minutes")
	recipeCmd.Flags().StringVar(&recipeImageURL, "image", "", "Image URL (500x500)")
	recipeCmd.Flags().StringArrayVarP(&recipeIngredients, "ingredient", "i", nil, `Ingredient as "name,qty,unit" (repeatable)`)
	recipeCmd.Flags().StringArrayVarP(&recipeInstructions, "instruction", "s", nil, "Instruction step (repeatable, ordered)")
	recipeCmd.Flags().StringVarP(&recipeFile, "file", "f", "", "Load recipe from JSON file (- for stdin)")
	recipeCmd.Flags().BoolVarP(&recipeJSON, "json", "j", false, "Output as JSON")
	recipeCmd.Flags().BoolVarP(&recipeOpen, "open", "o", false, "Open URL in browser")
	recipeCmd.Flags().BoolVarP(&recipeQuiet, "quiet", "q", false, "Only output the URL")
}

func runRecipe(cmd *cobra.Command, args []string) error {
	var recipe instacart.Recipe

	if recipeFile != "" {
		var data []byte
		var err error
		if recipeFile == "-" {
			data, err = os.ReadFile("/dev/stdin")
		} else {
			data, err = os.ReadFile(recipeFile)
		}
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		if err := json.Unmarshal(data, &recipe); err != nil {
			return fmt.Errorf("parse JSON: %w", err)
		}
	} else {
		if len(args) == 0 {
			return fmt.Errorf("recipe title is required (or use --file)")
		}
		recipe.Title = args[0]
		recipe.Servings = recipeServings
		recipe.Author = recipeAuthor
		recipe.CookingTime = recipeCookingTime
		recipe.ImageURL = recipeImageURL
		recipe.Instructions = recipeInstructions

		for _, raw := range recipeIngredients {
			ing, err := parseIngredient(raw)
			if err != nil {
				return err
			}
			recipe.Ingredients = append(recipe.Ingredients, ing)
		}
	}

	if len(recipe.Ingredients) == 0 {
		return fmt.Errorf("at least one ingredient is required")
	}

	client := instacart.NewClient(endpoint, getToken())
	url, err := client.CreateRecipe(context.Background(), recipe)
	if err != nil {
		return fmt.Errorf("create recipe: %w", err)
	}

	return printResult("Recipe", recipe.Title, url, recipeJSON, recipeQuiet, recipeOpen)
}

func parseIngredient(raw string) (instacart.Ingredient, error) {
	parts := strings.SplitN(raw, ",", 3)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return instacart.Ingredient{}, fmt.Errorf("invalid ingredient: %q", raw)
	}

	ing := instacart.Ingredient{Name: strings.TrimSpace(parts[0])}
	if len(parts) >= 2 {
		q, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err == nil {
			ing.Quantity = q
		}
	}
	if len(parts) >= 3 {
		ing.Unit = strings.TrimSpace(parts[2])
	}
	return ing, nil
}
