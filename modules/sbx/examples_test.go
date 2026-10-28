package sbx_test

import (
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/sbx"
)

// These examples omit Output directives so they compile without changing the
// process-wide provider configuration used by the integration tests.
func ExampleConfigure() {
	// configureProvider {
	if err := sbx.Configure(
		sbx.WithStartupTimeout(8*time.Minute),
		sbx.WithKits("./testdata/test-tools"),
		sbx.WithKitArgs(map[string]string{
			"test-tools.mode": "integration",
		}),
	); err != nil {
		log.Fatal(err)
	}
	// }
}

func ExampleWithBinary() {
	// withBinary {
	if err := sbx.Configure(sbx.WithBinary("/usr/local/bin/sbx")); err != nil {
		log.Fatal(err)
	}
	// }
}

func ExampleWithWorkspace() {
	// withWorkspace {
	if err := sbx.Configure(sbx.WithWorkspace(".")); err != nil {
		log.Fatal(err)
	}
	// }
}

func ExampleWithBindRoots() {
	// withBindRoots {
	if err := sbx.Configure(sbx.WithBindRoots("/opt/test-fixtures", "/srv/shared-fixtures")); err != nil {
		log.Fatal(err)
	}
	// }
}

func ExampleWithStartupTimeout() {
	// withStartupTimeout {
	if err := sbx.Configure(sbx.WithStartupTimeout(8 * time.Minute)); err != nil {
		log.Fatal(err)
	}
	// }
}

func ExampleWithAppName() {
	// withAppName {
	if err := sbx.Configure(sbx.WithAppName("tc-tests")); err != nil {
		log.Fatal(err)
	}
	// }
}

func ExampleWithKits() {
	// withKits {
	if err := sbx.Configure(sbx.WithKits("./testdata/test-tools", "docker.io/my-org/another-kit:1.0.0")); err != nil {
		log.Fatal(err)
	}
	// }
}

func ExampleWithKitArgs() {
	// withKitArgs {
	if err := sbx.Configure(sbx.WithKitArgs(map[string]string{
		"test-tools.mode": "integration",
	})); err != nil {
		log.Fatal(err)
	}
	// }
}
