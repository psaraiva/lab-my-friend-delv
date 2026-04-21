package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type RollResult struct {
	DiceName string `json:"dice_name"`
	Sides    int    `json:"sides"`
	Value    int    `json:"value"`
}

var nameRe = regexp.MustCompile(`^[a-zA-Z0-9]{1,25}$`)

var rootCmd = &cobra.Command{
	Use:   "app01 <dice-name>",
	Short: "Rolls a registered die and displays the result in a friendly way",
	Long: `app01 receives the name of a die previously registered in App03,
queries App02 to perform the roll, and displays the formatted result.

Example:
  app01 20sideddice
  app01 dragon`,
	Args: cobra.ExactArgs(1),
	RunE: runRoll,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRoll(_ *cobra.Command, args []string) error {
	name := args[0]

	if !nameRe.MatchString(name) {
		return fmt.Errorf("invalid name: must be 1 to 25 alphanumeric characters (a-Z, 0-9)")
	}

	result, err := fetchRoll("http://localhost:9022", name)
	if err != nil {
		return err
	}

	fmt.Println(formatResult(result))
	return nil
}

func fetchRoll(app02URL, name string) (*RollResult, error) {
	url := fmt.Sprintf("%s/roll/%s", strings.TrimRight(app02URL, "/"), name)

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("could not connect to the dice service (%s): %w", app02URL, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("dice '%s' not found — register it first using App03 (POST /dice)", name)
	case http.StatusServiceUnavailable:
		return nil, fmt.Errorf("dice service is currently unavailable, please try again in a moment")
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("unexpected response from the service (HTTP status %d)", resp.StatusCode)
	}

	var result RollResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to interpret service response: %w", err)
	}
	return &result, nil
}

func formatResult(r *RollResult) string {
	separator := strings.Repeat("-", 40)

	diceTypeName := fmt.Sprintf("D%d", r.Sides)

	var highlight string
	switch r.Value {
	case r.Sides:
		highlight = "  << MAXIMUM VALUE!"
	case 1:
		highlight = "  << MINIMUM VALUE!"
	}

	return fmt.Sprintf(`
%s
  Dice   : %s (%s)
  Sides  : %d
  Result --> %d%s
%s`,
		separator,
		r.DiceName,
		diceTypeName,
		r.Sides,
		r.Value,
		highlight,
		separator,
	)
}
