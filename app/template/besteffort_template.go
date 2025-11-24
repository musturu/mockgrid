package template

type BesteffortTemplate struct {
	LocalTemplate
	SendGridTemplate
}

func NewBesteffortTemplate(localDir, sendGridAPIKey string, sendGridURL string) *BesteffortTemplate {
	return &BesteffortTemplate{
		LocalTemplate:    *NewLocalTemplate(localDir),
		SendGridTemplate: *NewSendGridTemplate(sendGridAPIKey, sendGridURL),
	}
}

func (bt BesteffortTemplate) GetTemplate(templateID string) (*TemplateVersion, error) {
	tmpl, err := bt.LocalTemplate.GetTemplate(templateID)
	if err == nil {
		return tmpl, nil
	}

	return bt.SendGridTemplate.GetTemplate(templateID)
}
