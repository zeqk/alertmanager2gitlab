# Agent Instructions

## Project Shape

- This is a single Go module and `main` package. Keep changes close to the existing layout unless a new package clearly improves ownership.
- `main.go` owns HTTP handling and startup, `gitlab.go` owns GitLab API operations and issue deduplication, and `models.go` defines webhook/API payload types.
- Alert templates live in `templates/` and are loaded from relative paths at runtime. The title uses `html/template`; descriptions and comments use `text/template`. Preserve the existing template functions when changing title or description rendering.
- Read [README.md](README.md) for setup, environment variables, token permissions, Docker usage, and payload examples; link to it rather than duplicating those details here.

## Behavior to Preserve and Check

- `POST /alert` returns `202 Accepted` before payload decoding and GitLab work finish. Processing errors therefore cannot change the response; account for the asynchronous boundary in tests and behavior changes.
- Only top-level payloads whose `status` is `firing` are processed. Check this filter when changing alert-state behavior or deployment configuration.
- Configuration is read directly from environment variables. Do not assume README defaults are applied by local execution; verify the actual code and Docker configuration.
- Templates depend on the process working directory. Keep runtime paths and Docker/Kubernetes mounts aligned when changing deployment files.
- When changing Kubernetes routing, check the webhook hostname against the Service name. When changing title templates, keep the README sample payload and the template's annotation key consistent.

## Go Workflow

- Format Go changes with `gofmt`.
- Run `go test ./...` after code changes. There are currently no test files, so add focused tests for new behavior rather than treating a clean command as behavior coverage.
- Also run `go vet ./...` for meaningful Go changes; `go build ./...` is a quick compile check.
- Keep tests in package `main` unless production code is deliberately moved into importable packages.