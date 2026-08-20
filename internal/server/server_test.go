package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
)

type stubRegisterer struct {
	path string
	body string
}

func (s *stubRegisterer) Register(e *echo.Echo) {
	e.GET(s.path, func(c echo.Context) error {
		return c.String(http.StatusOK, s.body)
	})
}

func recordingMiddleware(name string, order *[]string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			*order = append(*order, name)
			return next(c)
		}
	}
}

func sendGet(t *testing.T, e *echo.Echo, path string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	return recorder
}

func TestNewWith_Routing(t *testing.T) {
	e := server.NewWith([]server.Registerer{
		&stubRegisterer{path: "/drugs", body: "drugs"},
		&stubRegisterer{path: "/facilities", body: "facilities"},
	}, nil)

	t.Run(`1 つ目の Registerer のルート`, func(t *testing.T) {
		recorder := sendGet(t, e, "/drugs")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "drugs", recorder.Body.String())
	})

	t.Run(`2 つ目の Registerer のルート`, func(t *testing.T) {
		recorder := sendGet(t, e, "/facilities")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "facilities", recorder.Body.String())
	})

	t.Run(`未登録のパスは 404 を JSON で返す`, func(t *testing.T) {
		recorder := sendGet(t, e, "/unknown")

		require.Equal(t, http.StatusNotFound, recorder.Code)

		var body errorx.Errorx
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		assert.Equal(t, "NOT_FOUND", body.Code, "Echo の 404 もアプリ定義の形式へ揃える")
	})
}

func TestNewWith_MiddlewareOrder(t *testing.T) {
	order := []string{}
	e := server.NewWith(
		[]server.Registerer{&stubRegisterer{path: "/drugs", body: "drugs"}},
		[]echo.MiddlewareFunc{
			recordingMiddleware("first", &order),
			recordingMiddleware("second", &order),
		},
	)

	sendGet(t, e, "/drugs")

	assert.Equal(t, []string{"first", "second"}, order, "先に定義した middleware から実行される")
}
