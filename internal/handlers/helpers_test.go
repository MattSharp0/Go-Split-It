package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a new HTTP request with path parameters
func createRequest(method, url string, body []byte) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	return req
}

// Helper function to create a request with path values
func createRequestWithPath(method, url, pathParamName, pathParamValue string, body []byte) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.SetPathValue(pathParamName, pathParamValue)
	return req
}

// Helper function to validate JSON response
func validateJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	t.Helper()
	assert.Equal(t, expectedStatus, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	if expectedBody != nil {
		var actualBody interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &actualBody)
		require.NoError(t, err, "Response body should be valid JSON")

		expectedJSON, err := json.Marshal(expectedBody)
		require.NoError(t, err)

		var expectedBodyUnmarshaled interface{}
		err = json.Unmarshal(expectedJSON, &expectedBodyUnmarshaled)
		require.NoError(t, err)

		assert.Equal(t, expectedBodyUnmarshaled, actualBody)
	}
}

// TestParsePathInt64 tests the ParsePathInt64 function
func TestParsePathInt64(t *testing.T) {
	tests := []struct {
		name          string
		pathValue     string
		expectedValue int64
		expectedOk    bool
		expectedCode  int
	}{
		{
			name:          "valid int64",
			pathValue:     "123",
			expectedValue: 123,
			expectedOk:    true,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "zero value",
			pathValue:     "0",
			expectedValue: 0,
			expectedOk:    true,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "large int64",
			pathValue:     "9223372036854775807",
			expectedValue: 9223372036854775807,
			expectedOk:    true,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "empty string",
			pathValue:     "",
			expectedValue: 0,
			expectedOk:    false,
			expectedCode:  http.StatusBadRequest,
		},
		{
			name:          "invalid format",
			pathValue:     "abc",
			expectedValue: 0,
			expectedOk:    false,
			expectedCode:  http.StatusBadRequest,
		},
		{
			name:          "float string",
			pathValue:     "123.45",
			expectedValue: 0,
			expectedOk:    false,
			expectedCode:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := createRequestWithPath("GET", "/test/{id}", "id", tt.pathValue, nil)

			value, ok := ParsePathInt64(rr, req, "id", "User ID is required")

			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectedOk, ok)
			if !tt.expectedOk {
				assert.Equal(t, tt.expectedCode, rr.Code)
			}
		})
	}
}

// TestParseQueryInt32 tests the ParseQueryInt32 function
func TestParseQueryInt32(t *testing.T) {
	tests := []struct {
		name         string
		queryParam   string
		defaultValue int32
		expected     int32
		expectError  bool
	}{
		{
			name:         "valid int32",
			queryParam:   "123",
			defaultValue: 0,
			expected:     123,
			expectError:  false,
		},
		{
			name:         "missing parameter",
			queryParam:   "",
			defaultValue: 100,
			expected:     100,
			expectError:  false,
		},
		{
			name:         "zero value",
			queryParam:   "0",
			defaultValue: 100,
			expected:     0,
			expectError:  false,
		},
		{
			name:         "invalid format",
			queryParam:   "abc",
			defaultValue: 100,
			expected:     100,
			expectError:  true,
		},
		{
			name:         "out of range (too large)",
			queryParam:   "2147483648",
			defaultValue: 100,
			expected:     100,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/test"
			if tt.queryParam != "" {
				url += "?param=" + tt.queryParam
			}
			req := createRequest("GET", url, nil)

			result, err := ParseQueryInt32(req, "param", tt.defaultValue)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.defaultValue, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestParseLimitOffset tests the ParseLimitOffset function
func TestParseLimitOffset(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedLimit int32
		expectedOff   int32
		expectError   bool
	}{
		{
			name:          "both parameters provided",
			queryParams:   map[string]string{"limit": "50", "offset": "10"},
			expectedLimit: 50,
			expectedOff:   10,
			expectError:   false,
		},
		{
			name:          "default values",
			queryParams:   map[string]string{},
			expectedLimit: 100,
			expectedOff:   0,
			expectError:   false,
		},
		{
			name:          "only limit provided",
			queryParams:   map[string]string{"limit": "25"},
			expectedLimit: 25,
			expectedOff:   0,
			expectError:   false,
		},
		{
			name:          "only offset provided",
			queryParams:   map[string]string{"offset": "20"},
			expectedLimit: 100,
			expectedOff:   20,
			expectError:   false,
		},
		{
			name:          "invalid limit",
			queryParams:   map[string]string{"limit": "abc", "offset": "10"},
			expectedLimit: 100,
			expectedOff:   0,
			expectError:   true,
		},
		{
			name:          "invalid offset",
			queryParams:   map[string]string{"limit": "50", "offset": "xyz"},
			expectedLimit: 50,
			expectedOff:   0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/test?"
			first := true
			for key, value := range tt.queryParams {
				if !first {
					url += "&"
				}
				url += key + "=" + value
				first = false
			}
			if len(tt.queryParams) == 0 {
				url = "/test"
			}
			req := createRequest("GET", url, nil)

			limit, offset, err := ParseLimitOffset(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLimit, limit)
				assert.Equal(t, tt.expectedOff, offset)
			}
		})
	}
}

// TestDecodeJSONBody tests the DecodeJSONBody function
func TestDecodeJSONBody(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		target      interface{}
		expectError bool
	}{
		{
			name:        "valid JSON object",
			body:        `{"name":"John"}`,
			target:      &map[string]string{},
			expectError: false,
		},
		{
			name:        "valid JSON array",
			body:        `[1,2,3]`,
			target:      &[]int{},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			body:        `{"name":}`,
			target:      &map[string]string{},
			expectError: true,
		},
		{
			name:        "empty body",
			body:        ``,
			target:      &map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createRequest("POST", "/test", []byte(tt.body))

			err := DecodeJSONBody(req, tt.target)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestWriteJSONResponse tests the WriteJSONResponse function
func TestWriteJSONResponse(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		data        interface{}
		expectError bool
	}{
		{
			name:        "success with object",
			statusCode:  http.StatusOK,
			data:        map[string]string{"message": "success"},
			expectError: false,
		},
		{
			name:        "created status",
			statusCode:  http.StatusCreated,
			data:        map[string]int64{"id": 123},
			expectError: false,
		},
		{
			name:        "empty object",
			statusCode:  http.StatusOK,
			data:        map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			err := WriteJSONResponse(rr, tt.statusCode, tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.statusCode, rr.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				var response interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandleDBError tests the HandleDBError function
func TestHandleDBError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedMessage string
		shouldHandle    bool
	}{
		{
			name:            "no error",
			err:             nil,
			expectedStatus:  0,
			expectedMessage: "",
			shouldHandle:    false,
		},
		{
			name:            "not found error (pgx.ErrNoRows)",
			err:             pgx.ErrNoRows,
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "not found",
			shouldHandle:    true,
		},
		{
			name:            "generic error",
			err:             errors.New("database connection failed"),
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "An error has occurred",
			shouldHandle:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			handled := HandleDBError(rr, tt.err, "not found", "An error has occurred", "test error")

			assert.Equal(t, tt.shouldHandle, handled)
			if tt.shouldHandle {
				assert.Equal(t, tt.expectedStatus, rr.Code)
				if tt.expectedMessage != "" {
					assert.Contains(t, rr.Body.String(), tt.expectedMessage)
				}
			}
		})
	}
}

// TestHandleDBListError tests the HandleDBListError function
func TestHandleDBListError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedMessage string
		shouldHandle    bool
	}{
		{
			name:            "no error",
			err:             nil,
			expectedStatus:  0,
			expectedMessage: "",
			shouldHandle:    false,
		},
		{
			name:            "database error",
			err:             assert.AnError,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "An error has occurred",
			shouldHandle:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			handled := HandleDBListError(rr, tt.err, "An error has occurred", "test error")

			assert.Equal(t, tt.shouldHandle, handled)
			if tt.shouldHandle {
				assert.Equal(t, tt.expectedStatus, rr.Code)
				if tt.expectedMessage != "" {
					assert.Contains(t, rr.Body.String(), tt.expectedMessage)
				}
			}
		})
	}
}
