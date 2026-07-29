package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const maxReportSize = 10 << 20

type reportInfo struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	UpdatedAt   time.Time `json:"updatedAt"`
	URL         string    `json:"url"`
	DownloadURL string    `json:"downloadUrl"`
}

func (s *Server) configureReports(reportDir, uploadToken string) error {
	if reportDir == "" {
		return errors.New("report directory is empty")
	}
	if err := os.MkdirAll(reportDir, 0o750); err != nil {
		return fmt.Errorf("create report directory: %w", err)
	}
	s.reportDir = reportDir
	s.reportUploadToken = uploadToken
	return nil
}

func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/reports" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(s.reportDir)
	if err != nil {
		http.Error(w, "list reports failed", http.StatusInternalServerError)
		return
	}

	reports := make([]reportInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !isReportFilename(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		reportURL := "/api/reports/" + url.PathEscape(entry.Name())
		reports = append(reports, reportInfo{
			Name:        entry.Name(),
			Size:        info.Size(),
			UpdatedAt:   info.ModTime(),
			URL:         reportURL,
			DownloadURL: reportURL + "?download=1",
		})
	}
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].UpdatedAt.Equal(reports[j].UpdatedAt) {
			return reports[i].Name < reports[j].Name
		}
		return reports[i].UpdatedAt.After(reports[j].UpdatedAt)
	})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	if err := json.NewEncoder(w).Encode(reports); err != nil {
		http.Error(w, "encode reports failed", http.StatusInternalServerError)
	}
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	rawName := strings.TrimPrefix(r.URL.Path, "/api/reports/")
	name, err := url.PathUnescape(rawName)
	if err != nil || !isReportFilename(name) {
		http.Error(w, "invalid report filename", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		s.serveReport(w, r, name)
	case http.MethodPut:
		s.uploadReport(w, r, name)
	default:
		w.Header().Set("Allow", "GET, HEAD, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) serveReport(w http.ResponseWriter, r *http.Request, name string) {
	path := filepath.Join(s.reportDir, name)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "open report failed", http.StatusInternalServerError)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "open report failed", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	disposition := "inline"
	if r.URL.Query().Get("download") == "1" {
		disposition = "attachment"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": name}))
	w.Header().Set("Content-Security-Policy", "sandbox allow-scripts; default-src 'self' data: blob: https:; style-src 'self' 'unsafe-inline' https:; script-src 'self' 'unsafe-inline' https:; img-src 'self' data: blob: https:; font-src 'self' data: https:; connect-src 'none'; form-action 'none'; base-uri 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.ServeContent(w, r, name, info.ModTime(), f)
}

func (s *Server) uploadReport(w http.ResponseWriter, r *http.Request, name string) {
	if s.reportUploadToken == "" {
		http.Error(w, "report upload is not configured", http.StatusServiceUnavailable)
		return
	}
	if !validBearerToken(r.Header.Get("Authorization"), s.reportUploadToken) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReportSize+1)
	content, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "report exceeds 10 MiB limit", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "read report failed", http.StatusBadRequest)
		return
	}
	if len(content) == 0 {
		http.Error(w, "report is empty", http.StatusBadRequest)
		return
	}
	if len(content) > maxReportSize {
		http.Error(w, "report exceeds 10 MiB limit", http.StatusRequestEntityTooLarge)
		return
	}

	target := filepath.Join(s.reportDir, name)
	_, statErr := os.Stat(target)
	created := errors.Is(statErr, os.ErrNotExist)
	if statErr != nil && !created {
		http.Error(w, "inspect report failed", http.StatusInternalServerError)
		return
	}

	temp, err := os.CreateTemp(s.reportDir, ".report-upload-*")
	if err != nil {
		http.Error(w, "create report failed", http.StatusInternalServerError)
		return
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if err := temp.Chmod(0o640); err != nil {
		temp.Close()
		http.Error(w, "prepare report failed", http.StatusInternalServerError)
		return
	}
	if _, err := temp.Write(content); err != nil {
		temp.Close()
		http.Error(w, "write report failed", http.StatusInternalServerError)
		return
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		http.Error(w, "sync report failed", http.StatusInternalServerError)
		return
	}
	if err := temp.Close(); err != nil {
		http.Error(w, "close report failed", http.StatusInternalServerError)
		return
	}
	if err := os.Rename(tempName, target); err != nil {
		http.Error(w, "publish report failed", http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"name": name,
		"size": len(content),
		"url":  "/api/reports/" + url.PathEscape(name),
	})
}

func isReportFilename(name string) bool {
	if name == "" || name == "." || name == ".." || len(name) > 180 {
		return false
	}
	if filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".htm" || ext == ".html"
}

func validBearerToken(authorization, expected string) bool {
	scheme, token, ok := strings.Cut(strings.TrimSpace(authorization), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return false
	}
	token = strings.TrimSpace(token)
	if token == "" || strings.ContainsAny(token, " \t\r\n") {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
