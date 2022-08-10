package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoad(t *testing.T) {
	testCases := map[string]struct {
		ConfigFile string
		Expected   Config
	}{
		"simple example": {
			ConfigFile: "./fixtures/example.yaml",
			Expected: Config{
				GPXSource: GPXSource{
					URLTemplate: "https://example.com/gpx?from={{ .From }}&to={{ .To }}",
					Username:    "example",
					Password:    "password",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			cfg, err := Load(testCase.ConfigFile)
			require.NoError(t, err)

			assert.Equal(t, testCase.Expected, cfg)
		})
	}
}
