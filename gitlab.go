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
	"text/template"

	log "github.com/sirupsen/logrus"
)

// Get issues by title
func getIssuesByTitle(title, projectRef string) ([]GitLabIssue, error) {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	apiUrl := os.Getenv("GITLAB_API_URL")

	projectsApiURL := fmt.Sprintf("%s/projects/%s/issues?state=opened&search=%s&in=title", apiUrl, projectRef, url.QueryEscape(title))
	// curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" "$GITLAB_API_URL/projects/$GITLAB_DEFAULT_PROJECT_ID/issues?state=opened&search=$title&in=title"
	req, _ := http.NewRequest("GET", projectsApiURL, nil)
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error fetching issues, status: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var issues []GitLabIssue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, err
	}

	return issues, nil
}

// Creates a GitLab issue
func createGitLabIssue(title string, payload AlertmanagerPayload, projectRef string) error {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	apiUrl := os.Getenv("GITLAB_API_URL")

	issues, err := getIssuesByTitle(title, projectRef)
	if err != nil {
		log.Error("Error checking if issue exists: ", err)
		return err
	}
	client := &http.Client{}

	if len(issues) > 0 {
		// Find the existing issue IID
		issueIID := issues[0].IID
		log.Warnf("Issue already exists in GitLab: %s (project: %s, IID: %d)", title, projectRef, issueIID)

		// Solo agregar comentarios si GITLAB_COMMENT_ENABLED es "true"
		if os.Getenv("GITLAB_COMMENT_ENABLED") == "true" {
			// Render comment from template and add to existing issue
			commentTemplateBytes, err := os.ReadFile("templates/comment.tmpl")
			if err != nil {
				log.Error("Error reading comment template: ", err)
				return nil
			}
			commentTmpl, err := template.New("comment").Parse(string(commentTemplateBytes))
			if err != nil {
				log.Error("Error parsing comment template: ", err)
				return nil
			}
			var commentBuf bytes.Buffer
			if err := commentTmpl.Execute(&commentBuf, payload); err != nil {
				log.Error("Error executing comment template: ", err)
				return nil
			}
			comment := commentBuf.String()

			// Post the comment
			commentApiURL := fmt.Sprintf("%s/projects/%s/issues/%d/notes", apiUrl, projectRef, issueIID)
			commentPayload := map[string]string{"body": comment}
			commentJson, _ := json.Marshal(commentPayload)
			commentReq, _ := http.NewRequest("POST", commentApiURL, bytes.NewBuffer(commentJson))
			commentReq.Header.Set("Content-Type", "application/json")
			commentReq.Header.Set("PRIVATE-TOKEN", gitlabToken)
			commentResp, err := client.Do(commentReq)
			if err != nil {
				log.Error("Error posting comment to existing issue:", err)
				return nil
			}
			defer commentResp.Body.Close()
			if commentResp.StatusCode >= 300 {
				log.Error("Error posting comment, status: ", commentResp.Status)
				return nil
			}
			log.Infof("Comment added to existing issue: %s (IID: %d)", title, issueIID)
		} else {
			log.Debugf("GITLAB_COMMENT_ENABLED is not 'true', skipping comment on existing issue: %s (IID: %d)", title, issueIID)
		}
		return nil
	}

	// Read and render description template
	descTemplateBytes, err := os.ReadFile("templates/description.tmpl")
	if err != nil {
		log.Error("Error reading description template: ", err)
		return err
	}
	funcMap := template.FuncMap{
		"replace": strings.ReplaceAll,
		"upper":   strings.ToUpper,
	}
	descTmpl, err := template.New("description").Funcs(funcMap).Parse(string(descTemplateBytes))
	if err != nil {
		log.Error("Error parsing description template: ", err)
		return err
	}
	var descBuf bytes.Buffer
	if err := descTmpl.Execute(&descBuf, payload); err != nil {
		log.Error("Error executing description template: ", err)
		return err
	}
	desc := descBuf.String()

	url := fmt.Sprintf("%s/projects/%s/issues", apiUrl, projectRef)
	payloadMap := map[string]string{
		"title":       title,
		"description": desc,
	}
	jsonPayload, _ := json.Marshal(payloadMap)

	log.Debugf("Creating issue in GitLab with payload: %s", string(jsonPayload))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", gitlabToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error creating issue in GitLab: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Error("Error creating issue, status: ", resp.Status)
		return fmt.Errorf("error creating issue, status: %s", resp.Status)
	}

	var createdIssue GitLabIssue
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &createdIssue); err != nil {
		log.Infof("Issue created: %s", title)
	} else {
		log.Infof("Issue created: %s (IID: %d)", title, createdIssue.IID)
	}
	return nil
}
