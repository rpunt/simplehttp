# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

simplehttp is a Go HTTP client library that wraps `net/http` to eliminate boilerplate. It provides a single-file implementation (`simplehttp.go`) where request parameters (headers, data, query params) are set as client attributes rather than passed per-call.

## Git Workflow

Never commit directly to main. Always create a feature branch for changes.

## Common Commands

- **Run tests:** `go test -v ./...`
- **Run a single test/subtest:** `go test -v -run TestWrapperMethods/GET ./...`
- **Check formatting:** `gofmt -l .` (CI enforces this via `make fmt`)
- **Lint:** `golangci-lint run -v` (config in `.golangci.yml`, ~40 linters enabled)
- **All checks:** `make all` (runs fmt + test)

## Architecture

The entire library is a single package with two exported types:

- **`HTTPClient`** — holds `BaseURL`, `Headers`, `Data`, `Params` (all `map[string]string`), and a `Client *http.Client` with a 10-second timeout. Created via `New(baseURL)`.
- **`HTTPResponse`** — contains `Body` (string), `Code` (int), `Headers` (map[string][]string).

All HTTP method functions (`Get`, `Post`, `Patch`, `Put`, `Delete`, `Head`) delegate to the unexported `sendRequest`, which marshals `Data` as JSON, applies headers and query params, executes the request, and reads the full response body into a string.

**Design note:** `context.Context` is intentionally omitted from the API to keep it simple. This library prioritizes ease of use over composability in larger context-propagating systems.

## Testing

Tests use `httptest.NewTLSServer` with a custom handler (`handleHTTP`) that routes by method and path. Test fixtures live in `fixtures/`. The only external dependency is `github.com/google/go-cmp` (used for test comparisons).

## CI

GitHub Actions runs `go test ./... -coverprofile=coverage.txt` across Go 1.21, 1.22, and 1.23 on ubuntu-latest. Coverage is uploaded to Codecov on the latest Go version. Triggered on push/PR to main (ignoring `.md` files).
