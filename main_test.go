package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	city := "moscow"
	total := len(cafeList[city])

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(total, 100)},
	}

	for _, tc := range requests {
		t.Run(fmt.Sprintf("count=%d", tc.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest(
				"GET",
				fmt.Sprintf("/cafe?city=%s&count=%d", city, tc.count),
				nil,
			)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())

			if body == "" {
				assert.Equal(t, tc.want, 0)
				return
			}

			parts := strings.Split(body, ",")
			assert.Equal(t, tc.want, len(parts))
		})
	}
}
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, tc := range requests {
		t.Run("search="+tc.search, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest(
				"GET",
				fmt.Sprintf("/cafe?city=%s&search=%s", city, tc.search),
				nil,
			)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())

			if body == "" {
				assert.Equal(t, tc.wantCount, 0)
				return
			}

			parts := strings.Split(body, ",")
			assert.Equal(t, tc.wantCount, len(parts))

			for _, name := range parts {
				assert.True(
					t,
					strings.Contains(strings.ToLower(name), strings.ToLower(tc.search)),
					"кафе '%s' не содержит '%s'",
					name, tc.search,
				)
			}
		})
	}
}
