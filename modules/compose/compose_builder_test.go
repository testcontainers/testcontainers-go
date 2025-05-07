package compose

import (
	"fmt"
	"html/template"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testdataPackage = "testdata"
)

func RenderComposeProfiles(t *testing.T) string {
	t.Helper()

	return writeTemplate(t, "docker-compose-profiles.yml")
}

func RenderComposeComplex(t *testing.T) (string, []int) {
	t.Helper()

	ports := []int{getFreePort(t), getFreePort(t)}

	return writeTemplate(t, "docker-compose-complex.yml", ports...), ports
}

func RenderComposeComplexForLocal(t *testing.T) (string, []int) {
	t.Helper()

	ports := []int{getFreePort(t), getFreePort(t)}

	return writeTemplateWithSrvType(t, "docker-compose-complex.yml", "local", ports...), ports
}

func RenderComposeOverride(t *testing.T) string {
	t.Helper()

	return writeTemplate(t, "docker-compose-override.yml", getFreePort(t))
}

func RenderComposeOverrideForLocal(t *testing.T) string {
	t.Helper()

	return writeTemplateWithSrvType(t, "docker-compose-override.yml", "local", getFreePort(t))
}

func RenderComposePostgres(t *testing.T) string {
	t.Helper()

	return writeTemplate(t, "docker-compose-postgres.yml", getFreePort(t))
}

func RenderComposePostgresForLocal(t *testing.T) string {
	t.Helper()

	return writeTemplateWithSrvType(t, "docker-compose-postgres.yml", "local", getFreePort(t))
}

func RenderComposeSimple(t *testing.T) (string, []int) {
	t.Helper()

	ports := []int{getFreePort(t)}
	return writeTemplate(t, "docker-compose-simple.yml", ports...), ports
}

func RenderComposeSimpleForLocal(t *testing.T) (string, []int) {
	t.Helper()

	ports := []int{getFreePort(t)}
	return writeTemplateWithSrvType(t, "docker-compose-simple.yml", "local", ports...), ports
}

func RenderComposeWithBuild(t *testing.T) string {
	t.Helper()

	return writeTemplate(t, "docker-compose-build.yml", getFreePort(t))
}

func RenderComposeWithName(t *testing.T) string {
	t.Helper()

	return writeTemplate(t, "docker-compose-container-name.yml", getFreePort(t))
}

func RenderComposeWithNameForLocal(t *testing.T) string {
	t.Helper()

	return writeTemplateWithSrvType(t, "docker-compose-container-name.yml", "local", getFreePort(t))
}

func RenderComposeWithoutExposedPorts(t *testing.T) string {
	t.Helper()

	return writeTemplate(t, "docker-compose-no-exposed-ports.yml")
}

func RenderComposeWithoutExposedPortsForLocal(t *testing.T) string {
	t.Helper()

	return writeTemplateWithSrvType(t, "docker-compose-no-exposed-ports.yml", "local")
}

func RenderComposeWithVolume(t *testing.T) string {
	t.Helper()

	return writeTemplate(t, "docker-compose-volume.yml", getFreePort(t))
}

func RenderComposeWithVolumeForLocal(t *testing.T) string {
	t.Helper()

	return writeTemplateWithSrvType(t, "docker-compose-volume.yml", "local", getFreePort(t))
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort(t *testing.T) int {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoErrorf(t, err, "failed to resolve TCP address")

	l, err := net.ListenTCP("tcp", addr)
	require.NoErrorf(t, err, "failed to listen on TCP address")
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

func writeTemplate(t *testing.T, templateFile string, port ...int) string {
	t.Helper()
	return writeTemplateWithSrvType(t, templateFile, "api", port...)
}

func writeTemplateWithSrvType(t *testing.T, templateFile string, srvType string, port ...int) string {
	t.Helper()

	tmpDir := t.TempDir()
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")

	tmpl, err := template.ParseFiles(filepath.Join(testdataPackage, templateFile))
	require.NoErrorf(t, err, "parsing template file")

	values := map[string]any{}
	for i, p := range port {
		values[fmt.Sprintf("Port_%d", i)] = p
	}

	values["ServiceType"] = srvType

	output, err := os.Create(composeFile)
	require.NoErrorf(t, err, "creating output file")
	defer func() {
		require.NoError(t, output.Close())
	}()

	executeTemplateFile := func(templateFile *template.Template, wr io.Writer, data any) error {
		return templateFile.Execute(wr, data)
	}

	err = executeTemplateFile(tmpl, output, values)
	require.NoErrorf(t, err, "executing template file")

	return composeFile
}
