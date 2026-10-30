<div align="center">

# SumUp CLI

Manage your SumUp account from your terminal.

[![Documentation][docs-badge]](https://developer.sumup.com)
[![License](https://img.shields.io/github/license/sumup/sumup-cli)](./LICENSE)

</div>

Use the SumUp CLI to create checkouts, manage card readers, and view your account data.

## Install

Choose one of the following installation methods:

```bash
# Homebrew
brew install sumup/cli/sumup

# mise
mise use -g github:sumup/sumup-cli@latest

# Go
go install github.com/sumup/sumup-cli/cmd/sumup@latest
```

## Set up Your API Key

Set the `SUMUP_API_KEY` environment variable to your SumUp API key:

```bash
export SUMUP_API_KEY="your_api_key"
```

You can also pass a key with `sumup --api-key "your_api_key" <command>`.

## Find Commands and Options

```bash
sumup --help
sumup readers --help
sumup checkouts create --help
```

Add `--help` to any command to see its options.

## Choose a Merchant

Save a default merchant so you do not need to repeat `--merchant-code`:

```bash
# Choose a merchant
sumup context set

# Show the saved merchant
sumup context get

# Clear the saved merchant
sumup context unset
```

Commands that accept `--merchant-code` use the saved merchant by default.
Pass `--merchant-code` to use a different merchant for one command.

## Create a Checkout

Replace `M123` with your merchant code. Use a unique reference for each checkout.

```bash
sumup checkouts create \
  --reference order-123 \
  --amount 19.99 \
  --currency EUR \
  --merchant-code M123 \
  --description "Ticket purchase"
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

Start a EUR 14.99 payment on a reader with three tip options.
Replace `reader_42` with your reader ID:

```bash
sumup readers checkout \
  --merchant-code M123 \
  --amount 14.99 \
  --currency EUR \
  --tip-rate 0.10 \
  --tip-rate 0.15 \
  --tip-rate 0.20 \
  --description "In-person order #123" \
  reader_42
```

For Pix payments on Brazilian merchants, add `--payment-type pix`.

To link a payment to your affiliate app, pass both `--affiliate-app-id` and `--affiliate-key`.
Use `--affiliate-foreign-transaction-id` to add your own transaction reference.

Start a payment on a SumUp Go reader. Set `--client-transaction-id` to a unique ID
for each new payment. Reuse the same ID when retrying that payment to avoid duplicates:

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

View or change the menu lock settings on a Solo reader. The reader must be online
to change its settings. If it is offline, you can view its last saved settings when available.

```bash
sumup readers settings get --merchant-code M123 reader_42
sumup readers settings update --merchant-code M123 --menu-lock-enabled=false reader_42
```

Set `SUMUP_READER_MENU_LOCK_PASSCODE` to a four-digit PIN to send it with the update.
If the reader has no PIN, include one when you enable menu lock.

## Create a Merchant Token

Create a short-lived merchant token for a SumUp SDK. This command requires
`--json` and returns the access token and its remaining lifetime:

```bash
sumup --json tokens create --merchant-code M123 --action checkout
```

## Use JSON Output

Use `--json` to get results as JSON for scripts:

```bash
sumup --json readers list --merchant-code M123
```

## Shell Completion

Enable command completion for your shell:

```bash
# bash
source <(sumup completion bash)

# zsh
source <(sumup completion zsh)

# fish
sumup completion fish > ~/.config/fish/completions/sumup.fish
```

Release archives also include completion scripts for bash, zsh, and fish,
plus a man page at `man/man1/sumup.1.gz`.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and contribution guidelines.

[docs-badge]: https://img.shields.io/badge/SumUp-documentation-white.svg?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgY29sb3I9IndoaXRlIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogICAgPHBhdGggZD0iTTIyLjI5IDBIMS43Qy43NyAwIDAgLjc3IDAgMS43MVYyMi4zYzAgLjkzLjc3IDEuNyAxLjcxIDEuN0gyMi4zYy45NCAwIDEuNzEtLjc3IDEuNzEtMS43MVYxLjdDMjQgLjc3IDIzLjIzIDAgMjIuMjkgMFptLTcuMjIgMTguMDdhNS42MiA1LjYyIDAgMCAxLTcuNjguMjQuMzYuMzYgMCAwIDEtLjAxLS40OWw3LjQ0LTcuNDRhLjM1LjM1IDAgMCAxIC40OSAwIDUuNiA1LjYgMCAwIDEtLjI0IDcuNjlabTEuNTUtMTEuOS03LjQ0IDcuNDVhLjM1LjM1IDAgMCAxLS41IDAgNS42MSA1LjYxIDAgMCAxIDcuOS03Ljk2bC4wMy4wM2MuMTMuMTMuMTQuMzUuMDEuNDlaIiBmaWxsPSJjdXJyZW50Q29sb3IiLz4KPC9zdmc+
