package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

// Handler to receive alerts
func alertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load templates from files
	titleTemplateBytes, err := os.ReadFile("templates/title.tmpl")
	if err != nil {
		http.Error(w, "Error reading title template", http.StatusInternalServerError)
		return
	}
	descTemplateBytes, err := os.ReadFile("templates/description.tmpl")
	if err != nil {
		http.Error(w, "Error reading description template", http.StatusInternalServerError)
		return
	}

	titleTmpl, err := template.New("title").Parse(string(titleTemplateBytes))
	if err != nil {
		http.Error(w, "Error parsing title template", http.StatusInternalServerError)
		return
	}
	descTmpl, err := template.New("description").Parse(string(descTemplateBytes))
	if err != nil {
		http.Error(w, "Error parsing description template", http.StatusInternalServerError)
		return
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var titleBuf, descBuf bytes.Buffer

	if err := titleTmpl.Execute(&titleBuf, payload); err != nil {
		log.Println("Error executing title template:", err)
	}
	if err := descTmpl.Execute(&descBuf, payload); err != nil {
		log.Println("Error executing description template:", err)
	}

	title := strings.TrimSpace(titleBuf.String())
	desc := descBuf.String()

	if err := createGitLabIssue(title, desc); err != nil {
		log.Println("Error creating issue:", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alert received and processed"))
}

func main() {
	http.HandleFunc("/alert", alertHandler)
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
