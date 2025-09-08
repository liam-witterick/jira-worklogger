package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Epic represents a Jira epic
type Epic struct {
	Key     string
	Summary string
	Type    string
}

// WorklogResult represents the result of posting a worklog
type WorklogResult struct {
	Issue   string
	Seconds int
	Success bool
	Code    int
	Body    string
}

// BuildWorklogPayload builds a worklog payload for the Jira API
func BuildWorklogPayload(startedISO string, seconds int, apiVersion string) map[string]interface{} {
	// Don't include comments as per user request
	if apiVersion == "3" {
		// Jira Cloud API v3 format
		return map[string]interface{}{
			"started":         startedISO,
			"timeSpentSeconds": seconds,
		}
	} else {
		// Jira Server/DC API v2 format
		return map[string]interface{}{
			"started":         startedISO,
			"timeSpentSeconds": seconds,
		}
	}
}

// PostWorklog posts a worklog to a Jira issue
func PostWorklog(settings *Settings, logger *Logger, issue string, seconds int, startedISO string) WorklogResult {
	result := WorklogResult{
		Issue:   issue,
		Seconds: seconds,
		Success: false,
	}

	if seconds <= 0 {
		result.Success = true
		result.Body = "Skipped zero seconds"
		return result
	}

	// Create authentication header
	auth := fmt.Sprintf("%s:%s", settings.JiraEmailOrUser, settings.JiraAPIToken)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// Create worklog URL
	baseURL := strings.TrimSuffix(settings.JiraBaseURL, "/")
	issueURL := fmt.Sprintf("%s/rest/api/%s/issue/%s/worklog", 
		baseURL, settings.APIVersion, url.PathEscape(issue))

	// Build payload
	payload := BuildWorklogPayload(startedISO, seconds, settings.APIVersion)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		result.Body = fmt.Sprintf("Failed to marshal payload: %v", err)
		return result
	}

	// Create request
	req, err := http.NewRequest("POST", issueURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		result.Body = fmt.Sprintf("Failed to create request: %v", err)
		return result
	}

	// Add headers
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Log debug info
	logger.Debug("Posting worklog to %s with %d seconds using API v%s", issue, seconds, settings.APIVersion)
	payloadStr, _ := json.MarshalIndent(payload, "", "  ")
	logger.Debug("Payload: %s", payloadStr)
	logger.Debug("POST request to: %s", issueURL)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Body = fmt.Sprintf("Failed to send request: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Body = fmt.Sprintf("Failed to read response: %v", err)
		return result
	}
	bodyStr := string(bodyBytes)
	result.Code = resp.StatusCode
	result.Body = bodyStr

	// Check if successful
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		result.Success = true
	} else {
		logger.Error("HTTP %d response: %s", resp.StatusCode, bodyStr)
	}

	return result
}

// GetAssignedIssues fetches issues assigned to the current user
func GetAssignedIssues(settings *Settings, logger *Logger) ([]map[string]interface{}, error) {
	// Create authentication header
	auth := fmt.Sprintf("%s:%s", settings.JiraEmailOrUser, settings.JiraAPIToken)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// Build URL and query
	baseURL := strings.TrimSuffix(settings.JiraBaseURL, "/")
	var apiURL string
	
	if settings.APIVersion == "3" {
		// Jira Cloud API v3
		apiURL = fmt.Sprintf("%s/rest/api/3/search/jql", baseURL)
	} else {
		// Jira Server/DC API v2
		apiURL = fmt.Sprintf("%s/rest/api/2/search", baseURL)
	}
	
	// Create JQL query
	jql := "assignee = currentUser() AND status NOT IN (Done, Closed, Completed, Wasted) AND key != 'CLOUD-1154' AND parent != 'CLOUD-1154'"
	
	// Build query parameters depending on the API version
	var fullURL string
	if settings.APIVersion == "3" {
		// For the new JQL endpoint, we need to POST the query with the fields parameter in the body
		fullURL = apiURL
	} else {
		// Keep the old format for API v2
		queryParams := url.Values{
			"jql":        {jql},
			"fields":     {"key,summary,parent,issuetype,customfield_10014"},
			"maxResults": {"100"},
		}
		fullURL = fmt.Sprintf("%s?%s", apiURL, queryParams.Encode())
	}
	
	logger.Debug("Fetching issues from API endpoint: %s", fullURL)
	
	var req *http.Request
	var err error
	
	if settings.APIVersion == "3" {
		// Create POST request with JSON body for API v3
		requestBody := map[string]interface{}{
			"jql": jql,
			"fields": []string{"key", "summary", "parent", "issuetype", "customfield_10014"},
			"maxResults": 100,
		}
		requestJSON, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON request: %v", err)
		}
		
		req, err = http.NewRequest("POST", fullURL, bytes.NewBuffer(requestJSON))
	} else {
		// Create GET request for API v2
		req, err = http.NewRequest("GET", fullURL, nil)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add headers
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Accept", "application/json")
	
	// Add Content-Type header for POST requests
	if settings.APIVersion == "3" {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
	}
	
	// Parse response
	var result struct {
		Issues []map[string]interface{} `json:"issues"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	logger.Debug("Found %d issues assigned to current user", len(result.Issues))
	return result.Issues, nil
}

// GetEpicsFromIssues extracts epics from issues
func GetEpicsFromIssues(issues []map[string]interface{}) map[string]Epic {
	epics := make(map[string]Epic)
	
	for _, issue := range issues {
		issueKey, _ := issue["key"].(string)
		fields, ok := issue["fields"].(map[string]interface{})
		if !ok || issueKey == "" {
			continue
		}
		
		// Check if issue is an epic
		issueType, _ := fields["issuetype"].(map[string]interface{})
		if issueType != nil {
			typeName, _ := issueType["name"].(string)
			if strings.ToLower(typeName) == "epic" && issueKey != "CLOUD-1154" {
				summary, _ := fields["summary"].(string)
				epics[issueKey] = Epic{
					Key:     issueKey,
					Summary: summary,
					Type:    "epic",
				}
			}
		}
		
		// Check parent
		parent, _ := fields["parent"].(map[string]interface{})
		if parent != nil {
			parentKey, _ := parent["key"].(string)
			if parentKey != "" && parentKey != "CLOUD-1154" {
				if _, exists := epics[parentKey]; !exists {
					parentFields, _ := parent["fields"].(map[string]interface{})
					summary := "No summary"
					if parentFields != nil {
						summary, _ = parentFields["summary"].(string)
					}
					epics[parentKey] = Epic{
						Key:     parentKey,
						Summary: summary,
						Type:    "parent",
					}
				}
			}
		}
		
		// Check for epic link (customfield_10014)
		if epicField, ok := fields["customfield_10014"].(string); ok && epicField != "" && epicField != "CLOUD-1154" {
			if _, exists := epics[epicField]; !exists {
				epics[epicField] = Epic{
					Key:     epicField,
					Summary: fmt.Sprintf("Epic: %s", epicField),
					Type:    "epic_link",
				}
			}
		}
	}
	
	return epics
}
