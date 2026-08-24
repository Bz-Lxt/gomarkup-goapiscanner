package main

import (
	"net/http"
	"strings"
	"time"
)

func usersSQLi(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	w.Header().Set("Content-Type", "application/json")
	if looksSQLi(id) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"sql_error":"You have an error in your SQL syntax","rows":[{"id":1},{"id":2}]}`))
		return
	}
	_, _ = w.Write([]byte(`{"id":"` + jsonEscape(id) + `"}`))
}

func usersBlind(w http.ResponseWriter, r *http.Request) {
	id := strings.ToUpper(r.URL.Query().Get("id"))
	w.Header().Set("Content-Type", "application/json")
	if strings.Contains(id, "SLEEP") || strings.Contains(id, "WAITFOR") || strings.Contains(id, "BENCHMARK") {
		time.Sleep(3 * time.Second)
		_, _ = w.Write([]byte(`{"ok":true,"blind":true}`))
		return
	}
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func searchXSS(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte("<html><body>Results for: " + q + "</body></html>"))
}

func adminSecret(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Lab-Secret", "exposed")
	_, _ = w.Write([]byte(`{"admin_token":"lab-root-token","role":"admin"}`))
}

func filesTraversal(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query().Get("path")
	w.Header().Set("Content-Type", "text/plain")
	if strings.Contains(p, "..") || strings.Contains(p, "etc/passwd") || strings.Contains(p, "etc%2fpasswd") {
		_, _ = w.Write([]byte("root:x:0:0:lab_passwd_leak"))
		return
	}
	_, _ = w.Write([]byte("file-ok"))
}

func pingCMDi(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	w.Header().Set("Content-Type", "text/plain")
	if strings.ContainsAny(host, ";|`$") || strings.Contains(host, "$(") || strings.Contains(host, "&&") {
		_, _ = w.Write([]byte("uid=0(root) lab_cmd_ok"))
		return
	}
	_, _ = w.Write([]byte("pong"))
}

func looksSQLi(s string) bool {
	u := strings.ToUpper(s)
	return strings.Contains(s, "'") || strings.Contains(s, "\"") ||
		strings.Contains(u, " OR ") || strings.Contains(u, "UNION") ||
		strings.Contains(u, "SLEEP") || strings.Contains(u, "--")
}

func jsonEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
