package context_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

func TestModule_Validate(outer *testing.T) {
	outer.Parallel()

	tests := []struct {
		name        string
		module      context.TestcontainersModule
		expectedErr error
	}{
		{
			name: "only alphabetical characters in name/title",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "alphanumerical characters in name",
			module: context.TestcontainersModule{
				Name:      "AmazingDB4tw",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "alphanumerical characters in title",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "AmazingDB4tw",
			},
		},
		{
			name: "non-alphanumerical characters in name",
			module: context.TestcontainersModule{
				Name:      "Amazing DB 4 The Win",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "non-alphanumerical characters in title",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "Amazing DB 4 The Win",
			},
			expectedErr: errors.New("invalid title: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "leading numerical character in name",
			module: context.TestcontainersModule{
				Name:      "1AmazingDB",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: 1AmazingDB. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "leading numerical character in title",
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
