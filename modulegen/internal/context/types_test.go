package context_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

func TestModule(t *testing.T) {
	tests := []struct {
		name               string
		module             context.TestcontainersModule
		expectedEntrypoint string
		expectedTitle      string
	}{
		{
			name: "module/title",
			module: context.TestcontainersModule{
				Name:      "mongoDB",
				IsModule:  true,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedEntrypoint: "Run",
			expectedTitle:      "MongoDB",
		},
		{
			name: "module/no-title",
			module: context.TestcontainersModule{
				Name:     "mongoDB",
				IsModule: true,
				Image:    "mongodb:latest",
			},
			expectedEntrypoint: "Run",
			expectedTitle:      "Mongodb",
		},
		{
			name: "example/title",
			module: context.TestcontainersModule{
				Name:      "mongoDB",
				IsModule:  false,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedEntrypoint: "run",
			expectedTitle:      "MongoDB",
		},
		{
			name: "example/no-title",
			module: context.TestcontainersModule{
				Name:     "mongoDB",
				IsModule: false,
				Image:    "mongodb:latest",
			},

			expectedEntrypoint: "run",
			expectedTitle:      "Mongodb",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module := test.module

			require.Equal(t, "mongodb", module.Lower())
			require.Equal(t, test.expectedTitle, module.Title())
			require.Equal(t, "Container", module.ContainerName())
			require.Equal(t, test.expectedEntrypoint, module.Entrypoint())
		})
	}
}

func TestModule_Validate(outer *testing.T) {
	outer.Parallel()

	tests := []struct {
		name        string
		module      context.TestcontainersModule
		expectedErr error
	}{
		{
			name: "name-title/alphanumerical",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "name/alphanumerical",
			module: context.TestcontainersModule{
				Name:      "AmazingDB4tw",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "title/alphanumerical",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "AmazingDB4tw",
			},
		},
		{
			name: "name/non-alphanumerical",
			module: context.TestcontainersModule{
				Name:      "Amazing DB 4 The Win",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "title/non-alphanumerical",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "Amazing DB 4 The Win",
			},
			expectedErr: errors.New("invalid title: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "name/leading-number",
			module: context.TestcontainersModule{
				Name:      "1AmazingDB",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: 1AmazingDB. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "title/leading-number",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "1AmazingDB",
			},
			expectedErr: errors.New("invalid title: 1AmazingDB. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
	}

	for _, test := range tests {
		outer.Run(test.name, func(t *testing.T) {
			if test.expectedErr != nil {
				require.EqualError(t, test.module.Validate(), test.expectedErr.Error())
			} else {
				require.NoError(t, test.module.Validate())
			}
		})
	}
}
