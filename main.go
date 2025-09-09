package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func showHelp() {
	fmt.Printf(`
Jira Worklogger v%s - Command-line tool for posting worklogs to Jira Cloud/Server

Usage: jira-worklogger [options]

Options:
  --help, -h             Show this help message and exit
  --version, -v          Show version information and exit
  --date DATE            Set the worklog date (format: YYYY-MM-DD, default: today)
  --entries ENTRIES      Specify time entries directly (e.g., "meetings=1h;support=30m")
                        Skips the interactive prompt when provided

Configuration:
  The tool looks for configuration in the following locations:
  1. Environment variable WORKLOG_CONFIG
  2. Current working directory (worklog_config.yaml)
  3. Executable directory
  4. User's home directory (~/.worklog_config.yaml)
  5. System-wide config (/etc/jira-worklogger/worklog_config.yaml)

Environment Variables:
  JIRA_BASE_URL   - Jira instance URL (e.g. https://company.atlassian.net)
  JIRA_EMAIL      - Your Jira email address
  JIRA_API_TOKEN  - Your Jira API token
  TIMEZONE        - Your timezone (e.g. Europe/London)
  JIRA_API_VERSION - API version (2 for Server, 3 for Cloud)
  LOG_LEVEL       - Logging verbosity (debug, info, warn, error)
`, Version)
	
	fmt.Println(`
Category Aliases:
  You can define category aliases in your config file under defaults.category_aliases
  These allow you to use shorthand names instead of typing full Jira issue keys.
  
  Example in worklog_config.yaml:
    defaults:
      category_aliases:
        meetings: "PROJ-123"
        support: "PROJ-456"
        docs: "PROJ-789"
        
  Usage:
    Time entries: meetings=1h; support=30m; docs=2h
`)
}

func promptUser(settings *Settings, epics map[string]Epic) (map[string]string, error) {
	reader := bufio.NewReader(os.Stdin)
	results := map[string]string{
		"date":    DefaultDateStr(),
		"entries": "",
	}

	fmt.Println("=== Jira Worklogger ===")

	// Show epics if available
	if len(epics) > 0 {
		fmt.Println("\nSuggested Epics (You Are Possibly Working On):")
		
		// Convert map to a slice for sorting
		epicsList := make([]Epic, 0, len(epics))
		for _, epic := range epics {
			epicsList = append(epicsList, epic)
		}
		
		// Display epics
		for i, epic := range epicsList {
			fmt.Printf("  %d. %s: %s\n", i+1, epic.Key, epic.Summary)
		}
		fmt.Println("\nTo log time to an epic, use its key in the time entries field.")
	}

	// Show category aliases if available
	if len(settings.CategoryAliases) > 0 {
		fmt.Println("\nAvailable Category Aliases:")
		
		// Get keys for sorting
		keys := make([]string, 0, len(settings.CategoryAliases))
		for k := range settings.CategoryAliases {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		// Display aliases
		for _, k := range keys {
			fmt.Printf("  %-10s -> %s\n", k, settings.CategoryAliases[k])
		}
	}

	// Prompt for date
	fmt.Printf("Date [YYYY-MM-DD] (default %s): ", results["date"])
	dateInput, _ := reader.ReadString('\n')
	dateInput = strings.TrimSpace(dateInput)
	if dateInput != "" {
		results["date"] = dateInput
	}

	// Prompt for time entries
	fmt.Print("Time entries (e.g., meetings=1h; munio=2h; PROJ-123=1.5h): ")
	entriesInput, _ := reader.ReadString('\n')
	results["entries"] = strings.TrimSpace(entriesInput)

	return results, nil
}

func main() {
	// Initialize command line options
	cmdLineOptions := map[string]string{
		"date":    "",
		"entries": "",
	}

	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		showHelp()
		return
	}
	
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Jira Worklogger v%s\n", Version)
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Commit: %s\n", Commit)
		return
	}

	// Parse command line arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			
			// Skip to next argument if this is the last one or next is another flag
			if i+1 >= len(os.Args) || strings.HasPrefix(os.Args[i+1], "--") {
				fmt.Printf("[error] No value provided for flag %s\n", arg)
				os.Exit(1)
			}
			
			// Get the value
			value := os.Args[i+1]
			i++ // Skip the next argument as it's a value
			
			// Store the value
			if _, exists := cmdLineOptions[key]; exists {
				cmdLineOptions[key] = value
			} else {
				fmt.Printf("[warn] Unknown flag: %s\n", arg)
			}
		}
	}

	// Load settings
	settings, err := LoadSettings()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := NewLogger(settings.LogLevel)
	logger.Info("Starting jira-worklogger with log level: %s", settings.LogLevel)

	// Check for placeholder API token
	if settings.JiraAPIToken == "YOUR_API_TOKEN" {
		logger.Error("Please update your API token in worklog_config.yaml - it's currently set to the placeholder value 'YOUR_API_TOKEN'")
		os.Exit(1)
	}

	// Fetch assigned issues and extract epics
	epics := make(map[string]Epic)
	issues, err := GetAssignedIssues(settings, logger)
	if err != nil {
		logger.Warn("Failed to load issues: %v", err)
	} else {
		epics = GetEpicsFromIssues(issues)
	}

	// Prepare user input (either from command-line or interactive prompts)
	var userInput map[string]string
	
	// Check if we have command-line parameters for non-interactive mode
	if cmdLineOptions["entries"] != "" {
		// Non-interactive mode
		userInput = map[string]string{
			"date":    cmdLineOptions["date"],
			"entries": cmdLineOptions["entries"],
		}
		logger.Info("Running in non-interactive mode with provided parameters")
	} else {
		// Interactive mode - prompt user for input
		userInput, err = promptUser(settings, epics)
		if err != nil {
			logger.Error("Failed to get user input: %v", err)
			os.Exit(1)
		}
	}

	// Parse date and create ISO timestamp
	dateStr := userInput["date"]
	if dateStr == "" {
		dateStr = DefaultDateStr()
	}
	
	startedISO, err := ISOStartForDate(dateStr, 17, 0, settings.Timezone, logger)
	if err != nil {
		logger.Error("Failed to parse date: %v", err)
		os.Exit(1)
	}

	// Parse time entries
	entries, err := ParseTimeEntries(userInput["entries"], settings.CategoryAliases, logger)
	if err != nil {
		logger.Error("Failed to parse time entries: %v", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		logger.Info("No time entries to post. Exiting.")
		return
	}

	// Post worklogs
	var successes []WorklogResult
	var failures []WorklogResult

	for _, entry := range entries {
		result := PostWorklog(settings, logger, entry.Issue, entry.Seconds, startedISO)
		if result.Success {
			successes = append(successes, result)
		} else {
			failures = append(failures, result)
		}
	}

	// Report results
	if len(successes) > 0 {
		// Calculate total time
		totalSeconds := 0
		for _, success := range successes {
			totalSeconds += success.Seconds
		}
		hours := totalSeconds / 3600
		minutes := (totalSeconds % 3600) / 60

		logger.Info("Posted %d worklogs (%dh %dm).", len(successes), hours, minutes)
		for _, success := range successes {
			h := success.Seconds / 3600
			m := (success.Seconds % 3600) / 60
			logger.Info("  - %s: %dh%dm", success.Issue, h, m)
		}
	}

	if len(failures) > 0 {
		logger.Error("Some entries failed:")
		for _, failure := range failures {
			// Limit the response body to 500 characters
			responseBody := failure.Body
			if len(responseBody) > 500 {
				responseBody = responseBody[:500]
			}
			logger.Error("  - %s: HTTP %d\n    %s", failure.Issue, failure.Code, responseBody)
		}
		os.Exit(1)
	}
}
