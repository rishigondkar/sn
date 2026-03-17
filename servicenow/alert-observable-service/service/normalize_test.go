package service

import (
	"testing"
)

func TestNormalizeObservableValue(t *testing.T) {
	tests := []struct {
		name  string
		otype string
		value string
		want  string
	}{
		{"ip canonical", "ip", "192.168.1.1", "192.168.1.1"},
		{"ipv4 canonical", "ipv4", "192.168.1.1", "192.168.1.1"},
		{"ipv6 canonical", "ipv6", "2001:db8::1", "2001:db8::1"},
		{"domain lower", "domain", "Example.COM", "example.com"},
		{"md5 lower", "md5", "ABC123", "abc123"},
		{"sha256 lower", "sha256", "ABC123", "abc123"},
		{"email lower", "email", "User@Example.COM", "user@example.com"},
		{"url trim", "url", "  https://example.com/path  ", "https://example.com/path"},
		{"cve upper", "cve", "cve-2024-1234", "CVE-2024-1234"},
		{"unknown type fallback", "other", "MixedCase", "mixedcase"},
		{"empty value", "domain", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeObservableValue(tt.otype, tt.value)
			if got != tt.want {
				t.Errorf("NormalizeObservableValue(%q, %q) = %q, want %q", tt.otype, tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateObservableType(t *testing.T) {
	valid := []string{"ip", "ipv4", "ipv6", "domain", "md5", "sha256", "url", "email", "file", "cve", "asn", "registry_key", "command_line", "IP", "Domain"}
	for _, v := range valid {
		if !ValidateObservableType(v) {
			t.Errorf("ValidateObservableType(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "other", "observable", "invalid_type"}
	for _, v := range invalid {
		if ValidateObservableType(v) {
			t.Errorf("ValidateObservableType(%q) = true, want false", v)
		}
	}
}

func TestClassifyObservableValue(t *testing.T) {
	tests := []struct {
		value   string
		wantType string
	}{
		{"https://example.com/path", "url"},
		{"http://evil.com", "url"},
		{"user@example.com", "email"},
		{"192.168.1.1", "ipv4"},
		{"2001:db8::1", "ipv6"},
		{"d41d8cd98f00b204e9800998ecf8427e", "md5"},
		{"a94a8fe5ccb19ba61c4c0873d391e987982fbbd3", "sha1"},
		{"example.com", "domain"},
		{"sub.example.com", "domain"},
		{"malware.exe", "windows_executable"},
		{"CVE-2024-12345", "cve"},
		{"AS12345", "asn"},
		{"00:11:22:33:44:55", "mac_address"},
		{"HKEY_LOCAL_MACHINE\\Software\\Foo", "registry_key"},
		{"Global\\MyMutex", "mutex"},
		{"/usr/bin/bash", "file_path"},
		{"192.168.0.0/24", "cidr"},
	}
	for _, tt := range tests {
		gotType, _ := ClassifyObservableValue(tt.value)
		if gotType != tt.wantType {
			t.Errorf("ClassifyObservableValue(%q) type = %q, want %q", tt.value, gotType, tt.wantType)
		}
	}
}

func TestValidateTrackingStatus(t *testing.T) {
	valid := []string{"new", "under_review", "enriched", "dismissed", "confirmed", ""}
	for _, v := range valid {
		if !ValidateTrackingStatus(v) {
			t.Errorf("ValidateTrackingStatus(%q) = false, want true", v)
		}
	}
	if ValidateTrackingStatus("invalid") {
		t.Error("ValidateTrackingStatus(\"invalid\") = true, want false")
	}
}
