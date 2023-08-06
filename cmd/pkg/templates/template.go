package templates

import (
	"os"
	"path/filepath"
	"text/template"
)

func Generate(t *template.Template, exampleFilePath string, name string, data any) error {
	err := os.MkdirAll(filepath.Dir(exampleFilePath), 0o777)
	if err != nil {
		return err
	}
	exampleFile, _ := os.Create(exampleFilePath)
	defer exampleFile.Close()

	err = t.ExecuteTemplate(exampleFile, name, data)
	if err != nil {
		return err
	}
	return nil
}
