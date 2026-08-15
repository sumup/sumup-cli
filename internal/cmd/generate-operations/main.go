package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const sdkModule = "github.com/sumup/sumup-go"

type moduleInfo struct {
	Path    string
	Version string
	Dir     string
	Replace *moduleInfo
}

type openAPIDocument struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
	Paths map[string]pathItem `json:"paths"`
}

type pathItem struct {
	Parameters []openAPIParameter `json:"parameters"`
	Delete     *openAPIOperation  `json:"delete"`
	Get        *openAPIOperation  `json:"get"`
	Patch      *openAPIOperation  `json:"patch"`
	Post       *openAPIOperation  `json:"post"`
	Put        *openAPIOperation  `json:"put"`
}

type openAPIOperation struct {
	OperationID string              `json:"operationId"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	Tags        []string            `json:"tags"`
	Parameters  []openAPIParameter  `json:"parameters"`
	RequestBody *openAPIRequestBody `json:"requestBody"`
	Codegen     struct {
		MethodName string `json:"method_name"`
	} `json:"x-codegen"`
}

type openAPIParameter struct {
	Name        string        `json:"name"`
	Location    string        `json:"in"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Schema      openAPISchema `json:"schema"`
}

type openAPIRequestBody struct {
	Required bool `json:"required"`
	Content  map[string]struct {
		Schema openAPISchema `json:"schema"`
	} `json:"content"`
}

type openAPISchema struct {
	Reference string `json:"$ref"`
	Type      string `json:"type"`
	Format    string `json:"format"`
}

type operation struct {
	ID          string
	Client      string
	SDKMethod   string
	HTTPMethod  string
	Path        string
	Summary     string
	Description string
	Parameters  []parameter
	RequestBody *requestBody
}

type parameter struct {
	Name        string
	Location    string
	Description string
	Type        string
	Format      string
	Required    bool
}

type requestBody struct {
	Schema   string
	Required bool
}

func main() {
	var outputPath string
	var specPath string
	var sdkVersion string
	flag.StringVar(&outputPath, "out", "catalog.gen.go", "generated Go output path")
	flag.StringVar(&specPath, "spec", "", "OpenAPI document path; defaults to the pinned SDK module")
	flag.StringVar(&sdkVersion, "sdk-version", "", "SDK version; defaults to the pinned module version")
	flag.Parse()

	if err := run(outputPath, specPath, sdkVersion); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "generate operations: %v\n", err)
		os.Exit(1)
	}
}

func run(outputPath, specPath, sdkVersion string) error {
	if specPath == "" || sdkVersion == "" {
		module, err := resolveModule()
		if err != nil {
			return err
		}
		if sdkVersion == "" {
			sdkVersion = module.Version
		}
		if specPath == "" {
			moduleDir := module.Dir
			if module.Replace != nil {
				moduleDir = module.Replace.Dir
			}
			if moduleDir == "" {
				return errors.New("pinned SDK module has no resolved directory")
			}
			specPath = filepath.Join(moduleDir, "openapi.json")
		}
	}

	spec, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read OpenAPI document: %w", err)
	}

	document, operations, err := parseOperations(spec)
	if err != nil {
		return err
	}

	generated, err := renderCatalog(document.Info.Version, sdkVersion, spec, operations)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, generated, 0o644); err != nil {
		return fmt.Errorf("write generated catalog: %w", err)
	}

	return nil
}

func resolveModule() (*moduleInfo, error) {
	command := exec.Command("go", "list", "-m", "-json", sdkModule)
	output, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("resolve pinned SDK module: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("resolve pinned SDK module: %w", err)
	}

	var module moduleInfo
	if err := json.Unmarshal(output, &module); err != nil {
		return nil, fmt.Errorf("decode pinned SDK module: %w", err)
	}
	if module.Path != sdkModule {
		return nil, fmt.Errorf("resolved unexpected SDK module %q", module.Path)
	}
	if module.Version == "" {
		return nil, errors.New("pinned SDK module has no version")
	}

	return &module, nil
}

func parseOperations(spec []byte) (*openAPIDocument, []operation, error) {
	var document openAPIDocument
	if err := json.Unmarshal(spec, &document); err != nil {
		return nil, nil, fmt.Errorf("decode OpenAPI document: %w", err)
	}
	if document.Info.Version == "" {
		return nil, nil, errors.New("OpenAPI document has no info.version")
	}

	paths := make([]string, 0, len(document.Paths))
	for path := range document.Paths {
		paths = append(paths, path)
	}
	slices.Sort(paths)

	operations := make([]operation, 0)
	for _, path := range paths {
		item := document.Paths[path]
		for _, candidate := range []struct {
			method    string
			operation *openAPIOperation
		}{
			{method: "DELETE", operation: item.Delete},
			{method: "GET", operation: item.Get},
			{method: "PATCH", operation: item.Patch},
			{method: "POST", operation: item.Post},
			{method: "PUT", operation: item.Put},
		} {
			if candidate.operation == nil {
				continue
			}

			parsed, err := parseOperation(path, candidate.method, item.Parameters, candidate.operation)
			if err != nil {
				return nil, nil, err
			}
			operations = append(operations, parsed)
		}
	}

	slices.SortFunc(operations, func(a, b operation) int {
		if result := strings.Compare(a.Client, b.Client); result != 0 {
			return result
		}
		if result := strings.Compare(a.SDKMethod, b.SDKMethod); result != 0 {
			return result
		}
		return strings.Compare(a.ID, b.ID)
	})

	return &document, operations, nil
}

func parseOperation(path, httpMethod string, pathParameters []openAPIParameter, source *openAPIOperation) (operation, error) {
	if source.OperationID == "" {
		return operation{}, fmt.Errorf("%s %s has no operationId", httpMethod, path)
	}
	if len(source.Tags) != 1 || source.Tags[0] == "" {
		return operation{}, fmt.Errorf("operation %q must have exactly one tag", source.OperationID)
	}
	if source.Codegen.MethodName == "" {
		return operation{}, fmt.Errorf("operation %q has no x-codegen.method_name", source.OperationID)
	}

	result := operation{
		ID:          source.OperationID,
		Client:      source.Tags[0],
		SDKMethod:   upperCamel(source.Codegen.MethodName),
		HTTPMethod:  httpMethod,
		Path:        path,
		Summary:     strings.TrimSpace(source.Summary),
		Description: strings.TrimSpace(source.Description),
	}
	for _, sourceParameter := range append(slices.Clone(pathParameters), source.Parameters...) {
		result.Parameters = append(result.Parameters, parameter{
			Name:        sourceParameter.Name,
			Location:    sourceParameter.Location,
			Description: strings.TrimSpace(sourceParameter.Description),
			Type:        schemaName(sourceParameter.Schema),
			Format:      sourceParameter.Schema.Format,
			Required:    sourceParameter.Required,
		})
	}
	if source.RequestBody != nil {
		body := requestBody{Required: source.RequestBody.Required}
		if mediaType, ok := source.RequestBody.Content["application/json"]; ok {
			body.Schema = schemaName(mediaType.Schema)
		}
		result.RequestBody = &body
	}

	return result, nil
}

func schemaName(schema openAPISchema) string {
	if schema.Reference != "" {
		parts := strings.Split(schema.Reference, "/")
		return parts[len(parts)-1]
	}
	return schema.Type
}

func upperCamel(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})
	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		result.WriteString(strings.ToUpper(part[:1]))
		result.WriteString(part[1:])
	}
	return result.String()
}

func renderCatalog(openAPIVersion, sdkVersion string, spec []byte, operations []operation) ([]byte, error) {
	hash := sha256.Sum256(spec)
	var output bytes.Buffer
	fmt.Fprintf(&output, "// Code generated by generate-operations from %s %s. DO NOT EDIT.\n", sdkModule, sdkVersion)
	fmt.Fprintf(&output, "// OpenAPI SHA-256: %s\n\n", hex.EncodeToString(hash[:]))
	output.WriteString("package apicommands\n\n")
	output.WriteString("const (\n")
	fmt.Fprintf(&output, "\tCatalogSDKModule = %q\n", sdkModule)
	fmt.Fprintf(&output, "\tCatalogSDKVersion = %q\n", sdkVersion)
	fmt.Fprintf(&output, "\tCatalogOpenAPIVersion = %q\n", openAPIVersion)
	fmt.Fprintf(&output, "\tCatalogSpecSHA256 = %q\n", hex.EncodeToString(hash[:]))
	output.WriteString(")\n\n")
	output.WriteString("var Operations = []Operation{\n")
	for _, operation := range operations {
		output.WriteString("\t{\n")
		fmt.Fprintf(&output, "\t\tID: %q,\n", operation.ID)
		fmt.Fprintf(&output, "\t\tClient: %q,\n", operation.Client)
		fmt.Fprintf(&output, "\t\tSDKMethod: %q,\n", operation.SDKMethod)
		fmt.Fprintf(&output, "\t\tHTTPMethod: %q,\n", operation.HTTPMethod)
		fmt.Fprintf(&output, "\t\tPath: %q,\n", operation.Path)
		fmt.Fprintf(&output, "\t\tSummary: %q,\n", operation.Summary)
		fmt.Fprintf(&output, "\t\tDescription: %q,\n", operation.Description)
		if len(operation.Parameters) > 0 {
			output.WriteString("\t\tParameters: []Parameter{\n")
			for _, parameter := range operation.Parameters {
				fmt.Fprintf(&output, "\t\t\t{Name: %q, Location: %q, Description: %q, Type: %q, Format: %q, Required: %t},\n",
					parameter.Name, parameter.Location, parameter.Description, parameter.Type, parameter.Format, parameter.Required)
			}
			output.WriteString("\t\t},\n")
		}
		if operation.RequestBody != nil {
			fmt.Fprintf(&output, "\t\tRequestBody: &RequestBody{Schema: %q, Required: %t},\n", operation.RequestBody.Schema, operation.RequestBody.Required)
		}
		output.WriteString("\t},\n")
	}
	output.WriteString("}\n")

	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated catalog: %w", err)
	}
	return formatted, nil
}
