package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPKDownload(t *testing.T) {
	apkPath := filepath.Join(t.TempDir(), apkDownloadName)
	content := []byte("test APK content")
	if err := os.WriteFile(apkPath, content, 0o600); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/downloads/assistant.apk", "/downloads/fitness-tracker.apk"} {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, path, nil)
			apkDownloadHandler(apkPath).ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/vnd.android.package-archive" {
				t.Fatalf("Content-Type = %q", got)
			}
			if got := recorder.Header().Get("Content-Disposition"); !strings.Contains(got, apkDownloadName) {
				t.Fatalf("Content-Disposition = %q", got)
			}
			if got := recorder.Body.Bytes(); string(got) != string(content) {
				t.Fatalf("body = %q, want %q", got, content)
			}
		})
	}
}

func TestAPKDownloadMissing(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/downloads/assistant.apk", nil)
	apkDownloadHandler("").ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestAPKDownloadRejectsUnsupportedMethod(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/downloads/assistant.apk", nil)
	apkDownloadHandler("unused").ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
