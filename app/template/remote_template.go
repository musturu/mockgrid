package template

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SendGridTemplate struct {
	sendgridKey string
	sendgridURL string
	client      *http.Client
}

func NewSendGridTemplate(sendgridKey string, sendgridURL string) *SendGridTemplate {
	return newSendGridTemplate(sendgridKey, sendgridURL, nil)
}

func newSendGridTemplate(sendgridKey string, sendgridURL string, client *http.Client) *SendGridTemplate {
	if sendgridURL == "" {
		sendgridURL = "https://api.sendgrid.com/v3/templates/"
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &SendGridTemplate{
		sendgridKey: sendgridKey,
		sendgridURL: sendgridURL,
		client:      client,
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

	resp, err := sgt.client.Do(req)
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
