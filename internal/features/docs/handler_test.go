package docs_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
	"github.com/inouey1008/kusuri-api-poc/internal/features/docs"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
)

func newConfig(environment string) config.Config {
	return config.Config{
		Environment:  environment,
		DocsUser:     "docs",
		DocsPassword: "secret",
	}
}

func send(t *testing.T, cfg config.Config, path string, credentials func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()

	e := server.NewWith([]server.Registerer{docs.NewHandler(cfg)}, nil)

	request := httptest.NewRequest(http.MethodGet, path, nil)
	if credentials != nil {
		credentials(request)
	}

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	return recorder
}

func TestHandler_Spec(t *testing.T) {
	t.Parallel()

	t.Run(`openapi.yaml の中身を返す`, func(t *testing.T) {
		t.Parallel()

		recorder := send(t, newConfig("local"), "/openapi.yaml", nil)

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "application/yaml")
		assert.Contains(t, recorder.Body.String(), "openapi:")
	})
}

func TestHandler_Page(t *testing.T) {
	t.Parallel()

	t.Run(`Redoc を表示する HTML を返す`, func(t *testing.T) {
		t.Parallel()

		recorder := send(t, newConfig("local"), "/docs", nil)

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, recorder.Body.String(), `spec-url="/openapi.yaml"`)
	})
}

func TestHandler_Script(t *testing.T) {
	t.Parallel()

	t.Run(`Redoc の JavaScript を返す`, func(t *testing.T) {
		t.Parallel()

		recorder := send(t, newConfig("local"), "/docs/redoc.standalone.js", nil)

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "application/javascript")
		assert.NotEmpty(t, recorder.Body.Bytes(), "空なら埋め込みに失敗している")
	})
}

func TestHandler_BasicAuth(t *testing.T) {
	t.Parallel()

	paths := []string{"/openapi.yaml", "/docs", "/docs/redoc.standalone.js"}

	t.Run(`local 以外では資格情報が無ければ 401 を返す`, func(t *testing.T) {
		t.Parallel()

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				t.Parallel()

				recorder := send(t, newConfig("dev"), path, nil)

				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			})
		}
	})

	t.Run(`正しい認証情報の場合、エラーにならない`, func(t *testing.T) {
		t.Parallel()

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				t.Parallel()

				recorder := send(t, newConfig("dev"), path, func(r *http.Request) {
					r.SetBasicAuth("docs", "secret")
				})

				assert.Equal(t, http.StatusOK, recorder.Code)
			})
		}
	})

	t.Run(`誤った認証情報の場合は、401 エラーになる`, func(t *testing.T) {
		t.Parallel()

		recorder := send(t, newConfig("dev"), "/docs", func(r *http.Request) {
			r.SetBasicAuth("docs", "wrong")
		})

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run(`local なら Basic 認証を掛けない`, func(t *testing.T) {
		t.Parallel()

		recorder := send(t, newConfig("local"), "/docs", nil)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}
