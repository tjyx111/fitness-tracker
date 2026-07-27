package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReportUploadListAndRead(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer server.csv.Close()
	server.reportUploadToken = "test-upload-token"

	html := `<!doctype html><html><head><title>分析</title></head><body>进展良好</body></html>`
	upload := httptest.NewRequest(http.MethodPut, "/api/reports/weekly-report.htm", strings.NewReader(html))
	upload.Header.Set("Authorization", "Bearer test-upload-token")
	uploadRecorder := httptest.NewRecorder()
	server.handleReport(uploadRecorder, upload)
	if uploadRecorder.Code != http.StatusCreated {
		t.Fatalf("upload code=%d body=%s", uploadRecorder.Code, uploadRecorder.Body.String())
	}

	stored, err := os.ReadFile(filepath.Join(server.reportDir, "weekly-report.htm"))
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != html {
		t.Fatalf("stored report differs: %q", stored)
	}

	listRecorder := httptest.NewRecorder()
	server.handleReports(listRecorder, httptest.NewRequest(http.MethodGet, "/api/reports", nil))
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list code=%d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	var reports []reportInfo
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &reports); err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || reports[0].Name != "weekly-report.htm" {
		t.Fatalf("unexpected reports: %#v", reports)
	}
	if reports[0].URL != "/api/reports/weekly-report.htm" || reports[0].Size != int64(len(html)) {
		t.Fatalf("unexpected report metadata: %#v", reports[0])
	}

	readRecorder := httptest.NewRecorder()
	server.handleReport(readRecorder, httptest.NewRequest(http.MethodGet, reports[0].URL, nil))
	if readRecorder.Code != http.StatusOK {
		t.Fatalf("read code=%d body=%s", readRecorder.Code, readRecorder.Body.String())
	}
	if readRecorder.Body.String() != html {
		t.Fatalf("read report differs: %q", readRecorder.Body.String())
	}
	if got := readRecorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type=%q", got)
	}
	if got := readRecorder.Header().Get("Content-Security-Policy"); !strings.Contains(got, "sandbox allow-scripts") {
		t.Fatalf("Content-Security-Policy=%q", got)
	}
}

func TestReportUploadRequiresConfiguredBearerToken(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer server.csv.Close()

	request := httptest.NewRequest(http.MethodPut, "/api/reports/report.htm", strings.NewReader("<html></html>"))
	recorder := httptest.NewRecorder()
	server.handleReport(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured upload code=%d body=%s", recorder.Code, recorder.Body.String())
	}

	server.reportUploadToken = "correct-token"
	request = httptest.NewRequest(http.MethodPut, "/api/reports/report.htm", strings.NewReader("<html></html>"))
	request.Header.Set("Authorization", "Bearer wrong-token")
	recorder = httptest.NewRecorder()
	server.handleReport(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized upload code=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestReportUploadReplacesExistingReport(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer server.csv.Close()
	server.reportUploadToken = "token"

	upload := func(content string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPut, "/api/reports/report.html", strings.NewReader(content))
		request.Header.Set("Authorization", "Bearer token")
		recorder := httptest.NewRecorder()
		server.handleReport(recorder, request)
		return recorder
	}

	if recorder := upload("<html>first</html>"); recorder.Code != http.StatusCreated {
		t.Fatalf("first upload code=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder := upload("<html>second</html>"); recorder.Code != http.StatusOK {
		t.Fatalf("replacement upload code=%d body=%s", recorder.Code, recorder.Body.String())
	}
	content, err := os.ReadFile(filepath.Join(server.reportDir, "report.html"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "<html>second</html>" {
		t.Fatalf("replacement content=%q", content)
	}
}

func TestReportHandlerRejectsInvalidNameAndOversizedBody(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer server.csv.Close()
	server.reportUploadToken = "token"

	invalid := httptest.NewRecorder()
	server.handleReport(invalid, httptest.NewRequest(http.MethodGet, "/api/reports/not-html.txt", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid name code=%d body=%s", invalid.Code, invalid.Body.String())
	}

	request := httptest.NewRequest(
		http.MethodPut,
		"/api/reports/large.htm",
		bytes.NewReader(make([]byte, maxReportSize+1)),
	)
	request.Header.Set("Authorization", "Bearer token")
	oversized := httptest.NewRecorder()
	server.handleReport(oversized, request)
	if oversized.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized upload code=%d body=%s", oversized.Code, oversized.Body.String())
	}
}
