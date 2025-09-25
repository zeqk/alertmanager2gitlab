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