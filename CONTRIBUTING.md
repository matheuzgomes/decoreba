# Contributing

Thanks for considering contributing to decoreba. This document covers the practical steps: reporting bugs, suggesting changes, and opening pull requests.

## Code of Conduct

This project follows the [Contributor Covenant](https://www.contributor-covenant.org/). By participating, you agree to act respectfully toward everyone involved.

## Before you open an issue

Search the existing issues to see if your question or problem has been raised before. If you find something related, add your context there instead of opening a duplicate.

## Reporting bugs

Open a GitHub issue and include:

- What you ran and what happened.
- What you expected to happen instead.
- Your operating system, shell, and decoreba version (`decoreba version`).
- The steps to reproduce, starting from a clean state if possible.

If the bug involves a crash or corrupted data, include the output and any error message.

## Reporting security issues

Do not open a public issue. Send a message to the maintainer directly through GitHub or open a draft security advisory.

## Suggesting features

Open an issue describing the problem you want to solve, not just the solution you have in mind. Explain how it fits into decoreba as a command vault. Feature requests without a clear use case are likely to be closed.

## Setting up the project

You need Go 1.25 or newer. `make` is required for the build command below.

```bash
git clone https://github.com/matheuzgomes/decoreba
cd decoreba
make build
go test ./...
```

The binary is written to `./decoreba`. You can run it directly:

```bash
./decoreba version
```

## Code style

- Run `gofmt -s` before every commit.
- Run `go vet ./...` and fix any warnings.
- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Package names are lowercase, no underscores.
- Exported functions and types have doc comments.
- Tests live next to the code they test, in the same package (`package cli`, not `package cli_test`).
- Zero external dependencies for the core binary. If you need a new dependency, open an issue first to discuss it.

## Commit messages

Write commit messages in the imperative: "Add fuzzy search option", not "Added fuzzy search". Reference the issue number when applicable: `Fixes #123`.

Keep commits atomic. One logical change per commit. If you need to fix a mistake, amend the commit or rebase instead of adding a "fixup" commit.

## Opening a pull request

1. Fork the repository and clone your fork.
2. Create a branch: `git checkout -b feature/description`.
3. Make your changes and run `go test ./...` and `go vet ./...`.
4. Commit with a descriptive message.
5. Push to your fork and open a pull request against `master`.
6. Respond to review comments. Pull requests that go silent for a month may be closed.

Before opening the pull request, make sure:

- `go test ./...` passes.
- `go vet ./...` is clean.
- New code has tests.
- Documentation is updated if the user-facing behavior changed.

## Getting help

Open an issue for bugs and feature requests. If you are unsure about a change, open a draft pull request early so we can discuss it before you invest too much time.

## Recognition

Contributors may be credited in the release notes. If you make a significant contribution, let me know if you want to be added to the project README.
