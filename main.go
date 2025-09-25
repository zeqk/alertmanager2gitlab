package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Handler para recibir alertas
func alertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	for _, alert := range payload.Alerts {
		title := fmt.Sprintf("%s - %s", alert.Labels["alertname"], alert.Labels["instance"])
		desc := alert.Annotations["summary"]
		if err := createGitLabIssue(title, desc); err != nil {
			log.Println("Error creando issue:", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Alerta recibida y procesada"))
}

func main() {
	http.HandleFunc("/alert", alertHandler)
	log.Println("Servidor escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
