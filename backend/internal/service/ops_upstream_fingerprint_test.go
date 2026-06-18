package service

import (
	"net/http"
	"strings"
	"testing"
)

func TestBuildOpsUpstreamFingerprintKeepsSafeHeadersAndRedacts(t *testing.T) {
	h := http.Header{}
	h.Set("Server", "cloudflare")
	h.Set("CF-Ray", "abc123-HKG")
	h.Set("X-Request-Id", "req_123")
	h.Set("OpenAI-Processing-MS", "42")
	h.Set("Via", "Bearer abcdefghijklmnop")
	h.Set("Authorization", "Bearer sk-secret")
	h.Set("Set-Cookie", "sid=secret")
	h.Set("X-Debug-Token", "sk-verysecretvalue")

	fp := buildOpsUpstreamFingerprint(h)
	if fp == nil {
		t.Fatal("expected fingerprint")
	}
	if fp.Headers["server"] != "cloudflare" || fp.Headers["cf-ray"] != "abc123-HKG" || fp.Headers["x-request-id"] != "req_123" {
		t.Fatalf("safe headers missing: %#v", fp.Headers)
	}
	if _, ok := fp.Headers["authorization"]; ok {
		t.Fatalf("authorization must not be captured: %#v", fp.Headers)
	}
	if _, ok := fp.Headers["set-cookie"]; ok {
		t.Fatalf("set-cookie must not be captured: %#v", fp.Headers)
	}
	if got := fp.Headers["openai-processing-ms"]; got != "42" {
		t.Fatalf("openai-processing-ms = %q", got)
	}
	if got := fp.Headers["via"]; got != "[redacted]" {
		t.Fatalf("via secret was not redacted: %q", got)
	}
	if _, ok := fp.Headers["x-debug-token"]; ok {
		t.Fatalf("unknown token header must not be captured: %#v", fp.Headers)
	}
}

func TestAppendOpsUpstreamErrorAddsFingerprintFromHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Server", "nginx")
	h.Set("Via", "1.1 proxy")

	ev := OpsUpstreamErrorEvent{Message: "boom"}
	ev.AttachResponseHeaders(h)

	if ev.Fingerprint == nil {
		t.Fatal("expected fingerprint")
	}
	if ev.Fingerprint.Headers["server"] != "nginx" || ev.Fingerprint.Headers["via"] != "1.1 proxy" {
		t.Fatalf("unexpected fingerprint: %#v", ev.Fingerprint.Headers)
	}
}

func TestSanitizeOpsUpstreamErrorsPreservesFingerprintHeaders(t *testing.T) {
	entry := &OpsInsertErrorLogInput{
		UpstreamErrors: []*OpsUpstreamErrorEvent{
			{
				Message: "boom",
				Fingerprint: &OpsUpstreamFingerprint{Headers: map[string]string{
					"Server":        "nginx",
					"x-request-id":  "req_123",
					"authorization": "Bearer should-not-exist",
				}},
			},
		},
	}
	if err := sanitizeOpsUpstreamErrors(entry); err != nil {
		t.Fatalf("sanitizeOpsUpstreamErrors returned error: %v", err)
	}
	if entry.UpstreamErrorsJSON == nil {
		t.Fatal("expected json")
	}
	raw := *entry.UpstreamErrorsJSON
	if !strings.Contains(raw, "fingerprint") || !strings.Contains(raw, "server") || !strings.Contains(raw, "x-request-id") {
		t.Fatalf("fingerprint headers missing: %s", raw)
	}
	if strings.Contains(raw, "should-not-exist") || strings.Contains(raw, "authorization") {
		t.Fatalf("sensitive header leaked: %s", raw)
	}
}
