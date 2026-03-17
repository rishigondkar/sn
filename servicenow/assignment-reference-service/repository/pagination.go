package repository

import "strconv"

// parseOffsetToken parses an opaque page token as offset (integer). Returns nil if invalid.
func parseOffsetToken(token string) *PageToken {
	offset, err := strconv.Atoi(token)
	if err != nil || offset < 0 {
		return nil
	}
	return &PageToken{Offset: offset}
}

// encodeOffsetToken encodes offset as page token.
func encodeOffsetToken(offset int) string {
	if offset <= 0 {
		return ""
	}
	return strconv.Itoa(offset)
}
