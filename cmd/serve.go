package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mustur/mockgrid/app/api"
	"github.com/mustur/mockgrid/app/api/store"
	"github.com/mustur/mockgrid/app/api/store/filesystem"
	"github.com/mustur/mockgrid/app/api/store/noop"
	"github.com/mustur/mockgrid/app/api/store/sqlite"
	"github.com/mustur/mockgrid/app/api/svc/sendmail"
	"github.com/mustur/mockgrid/app/config"
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

		msgStore, err := buildStore(cfg)
		if err != nil {
			return fmt.Errorf("initialize store: %w", err)
		}
		defer func() {
			if err := msgStore.Close(); err != nil {
				slog.Error("failed to close store", "err", err)
			}
		}()

		tpl := buildTemplater(cfg)
		listenAddr := fmt.Sprintf("%s:%d", cfg.MockgridHost, cfg.MockgridPort)

		// Build services
		mailSvc := sendmail.New(sendmail.Config{
			SMTPServer:    cfg.SMTPServer,
			SMTPPort:      cfg.SMTPPort,
			ListenAddr:    listenAddr,
			AttachmentDir: attachmentDir(cfg),
			AuthKey:       authKey(cfg),
			SMTPUser:      smtpUser(cfg),
			SMTPPass:      smtpPass(cfg),
		}, tpl, msgStore)

		// Create and start the server
		mg := api.New(listenAddr, mailSvc)

		slog.Info("starting mockgrid server", "address", listenAddr)
		cmd.SetContext(context.Background())
		return mg.Start()
	},
}

// buildTemplater creates the appropriate templater based on config.
func buildTemplater(cfg *config.Config) template.Templater {
	if cfg.Templates == nil {
		return template.NewBesteffortTemplate("", "", "")
	}
	switch cfg.Templates.Mode {
	case "local":
		return template.NewLocalTemplate(cfg.Templates.Directory)
	case "sendgrid":
		return template.NewSendGridTemplate(cfg.Templates.TemplateKey, "")
	default:
		return template.NewBesteffortTemplate(cfg.Templates.Directory, cfg.Templates.TemplateKey, "")
	}
}

// buildStore creates the appropriate message store based on config.
func buildStore(cfg *config.Config) (store.MessageStore, error) {
	if cfg.Storage == nil {
		return noop.New(), nil
	}

	switch cfg.Storage.Type {
	case "sqlite":
		if cfg.Storage.Path == "" {
			return nil, fmt.Errorf("sqlite storage requires a path")
		}
		return sqlite.New(cfg.Storage.Path)
	case "filesystem":
		if cfg.Storage.Path == "" {
			return nil, fmt.Errorf("filesystem storage requires a path")
		}
		return filesystem.New(cfg.Storage.Path)
	default:
		return noop.New(), nil
	}
}

// attachmentDir extracts the attachment directory from config.
func attachmentDir(cfg *config.Config) string {
	if cfg.Attachments != nil {
		return cfg.Attachments.Dir
	}
	return ""
}

// authKey extracts the auth key from config.
func authKey(cfg *config.Config) string {
	if cfg.Auth != nil {
		return cfg.Auth.SendgridKey
	}
	return ""
}

// smtpUser extracts the SMTP username from config.
func smtpUser(cfg *config.Config) string {
	if cfg.Auth != nil {
		return cfg.Auth.SMTPUser
	}
	return ""
}

// smtpPass extracts the SMTP password from config.
func smtpPass(cfg *config.Config) string {
	if cfg.Auth != nil {
		return cfg.Auth.SMTPPass
	}
	return ""
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
