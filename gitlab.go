package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Checks if there is an open issue with the same title
func issueExists(title string) (bool, error) {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	projectID := os.Getenv("GITLAB_PROJECT_ID")
	apiUrl := os.Getenv("GITLAB_API_URL")
	if gitlabToken == "" || projectID == "" {
		return false, fmt.Errorf("missing GITLAB_TOKEN or GITLAB_PROJECT_ID in environment variables")
	}

	projectsApiURL := fmt.Sprintf("%s/projects/%s/issues?state=opened&search=%s", apiUrl, projectID, url.QueryEscape(title))
	req, _ := http.NewRequest("GET", projectsApiURL, nil)
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return false, fmt.Errorf("error fetching issues, status: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var issues []GitLabIssue
	if err := json.Unmarshal(body, &issues); err != nil {
		return false, err
	}

	for _, issue := range issues {
		if issue.Title == title {
			return true, nil
		}
	}

	return false, nil
}

// Creates a GitLab issue
func createGitLabIssue(title, description string) error {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	projectID := os.Getenv("GITLAB_PROJECT_ID")
	apiUrl := os.Getenv("GITLAB_API_URL")

	exists, err := issueExists(title)
	if err != nil {
		return err
	}
	if exists {
		log.Println("Issue already exists in GitLab:", title)
		return nil
	}

	url := fmt.Sprintf("%s/projects/%s/issues", apiUrl, projectID)
	payload := map[string]string{
		"title":       title,
		"description": description,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("error creating issue, status: %s", resp.Status)
	}

	log.Println("Issue created:", title)
	return nil
}
