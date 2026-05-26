package instagrapi

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
)

// GenerateSignature creates the Instagram POST data signature.
func GenerateSignature(data string) string {
	return "signed_body=SIGNATURE." + urlEncode(data)
}

// GenToken generates a random token of specified length.
func GenToken(size int, withSymbols bool) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const symbols = "!@#$%^&*()_+-=[]{}|;:',.<>?/"
	charSet := letters
	if withSymbols {
		charSet += symbols
	}
	result := make([]byte, size)
	for i := range result {
		result[i] = charSet[randInt(len(charSet))]
	}
	return string(result)
}

// GenerateJazoest computes the jazoest parameter from a string.
func GenerateJazoest(symbols string) string {
	var sum int
	for _, r := range symbols {
		sum += int(r)
	}
	return "2" + strconv.Itoa(sum)
}

// InstagramIdCodec handles encoding/decoding Instagram IDs to/from shortcodes.
type InstagramIdCodec struct{}

var encodingChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// Encode converts a numeric ID to an Instagram shortcode.
func (c *InstagramIdCodec) Encode(num int64) string {
	if num == 0 {
		return string(encodingChars[0])
	}
	var result []byte
	base := len(encodingChars)
	for num > 0 {
		rem := int(num % int64(base))
		num /= int64(base)
		result = append(result, encodingChars[rem])
	}
	// reverse
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// Decode converts an Instagram shortcode to a numeric ID.
func (c *InstagramIdCodec) Decode(shortcode string) int64 {
	base := len(encodingChars)
	var num int64
	for i, ch := range shortcode {
		pos := strings.IndexRune(encodingChars, ch)
		if pos < 0 {
			continue
		}
		power := len(shortcode) - (i + 1)
		num += int64(pos) * big.NewInt(int64(base)).Exp(big.NewInt(int64(base)), big.NewInt(int64(power)), nil).Int64()
	}
	return num
}

// HashPassword creates a SHA-256 hash of the password (hex-encoded).
func HashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}
