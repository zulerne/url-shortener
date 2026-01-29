package tests

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"github.com/zulerne/url-shortener/internal/lib/random"
)

const (
	host = "localhost:8081"
)

func TestMain(m *testing.M) {
	os.Setenv("ENV", "local")
	os.Setenv("STORAGE_PATH", "./storage_test.db")
	os.Setenv("HTTP_ADDRESS", host)
	os.Setenv("HTTP_TIMEOUT", "5s")
	os.Setenv("HTTP_IDLE_TIMEOUT", "60s")
	os.Setenv("HTTP_SHUTDOWN_TIMEOUT", "5s")
	os.Setenv("HTTP_USER", "admin")
	os.Setenv("HTTP_PASSWORD", "admin")

	cmd := exec.Command("go", "run", "../cmd/url-shortener/main.go")
	cmd.Env = os.Environ()
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start app: %v\n", err)
		os.Exit(1)
	}

	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get("http://" + host + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !ready {
		fmt.Println("Server not ready, timing out")
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		os.Exit(1)
	}

	code := m.Run()

	syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	cmd.Wait()
	os.Remove("./storage_test.db")

	os.Exit(code)
}

func TestCreateUrlWithAlias(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(map[string]string{
			"url":   gofakeit.URL(),
			"alias": random.Alias(10),
		}).
		WithBasicAuth("admin", "admin").
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ContainsKey("alias")
}

func TestCreateUrlWithoutAlias(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(map[string]string{
			"url": gofakeit.URL(),
		}).
		WithBasicAuth("admin", "admin").
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ContainsKey("alias")
}

func TestCreateAndRedirectURL(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			// Save
			req := e.POST("/url").
				WithJSON(map[string]string{
					"url":   tc.url,
					"alias": tc.alias,
				}).
				WithBasicAuth("admin", "admin")

			if tc.error != "" {
				req.Expect().
					Status(http.StatusBadRequest).
					JSON().Object().
					ContainsKey("error").
					Value("error").String().Contains(tc.error)
				return
			}

			resp := req.Expect().
				Status(http.StatusOK).
				JSON().Object()

			alias := tc.alias
			if alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}

			// Redirect
			testRedirect(t, alias, tc.url)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	require.Equal(t, urlToRedirect, resp.Header.Get("Location"))
}
