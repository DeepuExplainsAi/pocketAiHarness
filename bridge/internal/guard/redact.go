package guard

import "regexp"

const Redacted = "[REDACTED]"

var (
	keyValue = regexp.MustCompile(`(?i)\b([A-Za-z0-9_]*(?:token|secret|password|passwd|api[_-]?key)[A-Za-z0-9_]*)\s*([=:])\s*("[^"]*"|'[^']*'|\S+)`)
	patterns = []*regexp.Regexp{
		regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{20,}\b`),
		regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{20,}\b`),
		regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
		regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{16,}`),
		regexp.MustCompile(`\bya29\.[A-Za-z0-9._-]{20,}`),
		regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`),
		regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`),
	}
)

// Redact removes secrets from text before it is logged or written to diagnostics.
func Redact(s string) string {
	s = keyValue.ReplaceAllString(s, "${1}${2}"+Redacted)
	for _, p := range patterns {
		s = p.ReplaceAllString(s, Redacted)
	}
	return s
}
