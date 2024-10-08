package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// TechInfo represents the structure of the output JSON.
type TechInfo struct {
	Host  string   `json:"host"`
	Count int      `json:"count"`
	Tech  []string `json:"tech"`
}

// httpxjsonCmd represents the httpxjson command
var httpxjsonCmd = &cobra.Command{
	Use:   "httpxjson",
	Short: "Parse httpx output to extract hosts and technologies, and format as JSON.",
	Long: `The 'httpxjson' command processes httpx output from standard input, extracts host information and associated technologies, and outputs the result as a JSON structure.

Examples:
  cat httpx.txt | vulntechx httpxjson -o httpxjson-output.json
`,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, _ := cmd.Flags().GetString("output")

		var file *os.File
		var err error

		// If an output file is specified, open it for writing
		if outputFile != "" {
			file, err = os.Create(outputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				return
			}
			defer file.Close()
		}

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			line := scanner.Text()

			// Split the line by space to separate URL and technologies
			parts := strings.SplitN(line, " [", 2)
			if len(parts) < 2 {
				continue // Skip lines that don't match the expected format
			}

			// Extract host and technology list
			host := strings.TrimSpace(parts[0])
			techList := strings.TrimSuffix(parts[1], "]")

			var techArray []string
			count := 0

			if techList != "" {
				techArray = strings.Split(techList, ",")
				count = len(techArray)
			}

			// Create the output structure
			techInfo := TechInfo{
				Host:  host,
				Count: count,
				Tech:  techArray,
			}

			// Convert to JSON and print
			jsonData, err := json.MarshalIndent(techInfo, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				continue
			}

			fmt.Println(string(jsonData))

			// Write to file if specified, otherwise print to stdout
			if outputFile != "" {
				_, err := file.WriteString(string(jsonData) + "\n")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(httpxjsonCmd)
	httpxjsonCmd.Flags().StringP("output", "o", "", "Output file to save JSON results")
}
