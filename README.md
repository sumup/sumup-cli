<div align="center">

# SumUp CLI

Command line tool for interacting with SumUp APIs.

[![Documentation][docs-badge]](https://developer.sumup.com)
[![License](https://img.shields.io/github/license/sumup/sumup-rs)](./LICENSE)

</div>

SumUp CLI tool allows you to manage your SumUp account, create checkouts, and much more all from your terminal.

## Getting started

```bash
go install github.com/sumup/sumup-cli/cmd/sumup
```

The CLI expects an API key via the `SUMUP_API_KEY` environment variable by default. You can also pass `--api-key` explicitly.

```bash
export SUMUP_API_KEY=your_api_key
```

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

[docs-badge]: https://img.shields.io/badge/SumUp-documentation-white.svg?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgY29sb3I9IndoaXRlIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogICAgPHBhdGggZD0iTTIyLjI5IDBIMS43Qy43NyAwIDAgLjc3IDAgMS43MVYyMi4zYzAgLjkzLjc3IDEuNyAxLjcxIDEuN0gyMi4zYy45NCAwIDEuNzEtLjc3IDEuNzEtMS43MVYxLjdDMjQgLjc3IDIzLjIzIDAgMjIuMjkgMFptLTcuMjIgMTguMDdhNS42MiA1LjYyIDAgMCAxLTcuNjguMjQuMzYuMzYgMCAwIDEtLjAxLS40OWw3LjQ0LTcuNDRhLjM1LjM1IDAgMCAxIC40OSAwIDUuNiA1LjYgMCAwIDEtLjI0IDcuNjlabTEuNTUtMTEuOS03LjQ0IDcuNDVhLjM1LjM1IDAgMCAxLS41IDAgNS42MSA1LjYxIDAgMCAxIDcuOS03Ljk2bC4wMy4wM2MuMTMuMTMuMTQuMzUuMDEuNDlaIiBmaWxsPSJjdXJyZW50Q29sb3IiLz4KPC9zdmc+
