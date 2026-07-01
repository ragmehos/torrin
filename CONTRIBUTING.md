# Contributing to Torrin

Thanks for your interest in contributing. Here's how to get started.

## Getting started

1. Fork the repo and clone your fork.
2. Create a branch: `git checkout -b my-feature`.
3. Make your changes.
4. Build, vet, and test:
   ```bash
   go build ./...
   go vet ./...
   go test ./...
   ```
5. Push and open a pull request.

Go 1.25+ is required. Most services can be run individually (`go run ./api`, `go run ./ingest`, …); the full stack runs via `deploy/docker compose up`.

## Pull requests

- Keep each PR focused on one thing.
- **PR titles must follow [Conventional Commits](https://www.conventionalcommits.org)** (`feat:`, `fix:`, `chore:`, `docs:`, …) — this is enforced by CI.
- Include or update **tests** — new behavior without tests fails the `require-tests` check.
- Write a clear description of *what* and *why*, and link any related issue.
- Make sure `go build ./...`, `go vet ./...`, and `go test ./...` pass.

## Reporting bugs

Open a [bug report](../../issues/new?template=bug_report.yml) with what you expected, what happened, steps to reproduce, and version/environment info.

## Feature requests

Open a [feature request](../../issues/new?template=feature_request.yml) describing the **problem / use case**, not just the solution — we'll discuss the best approach together.

## Security

Do not open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).

## License

By contributing, you agree that your contributions are licensed under the [AGPL-3.0](LICENSE).
