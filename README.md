# alertmanager2gitlab

A service written in **Go** that receives alerts from **Prometheus Alertmanager** via webhook and automatically creates **issues in GitLab** using the REST API.

## 🚀 Features

- Receives alerts in JSON format from Alertmanager.
- Creates a GitLab issue for each received alert.
- Prevents duplicates by checking if an open issue with the same title already exists.
- Configurable via **environment variables**.
- Lightweight Docker image.

---

## 📦 Installation

### 1. Clone the repository

```bash
git clone https://github.com/zeqk/alertmanager2gitlab.git
cd alertmanager2gitlab
```

### 2. Build and run

```bash
go build -o alertmanager2gitlab
./alertmanager2gitlab
```
## ⚙️ Configuration

This service uses the following environment variables:

- `GITLAB_TOKEN` → GitLab personal access token with permissions to create issues.
- `GITLAB_PROJECT_ID` → GitLab project ID where issues will be created.
- `GITLAB_API_URL` → (Default `https://gitlab.com/api/v4`)

Example:
```bash
export GITLAB_TOKEN="glpat-xxxxxx"
export GITLAB_PROJECT_ID="123456"
export GITLAB_API_URL="https://gitlab.com/api/v4"
```

## 🐳 Run with Docker

Build the image:

```bash
docker build -t alertmanager2gitlab .
```


Run the container:

```bash
docker run -d \
  -e GITLAB_TOKEN="glpat-xxxxxx" \
  -e GITLAB_PROJECT_ID="123456" \
  -e GITLAB_API_URL="https://gitlab.com/api/v4" \
  -p 8080:8080 \
  alertmanager2gitlab
```

## 🔎 Test with curl

Simulate an Alertmanager alert:

```bash
curl -X POST http://localhost:8080/alert \
  -H "Content-Type: application/json" \
  -d '{
    "version": "4",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPU",
          "instance": "server1"
        },
        "annotations": {
          "summary": "CPU usage above 90%"
        },
        "startsAt": "2025-09-25T05:00:00Z"
      }
    ],
    "commonLabels": {
      "alertname": "HighCPU",
      "instance": "server1"
    },
    "commonAnnotations": {
      "summary": "CPU usage above 90%",
      "description": "CPU usage on server1 exceeded threshold",
      "exception": "None"
    }
  }'
```

## 🔗 Configure in Alertmanager

In your alertmanager.yml:

```yaml
receivers:
  - name: 'gitlab-webhook'
    webhook_configs:
      - url: 'http://alert2gitlab:8080/alert'
```

## 📝 Customizing Issue Templates

This service uses Go templates to generate the title and description for each GitLab issue. By default, the templates are located in the `templates/` directory:

- `templates/title.tmpl`: Defines the issue title. Example:
  ```gotmpl
  {{ .CommonAnnotations.summary }}
  ```
- `templates/description.tmpl`: Defines the issue description. Example:
  ```gotmpl
  {{ .CommonAnnotations.description }}
  
  ```
  {{ .CommonAnnotations.exception }}
  ```
  
  URL: {{ .ExternalURL }}
  
  Common Labels:
  {{ range $key, $value := .CommonLabels }}- {{$key}}: {{$value}}
  {{ end }}
  ```

You can modify these templates to fit your needs. The templates use the [Go text/template](https://pkg.go.dev/text/template) syntax and have access to all fields in the Alertmanager webhook payload.

### Mounting Custom Templates in Docker

To use your own templates, mount them into the container at startup:

```bash
docker run -d \
  -e GITLAB_TOKEN="glpat-xxxxxx" \
  -e GITLAB_PROJECT_ID="123456" \
  -e GITLAB_API_URL="https://gitlab.com/api/v4" \
  -p 8080:8080 \
  -v /path/to/your/templates:/templates:ro \
  alertmanager2gitlab
```

Replace `/path/to/your/templates` with the directory containing your `title.tmpl` and `description.tmpl` files. The application will automatically load these templates at runtime.