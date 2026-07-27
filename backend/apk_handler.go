package main

import (
	"net/http"
	"os"
)

const apkDownloadName = "fitness-tracker.apk"

func resolveAPKPath() string {
	return os.Getenv("APK_FILE")
}

func apkDownloadHandler(apkPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if apkPath == "" {
			http.NotFound(w, r)
			return
		}

		apk, err := os.Open(apkPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "open APK", http.StatusInternalServerError)
			return
		}
		defer apk.Close()

		info, err := apk.Stat()
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/vnd.android.package-archive")
		w.Header().Set("Content-Disposition", `attachment; filename="`+apkDownloadName+`"`)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(w, r, apkDownloadName, info.ModTime(), apk)
	}
}
