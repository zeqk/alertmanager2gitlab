package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Checks if there is an open issue with the same title
func issueExists(title string) (bool, error) {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	projectID := os.Getenv("GITLAB_PROJECT_ID")
	apiUrl := os.Getenv("GITLAB_API_URL")
	if gitlabToken == "" || projectID == "" {
		return false, fmt.Errorf("missing GITLAB_TOKEN or GITLAB_PROJECT_ID in environment variables")
	}

	projectsApiURL := fmt.Sprintf("%s/projects/%s/issues?state=opened&search=%s&in=title", apiUrl, projectID, url.QueryEscape(title))
	// curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" "$GITLAB_API_URL/projects/$GITLAB_PROJECT_ID/issues?state=opened&search=$title&in=title"
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
		log.Debugf("Checking issue with title: %s", issue.Title)
		if strings.TrimSpace(issue.Title) == strings.TrimSpace(title) {
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
		log.Error("Error verificando si existe el issue: ", err)
		return err
	}
	if exists {
		log.Warnf("Issue ya existe en GitLab: %s", title)
		return nil
	}

	url := fmt.Sprintf("%s/projects/%s/issues", apiUrl, projectID)
	payload := map[string]string{
		"title":       title,
		"description": description,
	}
	jsonPayload, _ := json.Marshal(payload)

	log.Debugf("Creando issue en GitLab con payload: %s", string(jsonPayload))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error creando el issue en GitLab: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Error("Error creando el issue, status: ", resp.Status)
		return fmt.Errorf("error creating issue, status: %s", resp.Status)
	}

	var createdIssue GitLabIssue
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &createdIssue); err != nil {
		log.Infof("Issue creado: %s", title)
	} else {
		log.Infof("Issue creado: %s (IID: %d)", title, createdIssue.IID)
	}
	return nil
}
