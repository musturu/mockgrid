package template

import (
	"encoding/json"
	"fmt"
	"os"
)

type LocalTemplate struct {
	templateDir string
}

func NewLocalTemplate(templateDir string) *LocalTemplate {
	return &LocalTemplate{
		templateDir: templateDir,
	}
}

func (lt LocalTemplate) GetTemplate(templateID string) (*TemplateVersion, error) {
	filePath := fmt.Sprintf("%s/%s.html", lt.templateDir, templateID)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var tmplFile TemplateFile
	if err := json.Unmarshal(data, &tmplFile); err != nil {
		return nil, err
	}
	if len(tmplFile.Versions) == 0 {
		return nil, fmt.Errorf("no versions found in template file %s", filePath)
	}
	if len(tmplFile.Versions) == 1 {
		return &tmplFile.Versions[0], nil
	}

	for _, v := range tmplFile.Versions {
		if v.Active == 1 {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("no active versions found in template file %s", filePath)
}
