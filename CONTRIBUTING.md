# Contributing

## Development setup

Install the Go version listed in `go.mod` or a newer version. From the repository root,
download dependencies and build the CLI:

```bash
go mod download
make build
./bin/sumup --help
```

The binary is written to `bin/sumup`. Run `make help` to see the available tasks.

## Checks

Before opening a pull request, run:

```bash
make fmt
make lint
make test
```

Formatting and linting require `golangci-lint`. Use the version set by
`GOLANGCI_LINT_VERSION` in `.github/workflows/ci.yaml`.

## API command coverage

CLI commands are written by hand. The operation catalog is generated from the
OpenAPI document included with the `sumup-go` version in `go.mod`. It lists
operation IDs, SDK clients and methods, HTTP paths, parameters, and request body metadata.

After updating `sumup-go`, regenerate the catalog:

```bash
make generate
```

Update the generator rather than editing `internal/apicommands/catalog.gen.go`
by hand. API commands link to an OpenAPI operation ID with `apicommands.Bind`
in their command definitions. Tests check that each supported operation has
exactly one command and that the catalog matches the SDK. CI also checks that
the generated catalog is up to date.

## Developer portal code samples

Generate the JSON code sample catalog for the developer portal:

```bash
VERSION=v0.1.0 make generate-codesamples
```

Use the CLI version the samples describe. The generator creates a sample for
each named request body example. Operations without named examples get one
default sample.

Named samples use an ID such as `CreateCheckout.HostedCheckout`. The
`operationId` field holds the operation ID, and `example` holds the example name.

The generator uses the pinned SDK's OpenAPI document by default. To preview
examples from another compatible document, pass `--spec`:

```bash
go run ./internal/cmd/generate-samples \
  --cli-version v0.1.0 \
  --spec ../apis/specs/openapi31.json \
  --out code-samples.json
```

The Make target writes to `code-samples.json` by default. Set `CODESAMPLES_OUT`
to change the path. The catalog uses the same versioned schema as the SDK
sample catalogs. Do not commit the generated file to this repository.

When a release is published, `.github/workflows/release-code-samples.yaml`
generates the catalog and opens or updates a pull request in
`sumup/sumup-developer` to update `src/codesamples/cli.json`.

## Pull request titles and commits

Pull request titles must follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
because they are used to generate the changelog. For example:
`fix: handle missing merchant codes`. Use the same format for individual commits
when possible.
