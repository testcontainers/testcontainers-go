package template

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

// Generate writes the template to the writer, interpolating the data.
func Generate(t *template.Template, wr io.Writer, name string, data any) error {
	err := t.ExecuteTemplate(wr, name, data)
	if err != nil {
		return fmt.Errorf("execute template %s: %w", name, err)
	}
	return nil
}

// GenerateFile generates a file from a template. It will create the directory if it does not exist,
// finally calling the Generate function to perform the interpolation.
func GenerateFile(t *template.Template, exampleFilePath string, name string, data any) error {
	err := os.MkdirAll(filepath.Dir(exampleFilePath), 0o755)
	if err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	exampleFile, err := os.Create(exampleFilePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer exampleFile.Close()

	return Generate(t, exampleFile, name, data)
}
