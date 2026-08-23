package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
)

var required = map[string]string{
	"ENVIRONMENT":   "local",
	"PORT":          "8080",
	"DOCS_USER":     "docs",
	"DOCS_PASSWORD": "secret",
}

func setRequired(t *testing.T) {
	t.Helper()

	for key, value := range required {
		t.Setenv(key, value)
	}
}

func TestLoad(t *testing.T) {
	t.Run(`設定した値を読み込むことができる`, func(t *testing.T) {
		setRequired(t)
		t.Setenv("PORT", "9000")

		cfg, err := config.Load()

		require.NoError(t, err)
		assert.Equal(t, "9000", cfg.Port)
		assert.Equal(t, "docs", cfg.DocsUser)
	})

	t.Run(`必須の環境変数が未設定の場合、エラーになる`, func(t *testing.T) {
		for key := range required {
			t.Run(key, func(t *testing.T) {
				setRequired(t)
				t.Setenv(key, "")

				_, err := config.Load()

				require.Error(t, err)
			})
		}
	})
}

func TestConfig_IsLocal(t *testing.T) {
	t.Run(`ENVIRONMENT が local のときだけ true`, func(t *testing.T) {
		testCases := map[string]bool{
			"local": true,
			"dev":   false,
			"prod":  false,
		}

		for environment, expected := range testCases {
			t.Run(environment, func(t *testing.T) {
				cfg := config.Config{Environment: environment}

				assert.Equal(t, expected, cfg.IsLocal())
			})
		}
	})
}

func TestConfig_OnLambda(t *testing.T) {
	t.Run(`AWS_LAMBDA_RUNTIME_API が設定されていれば true`, func(t *testing.T) {
		cfg := config.Config{LambdaRuntimeAPI: "127.0.0.1:9001"}

		assert.True(t, cfg.OnLambda())
	})

	t.Run(`未設定なら false`, func(t *testing.T) {
		cfg := config.Config{}

		assert.False(t, cfg.OnLambda())
	})
}
