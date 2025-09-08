package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TimeEntry represents a parsed time entry
type TimeEntry struct {
	Issue   string
	Seconds int
}

// DefaultDateStr returns the current date in YYYY-MM-DD format
func DefaultDateStr() string {
	now := time.Now()
	return fmt.Sprintf("%04d-%02d-%02d", now.Year(), now.Month(), now.Day())
}

// ParseDate parses a YYYY-MM-DD formatted date string
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		dateStr = DefaultDateStr()
	}
	
	return time.Parse("2006-01-02", dateStr)
}

// ISOStartForDate creates an ISO8601 timestamp for the given date with the specified hour and minute
func ISOStartForDate(dateStr string, hour, minute int, timezone string, logger *Logger) (string, error) {
	date, err := ParseDate(dateStr)
	if err != nil {
		return "", err
	}

	// Get timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		logger.Warn("Could not load timezone %s: %v, using system default", timezone, err)
		loc = time.Local
	}

	// Set time components
	date = time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, loc)
	
	// Format with timezone offset
	return date.Format("2006-01-02T15:04:05.000-0700"), nil
}

// ToTimeSpentSeconds converts a time string to seconds
// Accepts formats like "1.5h", "90m", "1:30", etc.
func ToTimeSpentSeconds(timeStr string) (int, error) {
	if timeStr = strings.TrimSpace(strings.ToLower(timeStr)); timeStr == "" {
		return 0, nil
	}

	// Check for "Xh Ym" format
	hoursAndMinutes := regexp.MustCompile(`^(\d+(\.\d+)?)\s*h\s*(\d+)\s*m$`)
	if matches := hoursAndMinutes.FindStringSubmatch(timeStr); len(matches) >= 4 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		minutes, _ := strconv.Atoi(matches[3])
		return int(hours*3600) + minutes*60, nil
	}

	// Check for "Xh" format
	hours := regexp.MustCompile(`^(\d+(\.\d+)?)\s*h$`)
	if matches := hours.FindStringSubmatch(timeStr); len(matches) >= 2 {
		hours, _ := strconv.ParseFloat(matches[1], 64)
		return int(hours * 3600), nil
	}

	// Check for "Ym" format
	minutes := regexp.MustCompile(`^(\d+)\s*m$`)
	if matches := minutes.FindStringSubmatch(timeStr); len(matches) >= 2 {
		minutes, _ := strconv.Atoi(matches[1])
		return minutes * 60, nil
	}

	// Check for "X:Y" format (hours:minutes)
	hoursMinutes := regexp.MustCompile(`^(\d+):(\d+)$`)
	if matches := hoursMinutes.FindStringSubmatch(timeStr); len(matches) >= 3 {
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		return hours*3600 + minutes*60, nil
	}

	// Try to parse as a decimal number of hours
	hoursFloat, parseErr := strconv.ParseFloat(timeStr, 64)
	if parseErr == nil {
		return int(hoursFloat * 3600), nil
	}

	return 0, fmt.Errorf("unable to parse time: %s", timeStr)
}

// ParseTimeEntries parses a time entry string into issue keys and durations
func ParseTimeEntries(entriesStr string, aliases map[string]string, logger *Logger) ([]TimeEntry, error) {
	entries := []TimeEntry{}
	if entriesStr == "" {
		return entries, nil
	}

	// Split entries by semicolon or newline
	items := []string{}
	for _, line := range strings.Split(entriesStr, "\n") {
		for _, part := range strings.Split(line, ";") {
			if part = strings.TrimSpace(part); part != "" {
				items = append(items, part)
			}
		}
	}

	for _, item := range items {
		// Extract issue and time value
		var issue, timeValue string

		// Format: ISSUE=TIME
		if parts := strings.SplitN(item, "=", 2); len(parts) == 2 {
			issue = strings.TrimSpace(parts[0])
			timeValue = strings.TrimSpace(parts[1])
		} else if parts := strings.Fields(item); len(parts) >= 2 {
			// Format: ISSUE TIME
			issue = parts[0]
			timeValue = parts[1]
		} else if len(parts) == 1 {
			// Only issue key, treat as zero
			issue = parts[0]
			timeValue = "0"
		} else {
			continue
		}

		// Remove trailing colon if present (may appear when clicking on epics)
		issue = strings.TrimSuffix(issue, ":")

		// Check for category alias
		if aliasValue, ok := aliases[strings.ToLower(issue)]; ok {
			logger.Info("Using alias '%s' -> %s", issue, aliasValue)
			issue = aliasValue
		}

		// Parse time spent
		seconds, err := ToTimeSpentSeconds(timeValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time for entry %s: %v", item, err)
		}

		if issue != "" && seconds > 0 {
			entries = append(entries, TimeEntry{Issue: issue, Seconds: seconds})
		}
	}

	return entries, nil
}
