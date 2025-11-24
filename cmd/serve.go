package cmd

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/mustur/mockgrid/app/api"
	"github.com/mustur/mockgrid/app/template"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the mockgrid server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := configFromCmd(cmd)
		if cfg == nil {
			slog.Error("no config found in command context")
			return nil
		}

		// Print config values (values only) at startup
		pterm.Info.Println("Configuration values:")
		cfg.PrintValues()

		var tpl template.Templater
		if cfg.Templates == nil {
			tpl = template.NewBesteffortTemplate("", "", "")
		} else {
			switch cfg.Templates.Mode {
			case "local":
				tpl = template.NewLocalTemplate(cfg.Templates.Directory)
			case "sendgrid":
				tpl = template.NewSendGridTemplate(cfg.Templates.TemplateKey, "")
			default:
				tpl = template.NewBesteffortTemplate(cfg.Templates.Directory, cfg.Templates.TemplateKey, "")
			}
		}

		attachmentsDir := ""
		if cfg.Attachments != nil {
			attachmentsDir = cfg.Attachments.Dir
		}

		authKey := ""
		if cfg.Auth != nil {
			authKey = cfg.Auth.SendgridKey
		}

		mg := api.NewBuilder().WithTemplate(tpl).
			WithSMTP(cfg.SMTPServer, cfg.SMTPPort).
			WithListen(cfg.LocalSendGridHost, cfg.LocalSendgridPort).
			WithAttachments(attachmentsDir).
			WithAuth(authKey).
			Build()

		slog.Info("starting mockgrid server on " + cfg.LocalSendGridHost + ":" + strconv.Itoa(cfg.LocalSendgridPort))
		// Clean context of config
		cmd.SetContext(context.Background())
		cfg = nil
		return mg.Start()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
