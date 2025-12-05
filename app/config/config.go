package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

// Config holds all configuration values for the EmailServer.
type Config struct {
	SMTPServer        string            `yaml:"smtp_server"`
	SMTPPort          int               `yaml:"smtp_port"`
	LocalSendGridHost string            `yaml:"local_sendgrid_host"`
	LocalSendgridPort int               `yaml:"local_sendgrid_port"`
	Templates         *TemplateConfig   `yaml:"templates"`
	Attachments       *AttachmentConfig `yaml:"attachments"`
	Auth              *Auth             `yaml:"auth"`
	Storage           *StorageConfig    `yaml:"storage"`
}

type TemplateConfig struct {
	Mode        string // "local", "sendgrid", "besteffort"
	Directory   string
	TemplateKey string
}

type Auth struct {
	SendgridKey string `yaml:"sendgrid_key"`
	SMTPUser    string `yaml:"smtp_user"`
	SMTPPass    string `yaml:"smtp_pass"`
}

type AttachmentConfig struct {
	Dir string `yaml:"dir"`
}

// StorageConfig holds configuration for message persistence.
type StorageConfig struct {
	Type string `yaml:"type"` // "none", "sqlite", "filesystem"
	Path string `yaml:"path"` // path to sqlite db or filesystem directory
}

func LoadEmailServiceConfig(path string) (*Config, error) {

	var cfg Config

	cleanPath := filepath.Clean(path)
	// ensure path exists and is a file
	st, err := os.Stat(cleanPath)
	if err != nil {
		return nil, err
	}
	if st.IsDir() {
		return nil, errors.New("config path is a directory")
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) WithDefaults() {

	if cfg.LocalSendGridHost == "" {
		cfg.LocalSendGridHost = "0.0.0.0"
	}
	if cfg.LocalSendgridPort == 0 {
		cfg.LocalSendgridPort = 5900
	}
	if cfg.Attachments != nil && cfg.Attachments.Dir == "" {
		cfg.Attachments.Dir = "./attachments"
	}
	if cfg.SMTPServer == "" {
		cfg.SMTPServer = "localhost"
	}
	if cfg.SMTPPort == 0 {
		cfg.SMTPPort = 587
	}
	if cfg.Storage == nil {
		cfg.Storage = &StorageConfig{Type: "none"}
	}
}

func (c *Config) ValidateConfig() error {

	if c.SMTPServer == "" {
		return errors.New("SMTP server is not configured")
	}
	if c.Attachments == nil || c.Attachments.Dir == "" {
		pterm.Warning.Println("Attachment directory is not configured, skipping attachment handling")
	}

	templateDir := ""
	sendgridKey := ""
	mode := ""

	if c.Templates != nil {
		templateDir = c.Templates.Directory
		mode = c.Templates.Mode
	}
	if c.Auth != nil {
		sendgridKey = c.Auth.SendgridKey
	}

	switch mode {
	case "local":
		if templateDir == "" {
			return errors.New("mode is 'local' but template directory is not configured")
		}
		if stat, err := os.Stat(templateDir); err != nil || !stat.IsDir() {
			return errors.New("template directory does not exist or is not a directory")
		}
		pterm.Info.Println("Using local template directory for templates.")
	case "sendgrid":
		if sendgridKey == "" {
			return errors.New("mode is 'sendgrid' but Sendgrid key is not configured")
		}
		pterm.Info.Println("Using Sendgrid for templates.")
	case "besteffort":
		if templateDir != "" {
			if stat, err := os.Stat(templateDir); err == nil && stat.IsDir() {
				pterm.Info.Println("Using local template directory for templates (besteffort mode).")
			} else {
				pterm.Warning.Println("Local template directory not available, will try Sendgrid if key is present (besteffort mode).")
				if sendgridKey == "" {
					pterm.Warning.Println("Sendgrid key is also not configured, skipping template handling (besteffort mode).")
				}
			}
		} else if sendgridKey != "" {
			pterm.Info.Println("Using Sendgrid for templates (besteffort mode).")
		} else {
			pterm.Warning.Println("Neither template directory nor Sendgrid key are configured, skipping template handling (besteffort mode).")
		}
	default:
		if templateDir == "" && sendgridKey == "" {
			pterm.Warning.Println("Template directory and Sendgrid key are not configured, skipping template handling")
		}
		if templateDir != "" {
			if stat, err := os.Stat(templateDir); err != nil || !stat.IsDir() {
				return errors.New("template directory does not exist or is not a directory")
			}
		}
		if templateDir != "" && sendgridKey != "" {
			pterm.Info.Println("Template directory is configured as well as Sendgrid key, prioritizing local templates")
		}
	}

	return nil
}

// PrintValues prints the current configuration values (values only, no keys)
// Each configuration field is printed as its value on its own line. This is
// intended for human-readable startup logs and intentionally omits keys.
func (c *Config) PrintValues() {
	if c == nil {
		pterm.Info.Println("(no config)")
		return
	}

	// top-level scalar values
	pterm.Info.Println("SMTP Server:", c.SMTPServer)
	pterm.Info.Println("SMTP Port:", strconv.Itoa(c.SMTPPort))
	pterm.Info.Println("Local SendGrid Host:", c.LocalSendGridHost)
	pterm.Info.Println("Local SendGrid Port:", strconv.Itoa(c.LocalSendgridPort))

	// templates
	if c.Templates != nil {
		pterm.Info.Println("Templates Mode:", c.Templates.Mode)
		pterm.Info.Println("Templates Directory:", c.Templates.Directory)
		pterm.Info.Println("Templates Key:", maskSecret(c.Templates.TemplateKey))
	}

	// attachments
	if c.Attachments != nil {
		pterm.Info.Println("Attachments Directory:", c.Attachments.Dir)
	}

	// auth
	if c.Auth != nil {
		pterm.Info.Println("Auth Sendgrid Key:", maskSecret(c.Auth.SendgridKey))
		pterm.Info.Println("Auth SMTP User:", c.Auth.SMTPUser)
		pterm.Info.Println("Auth SMTP Pass:", maskSecret(c.Auth.SMTPPass))
	}

	// storage
	if c.Storage != nil {
		pterm.Info.Println("Storage Type:", c.Storage.Type)
		pterm.Info.Println("Storage Path:", c.Storage.Path)
	}
}

// maskSecret masks a secret string leaving first/last 4 characters visible when
// the string is long enough, otherwise replaces the whole string with asterisks.
func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return strings.Repeat("*", 8)
}

// LoadFromEnv constructs a Config by reading environment variables.
// It only sets values that are present in the environment; zero values
// indicate absence and can be overridden by a config file or flags.
func LoadFromEnv() *Config {
	cfg := &Config{}

	if v := os.Getenv("SMTP_SERVER"); v != "" {
		cfg.SMTPServer = v
	}
	if v := os.Getenv("SMTP_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.SMTPPort = i
		}
	}
	if v := os.Getenv("MOCKGRID_HOST"); v != "" {
		cfg.LocalSendGridHost = v
	}
	if v := os.Getenv("MOCKGRID_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.LocalSendgridPort = i
		}
	}

	// Templates
	var t TemplateConfig
	anyT := false
	if v := os.Getenv("TEMPLATES_MODE"); v != "" {
		t.Mode = v
		anyT = true
	}
	if v := os.Getenv("TEMPLATES_DIRECTORY"); v != "" {
		t.Directory = v
		anyT = true
	}
	if v := os.Getenv("TEMPLATES_SG_KEY"); v != "" {
		t.TemplateKey = v
		anyT = true
	}
	if anyT {
		cfg.Templates = &t
	}

	// Attachments
	if v := os.Getenv("ATTACHMENTS_DIR"); v != "" {
		cfg.Attachments = &AttachmentConfig{Dir: v}
	}

	// Auth
	var auth Auth
	anyAuth := false
	if v := os.Getenv("SENDGRID_KEY"); v != "" {
		auth.SendgridKey = v
		anyAuth = true
	}
	if v := os.Getenv("SMTP_USER"); v != "" {
		auth.SMTPUser = v
		anyAuth = true
	}
	if v := os.Getenv("SMTP_PASS"); v != "" {
		auth.SMTPPass = v
		anyAuth = true
	}
	if anyAuth {
		cfg.Auth = &auth
	}

	// Storage
	var storage StorageConfig
	anyStorage := false
	if v := os.Getenv("STORAGE_TYPE"); v != "" {
		storage.Type = v
		anyStorage = true
	}
	if v := os.Getenv("STORAGE_PATH"); v != "" {
		storage.Path = v
		anyStorage = true
	}
	if anyStorage {
		cfg.Storage = &storage
	}

	return cfg
}

// MergeConfig overlays non-zero values from 'over' onto 'base'.
// Values in 'over' take precedence when set (non-empty string or non-zero int).
func MergeConfig(base *Config, over *Config) *Config {
	if base == nil {
		if over == nil {
			return &Config{}
		}
		return over
	}
	if over == nil {
		return base
	}

	if over.SMTPServer != "" {
		base.SMTPServer = over.SMTPServer
	}
	if over.SMTPPort != 0 {
		base.SMTPPort = over.SMTPPort
	}
	if over.LocalSendGridHost != "" {
		base.LocalSendGridHost = over.LocalSendGridHost
	}
	if over.LocalSendgridPort != 0 {
		base.LocalSendgridPort = over.LocalSendgridPort
	}

	// Templates
	if over.Templates != nil {
		if base.Templates == nil {
			base.Templates = &TemplateConfig{}
		}
		if over.Templates.Mode != "" {
			base.Templates.Mode = over.Templates.Mode
		}
		if over.Templates.Directory != "" {
			base.Templates.Directory = over.Templates.Directory
		}
		if over.Templates.TemplateKey != "" {
			base.Templates.TemplateKey = over.Templates.TemplateKey
		}
	}

	// Attachments
	if over.Attachments != nil && over.Attachments.Dir != "" {
		if base.Attachments == nil {
			base.Attachments = &AttachmentConfig{}
		}
		base.Attachments.Dir = over.Attachments.Dir
	}

	// Auth
	if over.Auth != nil {
		if base.Auth == nil {
			base.Auth = &Auth{}
		}
		if over.Auth.SendgridKey != "" {
			base.Auth.SendgridKey = over.Auth.SendgridKey
		}
		if over.Auth.SMTPUser != "" {
			base.Auth.SMTPUser = over.Auth.SMTPUser
		}
		if over.Auth.SMTPPass != "" {
			base.Auth.SMTPPass = over.Auth.SMTPPass
		}
	}

	// Storage
	if over.Storage != nil {
		if base.Storage == nil {
			base.Storage = &StorageConfig{}
		}
		if over.Storage.Type != "" {
			base.Storage.Type = over.Storage.Type
		}
		if over.Storage.Path != "" {
			base.Storage.Path = over.Storage.Path
		}
	}

	return base
}
