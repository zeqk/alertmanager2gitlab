package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Handler to receive alerts
func alertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load templates from files
	titleTemplateBytes, err := os.ReadFile("templates/title.tmpl")
	if err != nil {
		log.Error("Error reading title template: ", err)
		http.Error(w, "Error reading title template", http.StatusInternalServerError)
		return
	}
	descTemplateBytes, err := os.ReadFile("templates/description.tmpl")
	if err != nil {
		log.Error("Error reading description template: ", err)
		http.Error(w, "Error reading description template", http.StatusInternalServerError)
		return
	}

	// Custom functions
	funcMap := template.FuncMap{
		"replace": strings.ReplaceAll,
		"upper":   strings.ToUpper,
	}
	titleTmpl, err := template.New("title").Funcs(funcMap).Parse(string(titleTemplateBytes))
	if err != nil {
		log.Error("Error parsing title template: ", err)
		http.Error(w, "Error parsing title template", http.StatusInternalServerError)
		return
	}
	descTmpl, err := template.New("description").Funcs(funcMap).Parse(string(descTemplateBytes))
	if err != nil {
		log.Error("Error parsing description template: ", err)
		http.Error(w, "Error parsing description template", http.StatusInternalServerError)
		return
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Warn("Invalid JSON: ", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var titleBuf, descBuf bytes.Buffer

	if err := titleTmpl.Execute(&titleBuf, payload); err != nil {
		log.Error("Error executing title template: ", err)
	}
	if err := descTmpl.Execute(&descBuf, payload); err != nil {
		log.Error("Error executing description template: ", err)
	}

	title := strings.TrimSpace(titleBuf.String())
	desc := descBuf.String()

	// Get project_path from CommonLabels if present
	projectPath := ""
	if val, ok := payload.CommonLabels["project_path"]; ok {
		projectPath = val
	}

	if err := createGitLabIssue(title, desc, projectPath); err != nil {
		log.Error("Error creating issue: ", err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Alert received and processed")); err != nil {
		log.Error("Error writing response: ", err)
	}
}

func main() {
	// Set log level from LOG_LEVEL environment variable
	level, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.SetLevel(log.InfoLevel) // Default level
		log.Warn("LOG_LEVEL invalid or not set, using InfoLevel")
	} else {
		log.SetLevel(level)
		log.Infof("LOG_LEVEL set to: %s", level)
	}

	gitlabToken := os.Getenv("GITLAB_TOKEN")
	projectID := os.Getenv("GITLAB_DEFAULT_PROJECT_ID")

	if gitlabToken == "" || projectID == "" {
		log.Error("missing GITLAB_TOKEN or GITLAB_DEFAULT_PROJECT_ID in environment variables")
	}

	// Log templates on startup
	titleTemplateBytes, err := os.ReadFile("templates/title.tmpl")
	if err != nil {
		log.Error("Error reading title.tmpl: ", err)
	} else {
		log.Info("Contents of title.tmpl:\n" + string(titleTemplateBytes))
	}
	descTemplateBytes, err := os.ReadFile("templates/description.tmpl")
	if err != nil {
		log.Error("Error reading description.tmpl: ", err)
	} else {
		log.Info("Contents of description.tmpl:\n" + string(descTemplateBytes))
	}

	http.HandleFunc("/alert", alertHandler)
	log.Info("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
