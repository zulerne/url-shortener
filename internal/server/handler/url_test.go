package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zulerne/url-shortener/internal/lib/logger"
	"github.com/zulerne/url-shortener/internal/server/handler"
	"github.com/zulerne/url-shortener/internal/storage"
)

type reqBody struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

func TestCreateURLHandler(t *testing.T) {
	slog.SetDefault(logger.NewDiscardLogger())

	cases := []struct {
		name      string
		input     reqBody
		code      int
		respError string
		mockSetup func(s *MockStorage)
	}{
		{
			name: "Success",
			input: reqBody{
				URL:   "https://google.com",
				Alias: "test_alias",
			},
			code: http.StatusOK,
			mockSetup: func(s *MockStorage) {
				s.EXPECT().
					SaveURL("https://google.com", "test_alias").
					Return(1, nil).
					Once()
			},
		},
		{
			name: "Empty alias (Auto-generated)",
			input: reqBody{
				URL: "https://google.com",
			},
			code: http.StatusOK,
			mockSetup: func(s *MockStorage) {
				s.EXPECT().
					SaveURL("https://google.com", mock.AnythingOfType("string")).
					Return(1, nil).
					Once()
			},
		},
		{
			name: "Empty URL",
			input: reqBody{
				Alias: "some_alias",
			},
			code:      http.StatusBadRequest,
			respError: "'URL' is required",
			mockSetup: func(s *MockStorage) {
			},
		},
		{
			name: "Invalid URL",
			input: reqBody{
				URL:   "not-a-valid-url",
				Alias: "some_alias",
			},
			code:      http.StatusBadRequest,
			respError: "'URL' is not a valid url",
			mockSetup: nil,
		},
		{
			name: "SaveURL Internal Error",
			input: reqBody{
				URL:   "https://google.com",
				Alias: "fail",
			},
			code:      http.StatusInternalServerError,
			respError: "failed to save url",
			mockSetup: func(s *MockStorage) {
				s.EXPECT().
					SaveURL("https://google.com", "fail").
					Return(0, errors.New("unexpected db error")).
					Once()
			},
		},
		{
			name: "Alias Already Exists",
			input: reqBody{
				URL:   "https://google.com",
				Alias: "exists",
			},
			code:      http.StatusConflict,
			respError: storage.ErrAliasExists.Error(),
			mockSetup: func(s *MockStorage) {
				s.EXPECT().
					SaveURL("https://google.com", "exists").
					Return(0, storage.ErrAliasExists).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storageMock := NewMockStorage(t)
			if tc.mockSetup != nil {
				tc.mockSetup(storageMock)
			}

			h := handler.NewHandler(storageMock, 6, "", "")

			body, _ := json.Marshal(tc.input)
			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth("", "")
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			require.Equal(t, tc.code, w.Code)

			if tc.respError != "" {
				var resp handler.CreateURLResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Equal(t, tc.respError, resp.Error)
			}
		})
	}
}

func TestCreateURLHandlerAuth(t *testing.T) {
	slog.SetDefault(logger.NewDiscardLogger())

	user := "user"
	pass := "pass"

	cases := []struct {
		name      string
		code      int
		user      string
		pass      string
		mockSetup func(s *MockStorage)
	}{
		{
			name: "Success",
			code: http.StatusOK,
			user: user,
			pass: pass,
			mockSetup: func(s *MockStorage) {
				s.EXPECT().
					SaveURL("https://google.com", "test_alias").
					Return(1, nil).
					Once()
			},
		},
		{
			name: "Unauthorized",
			code: http.StatusUnauthorized,
			user: user,
			pass: "wrong_pass",
		},
		{
			name: "Empty",
			code: http.StatusUnauthorized,
			user: "",
			pass: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storageMock := NewMockStorage(t)
			if tc.mockSetup != nil {
				tc.mockSetup(storageMock)
			}

			h := handler.NewHandler(storageMock, 6, user, pass)

			body, _ := json.Marshal(reqBody{
				URL:   "https://google.com",
				Alias: "test_alias",
			})
			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth(tc.user, tc.pass)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			require.Equal(t, tc.code, w.Code)
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	slog.SetDefault(logger.NewDiscardLogger())

	cases := []struct {
		name      string
		code      int
		alias     string
		mockSetup func(s *MockStorage)
	}{
		{
			name:  "Success",
			code:  http.StatusTemporaryRedirect,
			alias: "test_alias",
			mockSetup: func(s *MockStorage) {
				s.EXPECT().
					GetURL("test_alias").
					Return("https://google.com", nil).
					Once()
			},
		},
		{
			name:  "NotFound",
			code:  http.StatusNotFound,
			alias: "not_found",
			mockSetup: func(s *MockStorage) {
				s.EXPECT().
					GetURL("not_found").
					Return("", storage.ErrNotFound).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storageMock := NewMockStorage(t)
			if tc.mockSetup != nil {
				tc.mockSetup(storageMock)
			}

			h := handler.NewHandler(storageMock, 6, "", "")

			req := httptest.NewRequest(http.MethodGet, "/"+tc.alias, nil)
			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth("", "")
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			require.Equal(t, tc.code, w.Code)
		})
	}
}
