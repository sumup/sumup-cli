<div align="center">

# SumUp CLI

Command line tool for interacting with SumUp APIs.

[![Documentation][docs-badge]](https://developer.sumup.com)
[![License](https://img.shields.io/github/license/sumup/sumup-rs)](./LICENSE)

</div>

SumUp CLI tool allows you to manage your SumUp account, create checkouts, and much more all from your terminal.

## Getting started

### Install with Homebrew

```bash
brew install sumup/cli/sumup
```

### Install with Go

```bash
go install github.com/sumup/sumup-cli/cmd/sumup
```

The CLI expects an API key via the `SUMUP_API_KEY` environment variable by default. You can also pass `--api-key` explicitly.

```bash
export SUMUP_API_KEY=your_api_key
```

## Shell completion

Generate a completion script for your shell and load it:

```bash
# bash
source <(sumup completion bash)

# zsh
source <(sumup completion zsh)

# fish
sumup completion fish > ~/.config/fish/completions/sumup.fish
```

Release archives also include pre-generated bash, zsh, and fish completion scripts plus a man page (`man/man1/sumup.1.gz`).

## Managing merchant context

To avoid repeating the `--merchant-code` flag in every command, you can set a merchant context:

```bash
# Set the merchant context interactively
sumup context set

# View the current merchant context
sumup context get

# Unset the merchant context
sumup context unset
```

Once set, all commands that accept `--merchant-code` will use the context value by default. You can still override it by providing the flag explicitly.

## Create a checkout

```bash
sumup checkouts create \
  --reference order-123 \
  --amount 19.99 \
  --currency EUR \
  --merchant-code M123 \
  --description "Ticket purchase" \
  --return-url https://example.com/return \
  --redirect-url https://example.com/3ds \
  --customer-id cst_42 \
  --purpose "Event"
```

## Manage readers

List readers for a merchant:

```bash
sumup readers list --merchant-code M123
```

Pair a new reader with a pairing code:

```bash
sumup readers add \
  --merchant-code M123 \
  --pairing-code ABCDEF \
  --name "Front counter"
```

Trigger a checkout on a reader (this example charges EUR 14.99 and offers tip
rates):

```bash
sumup readers checkout \
  --merchant-code M123 \
  --reader-id reader_42 \
  --amount 14.99 \
  --currency EUR \
  --tip-rate 0.10 \
  --tip-rate 0.15 \
  --description "In-person order #123"
```

When using affiliate attribution, pass all affiliate flags: `--affiliate-app-id`, `--affiliate-key`, and `--affiliate-foreign-transaction-id`.

Trigger a checkout on a SumUp Go reader. `--client-transaction-id` is required and used as the idempotency key:

```bash
sumup readers go-checkout \
  --merchant-code M123 \
  --amount 14.99 \
  --currency EUR \
  --client-transaction-id 19e12390-72cf-4f9f-80b5-b0c8a67fa43f \
  reader_42
```

Check the last known status of a paired reader:

```bash
sumup readers status \
  --merchant-code M123 \
  reader_42
```

## OpenAPI command coverage

The CLI keeps its user-facing command implementations handwritten, but derives
an operation catalog from the OpenAPI document shipped with the exact
`sumup-go` version pinned in `go.mod`. The catalog records operation IDs, SDK
clients and methods, HTTP paths, parameters, and request-body metadata.

Run the generator after updating `sumup-go`:

```bash
make generate
```

Each API leaf binds its generated OpenAPI operation ID directly in the command
definition. Tests enforce a one-to-one relationship between the pinned SDK,
the generated catalog, and the CLI command tree, so an SDK upgrade fails CI
until every new endpoint has exactly one corresponding command.

### Developer portal code samples

Generate a deterministic, portal-compatible JSON catalog containing one CLI
code sample for every OpenAPI operation exposed by the CLI:

```bash
VERSION=v0.1.0 make generate-codesamples
```

The catalog is written to `code-samples.json` by default. Set
`CODESAMPLES_OUT` to use a different path. The generated file uses the same
versioned schema as the SDK sample catalogs and is not committed to this
repository. Published releases automatically open or update a pull request
that writes the catalog to `src/codesamples/cli.json` in
`sumup/sumup-developer`.

[docs-badge]: https://img.shields.io/badge/SumUp-documentation-white.svg?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgY29sb3I9IndoaXRlIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogICAgPHBhdGggZD0iTTIyLjI5IDBIMS43Qy43NyAwIDAgLjc3IDAgMS43MVYyMi4zYzAgLjkzLjc3IDEuNyAxLjcxIDEuN0gyMi4zYy45NCAwIDEuNzEtLjc3IDEuNzEtMS43MVYxLjdDMjQgLjc3IDIzLjIzIDAgMjIuMjkgMFptLTcuMjIgMTguMDdhNS42MiA1LjYyIDAgMCAxLTcuNjguMjQuMzYuMzYgMCAwIDEtLjAxLS40OWw3LjQ0LTcuNDRhLjM1LjM1IDAgMCAxIC40OSAwIDUuNiA1LjYgMCAwIDEtLjI0IDcuNjlabTEuNTUtMTEuOS03LjQ0IDcuNDVhLjM1LjM1IDAgMCAxLS41IDAgNS42MSA1LjYxIDAgMCAxIDcuOS03Ljk2bC4wMy4wM2MuMTMuMTMuMTQuMzUuMDEuNDlaIiBmaWxsPSJjdXJyZW50Q29sb3IiLz4KPC9zdmc+
