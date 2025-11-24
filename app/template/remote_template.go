package template

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SendGridTemplate struct {
	sendgridKey string
	sendgridURL string
}

func NewSendGridTemplate(sendgridKey string, sendgridURL string) *SendGridTemplate {
	if sendgridURL == "" {
		sendgridURL = "https://api.sendgrid.com/v3/templates/"
	}

	return &SendGridTemplate{
		sendgridKey: sendgridKey,
		sendgridURL: sendgridURL,
	}
}

func (sgt SendGridTemplate) GetTemplate(templateID string) (*TemplateVersion, error) {
	url := sgt.sendgridURL + templateID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+sgt.sendgridKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sendgrid API error: %s", string(body))
	}

	var tmplFile TemplateFile
	if err := json.NewDecoder(resp.Body).Decode(&tmplFile); err != nil {
		return nil, err
	}
	if len(tmplFile.Versions) == 0 {
		return nil, fmt.Errorf("no versions found in template ID %s", templateID)
	}
	if len(tmplFile.Versions) == 1 {
		return &tmplFile.Versions[0], nil
	}

	for _, v := range tmplFile.Versions {
		if v.Active == 1 {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("no active versions found in template ID %s", templateID)
}
