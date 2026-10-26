package readers

import (
	"context"
	"fmt"
	"regexp"

	"github.com/urfave/cli/v3"

	sumup "github.com/sumup/sumup-go"
	"github.com/sumup/sumup-go/secret"

	"github.com/sumup/sumup-cli/internal/apicommands"
	"github.com/sumup/sumup-cli/internal/app"
	"github.com/sumup/sumup-cli/internal/commands/util"
	"github.com/sumup/sumup-cli/internal/display"
	"github.com/sumup/sumup-cli/internal/display/attribute"
)

var menuLockPasscodePattern = regexp.MustCompile(`^[0-9]{4}$`)

func settingsCommand() *cli.Command {
	return &cli.Command{
		Name:  "settings",
		Usage: "Manage Solo reader settings.",
		Commands: []*cli.Command{
			apicommands.Bind("GetReaderSettings", &cli.Command{
				Name:      "get",
				Usage:     "Get current settings, or the last saved settings if the reader is offline.",
				ArgsUsage: "<reader-id>",
				Action:    getReaderSettings,
				Flags:     readerSettingsFlags(),
			}),
			apicommands.Bind("UpdateReaderSettings", &cli.Command{
				Name:      "update",
				Usage:     "Update menu lock settings on an online Solo reader.",
				ArgsUsage: "<reader-id>",
				Action:    updateReaderSettings,
				Flags: append(readerSettingsFlags(),
					&cli.BoolFlag{
						Name:  "menu-lock-enabled",
						Usage: "Enable menu lock; use --menu-lock-enabled=false to disable it.",
					},
					&cli.StringFlag{
						Name:    "menu-lock-passcode",
						Usage:   "Four-digit PIN for unlocking the reader menu.",
						Sources: cli.EnvVars("SUMUP_READER_MENU_LOCK_PASSCODE"),
					},
				),
			}),
		},
	}
}

func readerSettingsFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "merchant-code",
			Usage:   "Merchant code that owns the reader. Falls back to context.",
			Sources: cli.EnvVars("SUMUP_MERCHANT_CODE"),
		},
	}
}

func getReaderSettings(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}
	response, err := appCtx.Client.Readers.GetReaderSettings(ctx, merchantCode, readerID)
	if err != nil {
		return fmt.Errorf("get reader settings: %w", err)
	}
	if appCtx.JSONOutput {
		return display.PrintJSON(appCtx.Output, response)
	}
	return display.DataList(appCtx.Output, []attribute.KeyValue{
		attribute.ID(readerID),
		attribute.Optional("Menu Lock Enabled", response.Data.MenuLock.Enabled),
	})
}

func updateReaderSettings(ctx context.Context, cmd *cli.Command) error {
	appCtx, err := app.GetAppContext(cmd)
	if err != nil {
		return err
	}
	merchantCode, err := app.GetMerchantCode(cmd, "merchant-code")
	if err != nil {
		return err
	}
	readerID, err := util.RequireSingleArg(cmd, "reader ID")
	if err != nil {
		return err
	}
	if !cmd.IsSet("menu-lock-enabled") && !cmd.IsSet("menu-lock-passcode") {
		return fmt.Errorf("provide --menu-lock-enabled or --menu-lock-passcode")
	}
	body := sumup.ReadersUpdateReaderSettingsParams{}
	if cmd.IsSet("menu-lock-enabled") {
		body.MenuLock.Enabled = new(cmd.Bool("menu-lock-enabled"))
	}
	if cmd.IsSet("menu-lock-passcode") {
		passcode := cmd.String("menu-lock-passcode")
		if !menuLockPasscodePattern.MatchString(passcode) {
			return fmt.Errorf("menu lock passcode must contain exactly four digits")
		}
		body.MenuLock.Passcode = new(secret.New(passcode))
	}
	response, err := appCtx.Client.Readers.UpdateReaderSettings(ctx, merchantCode, readerID, body)
	if err != nil {
		return fmt.Errorf("update reader settings: %w", err)
	}
	return display.RenderMutation(appCtx.Output, appCtx.StatusOutput, appCtx.JSONOutput, display.MutationResult{
		JSONValue:      response,
		SuccessMessage: "Reader settings updated",
		Details: []attribute.KeyValue{
			attribute.ID(readerID),
			attribute.Attribute("Status", attribute.Styled(response.Data.Status)),
			attribute.Optional("Menu Lock Enabled", response.Data.MenuLock.Enabled),
		},
	})
}
