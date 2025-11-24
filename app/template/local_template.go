package template

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	// sanitize templateID to prevent directory traversal and ensure it resolves
	// under the configured templateDir
	safeID := filepath.Clean("/" + templateID) // prefix slash to force relative cleaning
	// ensure file has .html suffix
	if !strings.HasSuffix(safeID, ".html") {
		safeID = safeID + ".html"
	}
	filePath := filepath.Join(lt.templateDir, safeID)

	// ensure the resulting path is inside templateDir
	absDir, err := filepath.Abs(lt.templateDir)
	if err != nil {
		return nil, err
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("invalid template id: %s", templateID)
	}

	// read the file via an os.DirFS rooted at absDir to avoid direct file path access
	dirFS := os.DirFS(absDir)
	f, err := dirFS.Open(rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
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
