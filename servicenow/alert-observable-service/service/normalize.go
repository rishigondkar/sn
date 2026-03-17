package service

import (
	"net"
	"regexp"
	"strings"
	"unicode"
)

// Allowed observable types. Keys are stored in DB (lowercase). "ip" kept for backward compatibility (treated as ipv4).
var allowedObservableTypes = map[string]bool{
	"ip": true, "ipv4": true, "ipv6": true,
	"md5": true, "sha1": true, "sha256": true, "sha384": true, "sha512": true,
	"pehash": true, "imphash": true,
	"file": true, "file_path": true, "file_name": true, "windows_executable": true,
	"domain": true, "hostname": true, "tld": true,
	"email": true, "email_subject": true, "email_message_id": true, "email_body": true,
	"url": true, "uri": true,
	"registry_key": true,
	"cidr": true, "ipv4_network": true, "ipv6_network": true, "ipv4_netmask": true, "ipv6_netmask": true,
	"mutex": true, "cve": true, "asn": true,
	"mac_address": true, "atm_address": true,
	"username": true, "postal_address": true, "certificate_serial": true, "organization": true, "phone_number": true,
	"observable_composition": true, "command_line": true,
}

var (
	hex32  = regexp.MustCompile(`(?i)^[a-f0-9]{32}$`)
	hex40  = regexp.MustCompile(`(?i)^[a-f0-9]{40}$`)
	hex64  = regexp.MustCompile(`(?i)^[a-f0-9]{64}$`)
	hex96  = regexp.MustCompile(`(?i)^[a-f0-9]{96}$`)
	hex128 = regexp.MustCompile(`(?i)^[a-f0-9]{128}$`)
	cveRe  = regexp.MustCompile(`(?i)^CVE-\d{4}-\d{4,}$`)
	asnRe  = regexp.MustCompile(`(?i)^ASN?\s*\d+$|^AS\d+$`)
	// MAC: xx:xx:xx:xx:xx:xx or xx-xx-xx-xx-xx-xx (hex pairs)
	macRe       = regexp.MustCompile(`(?i)^([0-9a-f]{2}[:-]){5}[0-9a-f]{2}$`)
	domainLike  = regexp.MustCompile(`(?i)^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)
	registryRe  = regexp.MustCompile(`(?i)^(HKLM|HKCU|HKCR|HKU|HKCC|HKEY_)[\w\-\.\\]*$`)
	mutexRe     = regexp.MustCompile(`(?i)^(Global\\|Local\\)?[A-Za-z0-9_\-\.]+$`)
	messageIdRe = regexp.MustCompile(`^<[^>]+@[^>]+>$`)
)

// ClassifyObservableValue infers the observable type from a raw value and returns the suggested type and normalized value.
// Order: URL, URI, email, CVE, ASN, IPv4, IPv6, MAC, hashes (by length), registry, CIDR, domain/hostname/TLD, file path/name, mutex, email parts, phone, username, fallbacks.
func ClassifyObservableValue(value string) (observableType, normalized string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	v := value

	// URL: contains :// or starts with http(s)://
	if strings.Contains(v, "://") || strings.HasPrefix(strings.ToLower(v), "http://") || strings.HasPrefix(strings.ToLower(v), "https://") {
		return "url", normalizeURL(v)
	}
	// Email: single @ with local and domain parts
	if idx := strings.Index(v, "@"); idx > 0 && idx < len(v)-1 && !strings.Contains(v[idx+1:], "@") {
		return "email", normalizeEmail(v)
	}
	// CVE
	if cveRe.MatchString(v) {
		return "cve", strings.ToUpper(strings.TrimSpace(v))
	}
	// ASN
	if asnRe.MatchString(v) {
		return "asn", strings.ToUpper(strings.TrimSpace(v))
	}
	// IP (IPv4 vs IPv6) - before URI so IPv6 with : doesn't match URI
	if ip := net.ParseIP(v); ip != nil {
		if ip.To4() != nil {
			return "ipv4", ip.String()
		}
		return "ipv6", ip.String()
	}
	// MAC address - before URI so xx:xx:xx doesn't match URI
	if macRe.MatchString(v) {
		return "mac_address", normalizeMAC(v)
	}
	// URI: urn: or other scheme without ://
	if strings.HasPrefix(strings.ToLower(v), "urn:") || (strings.Contains(v, ":") && !strings.Contains(v, "://") && len(v) > 4 && v[0] != '/') {
		return "uri", strings.TrimSpace(v)
	}
	// Hashes by length (hex only)
	lower := strings.ToLower(strings.TrimSpace(v))
	if regexp.MustCompile(`(?i)^[a-f0-9]+$`).MatchString(v) {
		switch len(lower) {
		case 32:
			return "md5", lower
		case 40:
			return "sha1", lower
		case 64:
			return "sha256", lower
		case 96:
			return "sha384", lower
		case 128:
			return "sha512", lower
		}
	}
	// PEHASH / IMPHASH (often with prefix; if 64 hex then pehash, 32 hex already caught as md5 - accept optional prefix)
	if strings.HasPrefix(lower, "pehash:") || strings.HasPrefix(lower, "imphash:") {
		s := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(lower, "pehash:"), "imphash:"))
		if regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(s) {
			return "pehash", s
		}
		if regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(s) {
			return "imphash", s
		}
	}
	// Registry key
	if registryRe.MatchString(v) || strings.HasPrefix(strings.ToUpper(v), "HKEY_") {
		return "registry_key", strings.TrimSpace(v)
	}
	// CIDR or network (contains / and parses as IP network)
	if strings.Contains(v, "/") {
		if _, network, err := net.ParseCIDR(v); err == nil && network != nil {
			return "cidr", network.String()
		}
		// Netmask notation (e.g. 192.168.0.0/255.255.255.0) - treat as network
		if strings.Count(v, ".") >= 2 {
			return "ipv4_network", strings.TrimSpace(v)
		}
		if strings.Contains(v, ":") {
			return "ipv6_network", strings.TrimSpace(v)
		}
	}
	// Mutex (Global\ or Local\ with single segment - before file_path)
	var rest string
	if strings.HasPrefix(v, "Global\\") {
		rest = v[8:]
	} else if strings.HasPrefix(v, "Local\\") {
		rest = v[7:]
	}
	if rest != "" && !strings.Contains(rest, "\\") {
		return "mutex", strings.TrimSpace(v)
	}
	// File path (contains path separators)
	if strings.Contains(v, "\\") || strings.Contains(v, "/") {
		if hasWindowsExeExtension(v) {
			return "windows_executable", strings.TrimSpace(v)
		}
		return "file_path", strings.TrimSpace(v)
	}
	// File name (has extension, no path)
	if hasFileExtension(v) {
		if hasWindowsExeExtension(v) {
			return "windows_executable", strings.ToLower(strings.TrimSpace(v))
		}
		return "file_name", strings.ToLower(strings.TrimSpace(v))
	}
	// Domain: hostname-like with at least one dot (TLD)
	if domainLike.MatchString(v) && strings.Contains(v, ".") {
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			last := strings.ToLower(parts[len(parts)-1])
			// Common TLDs only in last segment and multiple parts -> domain
			if len(parts) > 1 && len(last) >= 2 && len(last) <= 6 {
				return "domain", normalizeDomain(v)
			}
		}
		return "domain", normalizeDomain(v)
	}
	// Hostname (letters, digits, dots, hyphens; single label or no dot)
	if domainLike.MatchString(v) {
		return "hostname", normalizeDomain(v)
	}
	// TLD: single short label (e.g. "com", "org")
	if domainLike.MatchString(v) && !strings.Contains(v, ".") && len(v) >= 2 && len(v) <= 10 {
		return "tld", strings.ToLower(v)
	}
	// Mutex (Global\Name or Local\Name or plain name that looks like mutex)
	if mutexRe.MatchString(v) || (strings.HasPrefix(v, "Global\\") || strings.HasPrefix(v, "Local\\")) {
		return "mutex", strings.TrimSpace(v)
	}
	// Email Message-ID (angle brackets with @)
	if messageIdRe.MatchString(v) {
		return "email_message_id", strings.TrimSpace(v)
	}
	// Email subject (starts with Subject: or looks like a subject line)
	if strings.HasPrefix(strings.ToLower(v), "subject:") {
		return "email_subject", strings.TrimSpace(v)
	}
	// Command line (starts with common shells or has multiple space-separated parts that look like args)
	if strings.HasPrefix(strings.ToLower(v), "cmd.exe") || strings.HasPrefix(strings.ToLower(v), "/c ") || strings.HasPrefix(strings.ToLower(v), "powershell") {
		return "command_line", strings.TrimSpace(v)
	}
	// Phone number (digits, optional + at start, spaces/dashes/dots)
	if regexp.MustCompile(`^\+?[\d\s\-\.\(\)]{10,}$`).MatchString(v) && regexp.MustCompile(`\d`).MatchString(v) {
		return "phone_number", normalizePhone(v)
	}
	// Username (single word, alphanumeric and common chars, no @)
	if regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`).MatchString(v) && len(v) <= 128 && !strings.Contains(v, "@") {
		return "username", strings.TrimSpace(v)
	}
	// Certificate serial (hex or decimal digits)
	if regexp.MustCompile(`(?i)^[a-f0-9]+$`).MatchString(v) && len(v) >= 8 && len(v) <= 64 {
		return "certificate_serial", strings.ToLower(v)
	}
	// Observable composition / multi-line or comma-separated -> generic
	if strings.Contains(v, "\n") || (strings.Contains(v, ",") && len(strings.Fields(v)) > 3) {
		return "observable_composition", strings.TrimSpace(v)
	}
	// Default: file (generic)
	return "file", strings.ToLower(strings.TrimSpace(v))
}

func hasFileExtension(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return true
	}
	exts := []string{".exe", ".dll", ".pdf", ".doc", ".docx", ".js", ".zip", ".tar", ".gz", ".sh", ".bat", ".ps1", ".py", ".txt", ".csv", ".log", ".sys", ".scr"}
	for _, ext := range exts {
		if strings.HasSuffix(s, ext) {
			return true
		}
	}
	return false
}

func hasWindowsExeExtension(s string) bool {
	lower := strings.ToLower(s)
	return strings.HasSuffix(lower, ".exe") || strings.HasSuffix(lower, ".dll") || strings.HasSuffix(lower, ".scr") || strings.HasSuffix(lower, ".sys")
}

// AllowedTrackingStatus values.
var allowedTrackingStatus = map[string]bool{
	"new": true, "under_review": true, "enriched": true, "dismissed": true, "confirmed": true,
}

// NormalizeObservableValue returns the canonical normalized value for the given type and value.
func NormalizeObservableValue(observableType, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	t := strings.ToLower(strings.TrimSpace(observableType))
	switch t {
	case "ip", "ipv4", "ipv6", "ipv4_network", "ipv6_network", "ipv4_netmask", "ipv6_netmask", "cidr":
		return normalizeIP(value)
	case "domain", "hostname", "tld":
		return normalizeDomain(value)
	case "md5", "sha1", "sha256", "sha384", "sha512", "pehash", "imphash":
		return normalizeHash(value)
	case "url":
		return normalizeURL(value)
	case "uri":
		return strings.TrimSpace(value)
	case "email":
		return normalizeEmail(value)
	case "mac_address":
		return normalizeMAC(value)
	case "file", "file_path", "file_name", "windows_executable", "registry_key", "mutex", "email_subject", "email_message_id", "email_body",
		"username", "postal_address", "organization", "observable_composition", "command_line":
		return strings.ToLower(strings.TrimSpace(value))
	case "cve":
		return strings.ToUpper(strings.TrimSpace(value))
	case "asn", "atm_address":
		return strings.ToUpper(strings.TrimSpace(value))
	case "certificate_serial":
		return strings.ToLower(strings.TrimSpace(value))
	case "phone_number":
		return normalizePhone(value)
	default:
		return strings.ToLower(value)
	}
}

func normalizeIP(s string) string {
	ip := net.ParseIP(s)
	if ip == nil {
		return strings.ToLower(strings.TrimSpace(s))
	}
	return ip.String()
}

func normalizeDomain(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeHash(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeURL(s string) string {
	return strings.TrimSpace(s)
}

func normalizeEmail(s string) string {
	at := strings.Index(s, "@")
	if at <= 0 {
		return strings.ToLower(strings.TrimSpace(s))
	}
	local := strings.ToLower(strings.TrimSpace(s[:at]))
	domain := strings.ToLower(strings.TrimSpace(s[at+1:]))
	return local + "@" + domain
}

func normalizeMAC(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Prefer colon form
	s = strings.ReplaceAll(s, "-", ":")
	return s
}

func normalizePhone(s string) string {
	// Keep only digits and leading + for normalization
	var b strings.Builder
	for _, r := range s {
		if r == '+' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ValidateObservableType returns true if the type is allowed.
func ValidateObservableType(t string) bool {
	return allowedObservableTypes[strings.ToLower(strings.TrimSpace(t))]
}

// ValidateTrackingStatus returns true if the status is allowed.
func ValidateTrackingStatus(s string) bool {
	if s == "" {
		return true
	}
	return allowedTrackingStatus[strings.ToLower(strings.TrimSpace(s))]
}

// isPrintableASCII returns true if s is safe for storage (no control chars).
func isPrintableASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || (r < 0x20 && r != '\t') {
			return false
		}
	}
	return true
}
