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
		log.Warn("Método no permitido")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load templates from files
	titleTemplateBytes, err := os.ReadFile("templates/title.tmpl")
	if err != nil {
		log.Error("Error leyendo title template: ", err)
		http.Error(w, "Error reading title template", http.StatusInternalServerError)
		return
	}
	descTemplateBytes, err := os.ReadFile("templates/description.tmpl")
	if err != nil {
		log.Error("Error leyendo description template: ", err)
		http.Error(w, "Error reading description template", http.StatusInternalServerError)
		return
	}

	// Funciones personalizadas
	funcMap := template.FuncMap{
		"replace": strings.ReplaceAll,
		"upper":   strings.ToUpper,
	}
	titleTmpl, err := template.New("title").Funcs(funcMap).Parse(string(titleTemplateBytes))
	if err != nil {
		log.Error("Error parseando title template: ", err)
		http.Error(w, "Error parsing title template", http.StatusInternalServerError)
		return
	}
	descTmpl, err := template.New("description").Funcs(funcMap).Parse(string(descTemplateBytes))
	if err != nil {
		log.Error("Error parseando description template: ", err)
		http.Error(w, "Error parsing description template", http.StatusInternalServerError)
		return
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Warn("JSON inválido: ", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var titleBuf, descBuf bytes.Buffer

	if err := titleTmpl.Execute(&titleBuf, payload); err != nil {
		log.Error("Error ejecutando title template: ", err)
	}
	if err := descTmpl.Execute(&descBuf, payload); err != nil {
		log.Error("Error ejecutando description template: ", err)
	}

	title := strings.TrimSpace(titleBuf.String())
	desc := descBuf.String()

	if err := createGitLabIssue(title, desc); err != nil {
		log.Error("Error creando issue: ", err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Alert received and processed")); err != nil {
		log.Error("Error escribiendo respuesta: ", err)
	}
}

func main() {
	// Configurar el nivel de log desde la variable de entorno LOG_LEVEL
	level, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.SetLevel(log.InfoLevel) // Nivel por defecto
		log.Warn("LOG_LEVEL inválido o no definido, usando InfoLevel")
	} else {
		log.SetLevel(level)
		log.Infof("LOG_LEVEL configurado en: %s", level)
	}

	// Log de los templates al iniciar
	titleTemplateBytes, err := os.ReadFile("templates/title.tmpl")
	if err != nil {
		log.Error("Error leyendo title.tmpl: ", err)
	} else {
		log.Info("Contenido de title.tmpl:\n" + string(titleTemplateBytes))
	}
	descTemplateBytes, err := os.ReadFile("templates/description.tmpl")
	if err != nil {
		log.Error("Error leyendo description.tmpl: ", err)
	} else {
		log.Info("Contenido de description.tmpl:\n" + string(descTemplateBytes))
	}

	http.HandleFunc("/alert", alertHandler)
	log.Info("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
