package template

import (
	"fmt"
	"log/slog"

	"github.com/aymerick/raymond"
	"github.com/mustur/mockgrid/app/api/objects"
)

// TemplateVersion represents a version in the SendGrid template JSON
type TemplateVersion struct {
	Subject      string `json:"subject"`
	HtmlContent  string `json:"html_content"`
	PlainContent string `json:"plain_content"`
	Active       int    `json:"active"`
}

// TemplateFile represents the SendGrid template JSON
type TemplateFile struct {
	Versions []TemplateVersion `json:"versions"`
}

type Templater interface {
	GetTemplate(templateID string) (*TemplateVersion, error)
}

// RenderAndPopulateFromTemplate fetches and renders templates for each personalization.
func RenderAndPopulateFromTemplate(postRequest *objects.PostRequest, tpl Templater) error {
	templateID := postRequest.TemplateID
	if templateID == "" {
		slog.Warn("RenderAndPopulateFromTemplate - no template_id provided, skipping template rendering")
		return nil
	}
	for i, personalization := range postRequest.Personalizations {
		tmpl, err := tpl.GetTemplate(templateID)
		if err != nil {
			return fmt.Errorf("failed to fetch template %s: %w", templateID, err)
		}

		data := personalization.DynamicTemplateData
		render := func(tmplStr string) string {
			result, err := raymond.Render(tmplStr, data)
			if err != nil {
				return tmplStr
			}
			return result
		}

		subject := render(tmpl.Subject)
		htmlContent := render(tmpl.HtmlContent)
		plainContent := render(tmpl.PlainContent)

		postRequest.Personalizations[i].Subject = subject

		if len(postRequest.Content) == 0 {
			if htmlContent != "" {
				postRequest.Content = append(postRequest.Content, struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				}{
					Type:  "text/html",
					Value: htmlContent,
				})
			}
			if plainContent != "" {
				postRequest.Content = append(postRequest.Content, struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				}{
					Type:  "text/plain",
					Value: plainContent,
				})
			}
		}
	}
	return nil
}
