package codesamples

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const sdkModule = "github.com/sumup/sumup-go"

type openAPIDocument struct {
	Paths      map[string]*pathItem `json:"paths"`
	Components struct {
		Schemas    map[string]*specSchema `json:"schemas"`
		Parameters map[string]*parameter  `json:"parameters"`
	} `json:"components"`
}

type pathItem struct {
	Parameters []parameter `json:"parameters"`
	Delete     *operation  `json:"delete"`
	Get        *operation  `json:"get"`
	Patch      *operation  `json:"patch"`
	Post       *operation  `json:"post"`
	Put        *operation  `json:"put"`
}

type operation struct {
	Parameters  []parameter  `json:"parameters"`
	RequestBody *requestBody `json:"requestBody"`
}

type requestBody struct {
	Content map[string]*mediaType `json:"content"`
}

type mediaType struct {
	Schema   *specSchema              `json:"schema"`
	Example  json.RawMessage          `json:"example"`
	Examples map[string]*namedExample `json:"examples"`
}

type namedExample struct {
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Value       json.RawMessage `json:"value"`
}

type parameter struct {
	Ref      string                   `json:"$ref"`
	Name     string                   `json:"name"`
	Location string                   `json:"in"`
	Required bool                     `json:"required"`
	Example  json.RawMessage          `json:"example"`
	Examples map[string]*namedExample `json:"examples"`
	Schema   *specSchema              `json:"schema"`
}

type specSchema struct {
	Ref        string                 `json:"$ref"`
	Type       json.RawMessage        `json:"type"`
	Format     string                 `json:"format"`
	Example    json.RawMessage        `json:"example"`
	Examples   []json.RawMessage      `json:"examples"`
	Default    json.RawMessage        `json:"default"`
	Enum       []json.RawMessage      `json:"enum"`
	Properties map[string]*specSchema `json:"properties"`
	Items      *specSchema            `json:"items"`
	AllOf      []*specSchema          `json:"allOf"`
	OneOf      []*specSchema          `json:"oneOf"`
	AnyOf      []*specSchema          `json:"anyOf"`
}

type moduleInfo struct {
	Path    string
	Version string
	Dir     string
	Replace *moduleInfo
}

type operationExample struct {
	name         string
	summary      string
	description  string
	body         map[string]any
	bodyProvided bool
	bodySchema   *specSchema
	parameters   []*parameter
}

func loadPinnedSpec() (*openAPIDocument, error) {
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
	moduleDir := module.Dir
	if module.Replace != nil {
		moduleDir = module.Replace.Dir
	}
	if moduleDir == "" {
		return nil, errors.New("pinned SDK module has no resolved directory")
	}

	return parseSpecFile(filepath.Join(moduleDir, "openapi.json"))
}

func parseSpecFile(path string) (*openAPIDocument, error) {
	spec, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read OpenAPI document: %w", err)
	}
	return parseSpec(spec)
}

func parseSpec(spec []byte) (*openAPIDocument, error) {
	var document openAPIDocument
	if err := json.Unmarshal(spec, &document); err != nil {
		return nil, fmt.Errorf("decode OpenAPI document: %w", err)
	}
	return &document, nil
}

func (document *openAPIDocument) exampleFor(httpMethod, apiPath string) operationExample {
	item := document.Paths[apiPath]
	if item == nil {
		return operationExample{}
	}

	source := item.operation(httpMethod)
	parameters := document.resolveParameters(append(slices.Clone(item.Parameters), operationParameters(source)...))
	example := operationExample{parameters: parameters}
	if source == nil || source.RequestBody == nil {
		return example
	}

	mediaType, ok := jsonMediaType(source.RequestBody.Content)
	if !ok || mediaType == nil {
		return example
	}
	example.bodySchema = mediaType.Schema

	if len(mediaType.Examples) > 0 {
		names := make([]string, 0, len(mediaType.Examples))
		for name := range mediaType.Examples {
			names = append(names, name)
		}
		slices.Sort(names)
		for _, name := range names {
			named := mediaType.Examples[name]
			if named == nil {
				continue
			}
			value, provided := decodeRaw(named.Value)
			if !provided {
				continue
			}
			example.name = name
			example.summary = strings.TrimSpace(named.Summary)
			example.description = strings.TrimSpace(named.Description)
			example.body, example.bodyProvided = objectValue(value)
			return example
		}
	}

	if value, provided := decodeRaw(mediaType.Example); provided {
		example.body, example.bodyProvided = objectValue(value)
		return example
	}
	if value, provided := document.schemaExample(mediaType.Schema, nil); provided {
		example.body, example.bodyProvided = objectValue(value)
		if example.bodyProvided {
			return example
		}
	}
	example.body = document.objectFromSchema(mediaType.Schema, nil)
	return example
}

func (item *pathItem) operation(httpMethod string) *operation {
	if item == nil {
		return nil
	}
	switch strings.ToUpper(httpMethod) {
	case "DELETE":
		return item.Delete
	case "GET":
		return item.Get
	case "PATCH":
		return item.Patch
	case "POST":
		return item.Post
	case "PUT":
		return item.Put
	default:
		return nil
	}
}

func operationParameters(source *operation) []parameter {
	if source == nil {
		return nil
	}
	return source.Parameters
}

func jsonMediaType(content map[string]*mediaType) (*mediaType, bool) {
	if content == nil {
		return nil, false
	}
	if mediaType, ok := content["application/json"]; ok {
		return mediaType, true
	}
	return nil, false
}

func (document *openAPIDocument) resolveParameters(parameters []parameter) []*parameter {
	resolved := make([]*parameter, 0, len(parameters))
	for i := range parameters {
		parameter := document.resolveParameter(&parameters[i], nil)
		if parameter != nil {
			resolved = append(resolved, parameter)
		}
	}
	return resolved
}

func (document *openAPIDocument) resolveParameter(parameter *parameter, seen map[string]struct{}) *parameter {
	if parameter == nil {
		return nil
	}
	if parameter.Ref == "" {
		return parameter
	}
	name := componentName(parameter.Ref)
	if name == "" || document.Components.Parameters == nil {
		return parameter
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	if _, ok := seen[parameter.Ref]; ok {
		return parameter
	}
	seen[parameter.Ref] = struct{}{}
	target, ok := document.Components.Parameters[name]
	if !ok {
		return parameter
	}
	return document.resolveParameter(target, seen)
}

func (document *openAPIDocument) resolveSchema(schema *specSchema, seen map[string]struct{}) *specSchema {
	if schema == nil {
		return nil
	}
	if schema.Ref == "" {
		return schema
	}
	name := componentName(schema.Ref)
	if name == "" || document.Components.Schemas == nil {
		return schema
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	if _, ok := seen[schema.Ref]; ok {
		return schema
	}
	seen[schema.Ref] = struct{}{}
	target, ok := document.Components.Schemas[name]
	if !ok {
		return schema
	}
	return document.resolveSchema(target, seen)
}

func (document *openAPIDocument) schemaExample(schema *specSchema, seen map[string]struct{}) (any, bool) {
	schema = document.resolveSchema(schema, seen)
	if schema == nil {
		return nil, false
	}
	if value, ok := decodeRaw(schema.Example); ok {
		return value, true
	}
	for _, example := range schema.Examples {
		if value, ok := decodeRaw(example); ok {
			return value, true
		}
	}
	if value, ok := decodeRaw(schema.Default); ok {
		return value, true
	}
	if len(schema.Enum) > 0 {
		return decodeRaw(schema.Enum[0])
	}
	for _, candidate := range slices.Concat(schema.AllOf, schema.OneOf, schema.AnyOf) {
		if value, ok := document.schemaExample(candidate, seen); ok {
			return value, true
		}
	}
	return nil, false
}

func (document *openAPIDocument) schemaFallback(schema *specSchema) any {
	schema = document.resolveSchema(schema, nil)
	if schema == nil {
		return nil
	}
	if value, ok := document.schemaExample(schema, nil); ok {
		return value
	}
	types := schema.types()
	switch {
	case slices.Contains(types, "string"):
		switch schema.Format {
		case "date-time":
			return "2025-01-01T00:00:00Z"
		case "date":
			return "2025-01-01"
		case "time":
			return "12:00:00"
		case "email":
			return "developer@example.com"
		case "uri", "url":
			return "https://example.com"
		case "uuid":
			return "00000000-0000-4000-8000-000000000000"
		case "hostname":
			return "example.com"
		case "password":
			return "secret"
		default:
			return "string"
		}
	case slices.Contains(types, "integer"):
		return 1
	case slices.Contains(types, "number"):
		return 1.0
	case slices.Contains(types, "boolean"):
		return true
	case slices.Contains(types, "array"):
		if item, ok := document.schemaExample(schema.Items, nil); ok {
			return []any{item}
		}
		if fallback := document.schemaFallback(schema.Items); fallback != nil {
			return []any{fallback}
		}
	}
	return nil
}

func (schema *specSchema) types() []string {
	if schema == nil || len(schema.Type) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(schema.Type, &value); err != nil {
		return nil
	}
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []any:
		types := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if ok {
				types = append(types, text)
			}
		}
		return types
	default:
		return nil
	}
}

func (document *openAPIDocument) objectFromSchema(schema *specSchema, seen map[string]struct{}) map[string]any {
	schema = document.resolveSchema(schema, seen)
	if schema == nil {
		return nil
	}
	if value, ok := document.schemaExample(schema, seen); ok {
		if object, ok := objectValue(value); ok {
			return object
		}
	}

	result := make(map[string]any)
	for name, property := range document.directProperties(schema, seen) {
		if value, ok := document.schemaExample(property, seen); ok {
			result[name] = value
			continue
		}
		if nested := document.objectFromSchema(property, seen); len(nested) > 0 {
			result[name] = nested
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func (document *openAPIDocument) directProperties(schema *specSchema, seen map[string]struct{}) map[string]*specSchema {
	schema = document.resolveSchema(schema, seen)
	if schema == nil {
		return nil
	}
	result := make(map[string]*specSchema)
	for _, candidate := range slices.Concat([]*specSchema{schema}, schema.AllOf, schema.OneOf, schema.AnyOf) {
		resolved := document.resolveSchema(candidate, seen)
		if resolved == nil {
			continue
		}
		for name, property := range resolved.Properties {
			result[name] = property
		}
	}
	return result
}

func (document *openAPIDocument) propertySchema(schema *specSchema, flag string) *specSchema {
	properties := document.properties(schema, nil)
	if len(properties) == 0 {
		return nil
	}
	paths := make([][]string, 0, len(properties))
	for path := range properties {
		paths = append(paths, strings.Split(path, "."))
	}
	matched, ok := bestPath(flag, paths)
	if !ok {
		return nil
	}
	return properties[strings.Join(matched, ".")]
}

func (document *openAPIDocument) properties(schema *specSchema, seen map[string]struct{}) map[string]*specSchema {
	schema = document.resolveSchema(schema, seen)
	if schema == nil {
		return nil
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	result := make(map[string]*specSchema)
	for _, candidate := range slices.Concat([]*specSchema{schema}, schema.AllOf, schema.OneOf, schema.AnyOf) {
		resolved := document.resolveSchema(candidate, seen)
		if resolved == nil {
			continue
		}
		for name, property := range resolved.Properties {
			result[name] = property
			for nestedName, nested := range document.properties(property, seen) {
				result[name+"."+nestedName] = nested
			}
		}
	}
	return result
}

func (example operationExample) parameter(name string) *parameter {
	want := normalizeName(name)
	var best *parameter
	bestScore := 0
	for _, parameter := range example.parameters {
		got := normalizeName(parameter.Name)
		score := 0
		switch {
		case got == want:
			score = 400
		case got == want+"s", want == got+"s":
			score = 300
		case strings.HasSuffix(want, "_"+got):
			score = 200 - len(got)
		case strings.HasSuffix(got, "_"+want):
			score = 100 - len(got)
		}
		if score > bestScore {
			bestScore = score
			best = parameter
		}
	}
	return best
}

func (example operationExample) parameterExample(document *openAPIDocument, name string) (any, bool) {
	parameter := example.parameter(name)
	if parameter == nil {
		return nil, false
	}
	if value, ok := decodeRaw(parameter.Example); ok {
		return value, true
	}
	if parameter.Examples != nil {
		names := make([]string, 0, len(parameter.Examples))
		for exampleName := range parameter.Examples {
			names = append(names, exampleName)
		}
		slices.Sort(names)
		for _, exampleName := range names {
			named := parameter.Examples[exampleName]
			if named == nil {
				continue
			}
			if value, ok := decodeRaw(named.Value); ok {
				return value, true
			}
		}
	}
	return document.schemaExample(parameter.Schema, nil)
}

func decodeRaw(raw json.RawMessage) (any, bool) {
	if len(raw) == 0 {
		return nil, false
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, false
	}
	return value, true
}

func objectValue(value any) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	return object, true
}

func componentName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func normalizeName(name string) string {
	name = strings.TrimSuffix(name, "[]")
	return strings.ReplaceAll(name, "-", "_")
}
