package config

import "github.com/caarlos0/env/v11"

const localEnv = "local"

type Config struct {
	Environment            string  `env:"ENVIRONMENT,required,notEmpty"`
	Port                   string  `env:"PORT,required,notEmpty"` // ローカルでのみ使用。Lambda では listenAndServe を通らない。
	DocsUser               string  `env:"DOCS_USER,required,notEmpty"`
	DocsPassword           string  `env:"DOCS_PASSWORD,required,notEmpty"`
	LambdaRuntimeAPI       string  `env:"AWS_LAMBDA_RUNTIME_API"` // AWS が実行環境に設定。ローカルでは空。
	SentryDSN              string  `env:"SENTRY_DSN"`             // 未設定なら Sentry 連携を無効化する。
	SentryTracesSampleRate float64 `env:"SENTRY_TRACES_SAMPLE_RATE" envDefault:"0.0"`
}

func (c Config) OnLambda() bool {
	return c.LambdaRuntimeAPI != ""
}

func (c Config) SentryEnabled() bool {
	return c.SentryDSN != ""
}

func (c Config) IsLocal() bool {
	return c.Environment == localEnv
}

func Load() (Config, error) {
	return env.ParseAs[Config]()
}
