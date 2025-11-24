package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/mustur/mockgrid/app/config"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// typed key to avoid collisions
type cfgKey struct{}

func configFromCmd(cmd *cobra.Command) *config.Config {
	if cmd == nil || cmd.Context() == nil {
		return nil
	}
	if v := cmd.Context().Value(cfgKey{}); v != nil {
		if c, ok := v.(*config.Config); ok {
			return c
		}
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "mockgrid",
	Short: "A mock email server that simulates SendGrid's API",
	Long: `Email Server is a mock email server designed to simulate SendGrid's API.
It allows you to test email sending functionality without actually sending emails.
It supports SMTP for sending emails and can render templates similar to SendGrid.`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {

		// Load config from env first (lowest priority)
		envCfg := config.LoadFromEnv()

		// Load config from file if provided
		fileCfg := &config.Config{}
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			pterm.Error.Println("Failed to read --config flag:", err)
			return err
		}
		if configPath != "" {
			loaded, err := config.LoadEmailServiceConfig(configPath)
			if err != nil {
				pterm.Error.Println("Failed to load configuration:", err)
				return err
			}
			fileCfg = loaded
		}

		// Load config from flags (highest priority)
		flagCfg := &config.Config{}
		// top-level flags
		if v, _ := cmd.Flags().GetString("smtp-server"); v != "" {
			flagCfg.SMTPServer = v
		}
		if v, _ := cmd.Flags().GetInt("smtp-port"); v != 0 {
			flagCfg.SMTPPort = v
		}
		if v, _ := cmd.Flags().GetString("local-sendgrid-host"); v != "" {
			flagCfg.LocalSendGridHost = v
		}
		if v, _ := cmd.Flags().GetInt("local-sendgrid-port"); v != 0 {
			flagCfg.LocalSendgridPort = v
		}

		// templates
		tmpl := &config.TemplateConfig{}
		anyT := false
		if v, _ := cmd.Flags().GetString("templates-mode"); v != "" {
			tmpl.Mode = v
			anyT = true
		}
		if v, _ := cmd.Flags().GetString("templates-directory"); v != "" {
			tmpl.Directory = v
			anyT = true
		}
		if v, _ := cmd.Flags().GetString("templates-key"); v != "" {
			tmpl.TemplateKey = v
			anyT = true
		}
		if anyT {
			flagCfg.Templates = tmpl
		}

		// attachments
		if v, _ := cmd.Flags().GetString("attachments-dir"); v != "" {
			flagCfg.Attachments = &config.AttachmentConfig{Dir: v}
		}

		// auth
		if v, _ := cmd.Flags().GetString("sendgrid-key"); v != "" {
			flagCfg.Auth = &config.Auth{SendgridKey: v}
		}

		// Merge order: envCfg <- fileCfg <- flagCfg
		merged := config.MergeConfig(envCfg, fileCfg)
		merged = config.MergeConfig(merged, flagCfg)

		merged.WithDefaults()
		if err := merged.ValidateConfig(); err != nil {
			return err
		}

		ctx := context.WithValue(cmd.Context(), cfgKey{}, merged)
		cmd.SetContext(ctx)

		if merged.Attachments != nil && merged.Attachments.Dir != "" {
			if err := os.MkdirAll(merged.Attachments.Dir, 0o750); err != nil {
				slog.Error("EmailServer.Start: failed to create attachment directory", "err", err)
				return err
			}
		}

		return nil
	},
	// accept cmd so we can read the attached config
	PersistentPostRun: func(cmd *cobra.Command, _ []string) {
		c := configFromCmd(cmd)
		if c == nil {
			slog.Info("no config in context; skipping attachment cleanup")
			return
		}
		if c.Attachments == nil || c.Attachments.Dir == "" {
			slog.Info("no attachments directory configured; skipping cleanup")
			return
		}
		if err := os.RemoveAll(c.Attachments.Dir); err != nil {
			slog.Error("EmailServer.Start: Failed to remove attachment directory", "err", err)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
	// config override flags
	rootCmd.PersistentFlags().String("smtp-server", "", "SMTP server hostname")
	rootCmd.PersistentFlags().Int("smtp-port", 0, "SMTP server port")
	rootCmd.PersistentFlags().String("local-sendgrid-host", "", "Local SendGrid host")
	rootCmd.PersistentFlags().Int("local-sendgrid-port", 0, "Local SendGrid port")
	rootCmd.PersistentFlags().String("templates-mode", "", "Templates mode: local|sendgrid|besteffort")
	rootCmd.PersistentFlags().String("templates-directory", "", "Local templates directory")
	rootCmd.PersistentFlags().String("templates-key", "", "Templates key for remote provider")
	rootCmd.PersistentFlags().String("attachments-dir", "", "Directory to store attachments")
	rootCmd.PersistentFlags().String("sendgrid-key", "", "Sendgrid API key")
	rootCmd.AddCommand(serveCmd)
}
