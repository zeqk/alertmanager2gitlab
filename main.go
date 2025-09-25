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
		log.Printf("Error reading title template: %v", err)
		http.Error(w, "Error reading title template", http.StatusInternalServerError)
		return
	}
	descTemplateBytes, err := os.ReadFile("templates/description.tmpl")
	if err != nil {
		log.Printf("Error reading description template: %v", err)
		http.Error(w, "Error reading description template", http.StatusInternalServerError)
		return
	}

	titleTmpl, err := template.New("title").Parse(string(titleTemplateBytes))
	if err != nil {
		log.Printf("Error parsing title template: %v", err)
		http.Error(w, "Error parsing title template", http.StatusInternalServerError)
		return
	}
	descTmpl, err := template.New("description").Parse(string(descTemplateBytes))
	if err != nil {
		log.Printf("Error parsing description template: %v", err)
		http.Error(w, "Error parsing description template", http.StatusInternalServerError)
		return
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var titleBuf, descBuf bytes.Buffer

	if err := titleTmpl.Execute(&titleBuf, payload); err != nil {
		log.Printf("Error executing title template: %v", err)
	}
	if err := descTmpl.Execute(&descBuf, payload); err != nil {
		log.Printf("Error executing description template: %v", err)
	}

	title := strings.TrimSpace(titleBuf.String())
	desc := descBuf.String()

	if err := createGitLabIssue(title, desc); err != nil {
		log.Printf("Error creating issue: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Alert received and processed")); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func main() {
	http.HandleFunc("/alert", alertHandler)
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
