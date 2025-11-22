package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	sumup "github.com/sumup/sumup-go"
	sumupclient "github.com/sumup/sumup-go/client"
	"github.com/urfave/cli/v3"

	"github.com/sumup/sumup-cli/internal/config"
)

// ContextKey is used to store the initialized context in the CLI metadata map.
const ContextKey = "app-context"

// Context carries shared dependencies for commands.
type Context struct {
	Client          *sumup.Client
	JSONOutput      bool
	ExactTimestamps bool
	Locale          string
}

// NewContext constructs the CLI context with an initialized SumUp API client.
func NewContext(apiKey, baseURL string, jsonOutput bool, exactTimestamps bool) (*Context, error) {
	var opts []sumupclient.ClientOption
	if apiKey != "" {
		opts = append(opts, sumupclient.WithAPIKey(apiKey))
	}
	if baseURL != "" {
		opts = append(opts, sumupclient.WithBaseURL(baseURL))
	}

	client := sumup.NewClient(opts...)
	return &Context{
		Client:          client,
		JSONOutput:      jsonOutput,
		ExactTimestamps: exactTimestamps,
		Locale:          detectLocale(),
	}, nil
}

func detectLocale() string {
	envs := []string{"LC_ALL", "LC_TIME", "LANG"}
	for _, key := range envs {
		if value := normalizeLocale(os.Getenv(key)); value != "" {
			return value
		}
	}
	return "en"
}

func normalizeLocale(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, "."); idx != -1 {
		value = value[:idx]
	}
	if idx := strings.Index(value, "@"); idx != -1 {
		value = value[:idx]
	}
	value = strings.ReplaceAll(value, "_", "-")
	return value
}

// GetAppContext retrieves the application context from the CLI command metadata.
func GetAppContext(cmd *cli.Command) (*Context, error) {
	raw := cmd.Root().Metadata[ContextKey]
	appCtx, ok := raw.(*Context)
	if !ok || appCtx == nil {
		return nil, errors.New("application context has not been initialized")
	}
	return appCtx, nil
}

// GetMerchantCode retrieves the merchant code from the command flag,
// falling back to the stored context if the flag is not set.
func GetMerchantCode(cmd *cli.Command, flagName string) (string, error) {
	if cmd.IsSet(flagName) {
		return cmd.String(flagName), nil
	}

	merchantCode, err := config.GetCurrentMerchantCode()
	if err != nil {
		return "", fmt.Errorf("failed to load merchant context: %w", err)
	}

	if merchantCode == "" {
		return "", errors.New("merchant code is required. Provide --merchant-code flag or set context with 'sumup context set'")
	}

	return merchantCode, nil
}
