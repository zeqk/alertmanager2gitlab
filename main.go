package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Handler to receive alerts
func alertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for _, alert := range payload.Alerts {
		title := fmt.Sprintf("%s - %s", alert.Labels["alertname"], alert.Labels["instance"])
		desc := alert.Annotations["summary"]
		if err := createGitLabIssue(title, desc); err != nil {
			log.Println("Error creating issue:", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alert received and processed"))
}

func main() {
	http.HandleFunc("/alert", alertHandler)
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
