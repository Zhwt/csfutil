package csfutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// DecodeUTF16 decodes UTF-16 string stored in []byte slice and returns the
// string. Returns error only if b is not multiple of 2.
func DecodeUTF16(b []byte) (string, error) {
	if len(b) == 0 {
		return "", nil
	}

	if len(b)%2 != 0 {
		return "", fmt.Errorf("not valid UTF-16 bytes, length needs to be multiple of 2, got length: %d", len(b))
	}

	ints := make([]uint16, len(b)/2)
	binary.Read(bytes.NewReader(b), binary.LittleEndian, &ints)

	return string(utf16.Decode(ints)), nil
}

// EncodeUTF16 encodes a string to its corresponding UTF-16 form.
func EncodeUTF16(s string) []byte {
	codes := utf16.Encode([]rune(s))

	b := make([]byte, len(codes)*2)
	for i, r := range codes {
		b[i*2] = byte(r)
		b[i*2+1] = byte(r >> 8)
	}

	return b
}
