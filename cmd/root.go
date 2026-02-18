package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var endpoint string
var token string

var rootCmd = &cobra.Command{
	Use:   "instacart",
	Short: "Create recipes and shopping lists on Instacart",
	Long:  "A CLI for creating recipes and shopping lists on Instacart Marketplace via their MCP endpoint.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "Override MCP endpoint URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Bearer token for authentication (or set INSTACART_MCP_TOKEN)")
}

func Execute() error {
	return rootCmd.Execute()
}

func getToken() string {
	if token != "" {
		return token
	}
	if t := os.Getenv("INSTACART_MCP_TOKEN"); t != "" {
		return t
	}
	return ""
}

func printResult(kind, title, url string, asJSON, quiet, open bool) error {
	if quiet {
		fmt.Println(url)
	} else if asJSON {
		out := map[string]string{
			"kind":  kind,
			"title": title,
			"url":   url,
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("✅ %s %q created!\n🔗 %s\n", kind, title, url)
	}

	if open && url != "" {
		openBrowser(url)
	}
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
