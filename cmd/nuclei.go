package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

// Structure to map the JSON data
type TechData struct {
	Host string   `json:"host"`
	Tech []string `json:"tech"`
}

// nucleiCmd represents the nuclei command
var nucleiCmd = &cobra.Command{
	Use:   "nuclei",
	Short: "Run Nuclei scans on multiple hosts in parallel, filtering by technology stack.",
	Long: `The 'nuclei' command processes a JSON file containing hosts and their technology stacks, runs Nuclei scans in parallel based on the specified technologies, and allows appending output to a file. 
You can customize the Nuclei command using a template, and control the level of parallelism with the provided flags.

Examples:
  vulntechx nuclei --file httpxjson-output.json --cmd "nuclei -duc -nc -t ~/nucleihub-templates -tags {tech} -es unknown,info,low" --parallel 10 --process --append-output nuclei-output.txt

  or
  vulntechx nuclei --file httpxjson-output.json --cmd "nuclei -duc -nc -t ~/nucleihub-templates -tc {tech} -es unknown,info,low" --parallel 10 --process --append-output nuclei-output.txt

  or
  vulntechx nuclei --file httpxjson-output.json --cmd "nuclei -duc -nc -t ~/nucleihub-templates -tc {tech} -es unknown,info,low" --parallel 10 --process --exclude-tech "hsts,bootstrap" --append-output nuclei-output.txt
`,
	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := cmd.Flags().GetString("file")
		nucleiCmdStr, _ := cmd.Flags().GetString("cmd")
		verbose, _ := cmd.Flags().GetBool("verbose")
		process, _ := cmd.Flags().GetBool("process")
		parallel, _ := cmd.Flags().GetInt("parallel")
		appendOutput, _ := cmd.Flags().GetString("append-output")
		excludeTech, _ := cmd.Flags().GetString("exclude-tech")

		if fileName == "" {
			fmt.Println("Usage: nucleitechx -file <file> -cmd <nuclei command> -parallel <number of parallel processes> --append-output <output file>")
			os.Exit(1)
		}

		excludeList := strings.Split(excludeTech, ",")
		for i := range excludeList {
			excludeList[i] = strings.TrimSpace(strings.ToLower(excludeList[i]))
		}

		file, err := os.Open(fileName)
		if err != nil {
			fmt.Printf("Error opening file: %s\n", err)
			os.Exit(1)
		}
		defer file.Close()

		// Open the output file for appending if the --append-output flag is specified
		var outputFile *os.File
		if appendOutput != "" {
			outputFile, err = os.OpenFile(appendOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Error opening output file: %s\n", err)
				os.Exit(1)
			}
			defer outputFile.Close()
		}

		decoder := json.NewDecoder(file)
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, parallel) // Limit the number of parallel executions

		for {
			var techData TechData
			if err := decoder.Decode(&techData); err == io.EOF {
				break
			} else if err != nil {
				fmt.Printf("Error decoding JSON: %s\n", err)
				os.Exit(1)
			}

			// Skip processing if tech is nil
			if techData.Tech == nil {
				if verbose {
					fmt.Printf("Skipping URL with tech field as null: %s\n", techData.Host)
				}
				continue
			}

			wg.Add(1)
			semaphore <- struct{}{} // Acquire a semaphore
			go func(techData TechData) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release the semaphore

				// Process tech field
				var techs []string
				for _, t := range techData.Tech {
					parts := strings.SplitN(t, ":", 2)
					if len(parts) > 0 {
						tech := strings.TrimSpace(parts[0])
						// Ignore technologies with spaces
						if !strings.Contains(tech, " ") && !contains(excludeList, strings.ToLower(tech)) {
							techs = append(techs, tech)
						}
					}
				}

				// Skip if techs is empty
				if len(techs) == 0 {
					fmt.Printf("SKIPPED: %s tech is empty\n", techData.Host)
					return
				}

				tech := strings.ToLower(strings.Join(techs, ","))

				var cmdStr string
				if strings.Contains(nucleiCmdStr, "-tc") {
					// Modify to use the -tc format
					var conditions []string
					for _, t := range techs {
						conditions = append(conditions, fmt.Sprintf("contains(to_lower(name),'%s')", strings.ToLower(t)))
					}
					cmdStr = strings.Replace(nucleiCmdStr, "{tech}", fmt.Sprintf("\"%s\"", strings.Join(conditions, " || ")), -1)
				} else if strings.Contains(nucleiCmdStr, "-tags") {
					// Use the -tags format as-is
					cmdStr = strings.Replace(nucleiCmdStr, "{tech}", tech, -1)
				}

				if process {
					fmt.Printf("Running Nuclei: [echo \"%s\" | %s]\n", techData.Host, cmdStr)
				}

				// Run the nuclei command
				cmd := exec.Command("sh", "-c", cmdStr)
				cmd.Stdin = strings.NewReader(techData.Host)
				stdoutPipe, _ := cmd.StdoutPipe()
				stderrPipe, _ := cmd.StderrPipe()

				if err := cmd.Start(); err != nil {
					if verbose {
						fmt.Printf("Error starting nuclei command: %s\n", err)
					}
					return
				}

				// Handle the output
				scanner := bufio.NewScanner(io.MultiReader(stdoutPipe, stderrPipe))
				for scanner.Scan() {
					line := scanner.Text()
					fmt.Println(line)

					// Check if the line starts with three sets of square brackets
					parts := strings.Fields(line)
					if len(parts) >= 3 && strings.HasPrefix(parts[0], "[") && strings.HasPrefix(parts[1], "[") && strings.HasPrefix(parts[2], "[") {
						if appendOutput != "" {
							// Append the filtered output line to the specified file
							if _, err := outputFile.WriteString(line + "\n"); err != nil && verbose {
								fmt.Printf("Error writing to output file: %s\n", err)
							}
						}
					}
				}

				if err := cmd.Wait(); err != nil && verbose {
					fmt.Printf("Error waiting for nuclei command: %s\n", err)
				}

			}(techData)
		}

		wg.Wait() // Wait for all goroutines to finish
	},
}

// Utility function to check if a tech is in the exclusion list
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(nucleiCmd)

	nucleiCmd.Flags().String("file", "", "Path to the JSON file")
	nucleiCmd.Flags().String("cmd", "", "The nuclei command template")
	nucleiCmd.Flags().Bool("verbose", false, "Enable verbose output for debugging purposes.")
	nucleiCmd.Flags().Bool("process", false, "Show which URL is running on Nuclei.")
	nucleiCmd.Flags().Int("parallel", 50, "Number of parallel processes")
	nucleiCmd.Flags().String("append-output", "", "File path to append output")
	nucleiCmd.Flags().String("exclude-tech", "", "Comma-separated list of technologies to exclude")
}
