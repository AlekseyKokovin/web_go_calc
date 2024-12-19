package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	router := http.NewServeMux()
	router.HandleFunc("/api/v1/calculate", CalculateHandler)

	testCases := []struct {
		name       string
		input      CalculateRequest
		wantStatus int
		wantResult float64
		wantError  string
	}{
		{
			name: "valid expression",
			input: CalculateRequest{
				Expression: "2+3*4",
			},
			wantStatus: http.StatusOK,
			wantResult: 14,
		},
		{
			name: "valid expression",
			input: CalculateRequest{
				Expression: "2+3*(4+4)",
			},
			wantStatus: http.StatusOK,
			wantResult: 26,
		},
		{
			name: "valid expression",
			input: CalculateRequest{
				Expression: "(2+3)*4",
			},
			wantStatus: http.StatusOK,
			wantResult: 20,
		},
		{
			name: "valid expression",
			input: CalculateRequest{
				Expression: "2+(3*4)",
			},
			wantStatus: http.StatusOK,
			wantResult: 14,
		},
		{
			name: "invalid expression",
			input: CalculateRequest{
				Expression: "2+3*x",
			},
			wantStatus: 422,
			wantError:  "Expression is not valid",
		},
		{
			name: "internal server error",
			input: CalculateRequest{
				Expression: "2+3/0",
			},
			wantStatus: 500,
			wantError:  "Internal server error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tc.input)
			req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonData))
			rec := httptest.NewRecorder()

			CalculateHandler(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rec.Code)
			}

			if tc.wantResult != 0 {
				var resp CalculateResult
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if resp.Result != tc.wantResult {
					t.Errorf("want result %f, got %f", tc.wantResult, resp.Result)
				}
			}

			if tc.wantError != "" {
				var resp CalculateError
				err := json.Unmarshal(rec.Body.Bytes(), &resp)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if resp.Error != tc.wantError {
					t.Errorf("want result %s, got %s", tc.wantError, resp.Error)
				}
			}
		})
	}
}
