package instagrapi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
)

// passwordEncrypt encrypts a password using Instagram's RSA+AES scheme.
func (c *Client) passwordEncrypt(password string) string {
	publicKeyID, publicKey := c.getPasswordPublicKey()
	if publicKey == "" {
		return ""
	}

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	iv := make([]byte, 12)
	rand.Read(iv)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return ""
	}

	recipientKey, err := x509.ParsePKIXPublicKey(decodedPublicKey)
	if err != nil {
		return ""
	}

	rsaPubKey, ok := recipientKey.(*rsa.PublicKey)
	if !ok {
		return ""
	}

	rsaEncrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, sessionKey)
	if err != nil {
		return ""
	}

	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return ""
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	nonce := make([]byte, aesGCM.NonceSize())
	rand.Read(nonce)

	taggedData := aesGCM.Seal(nil, iv, []byte(password), []byte(timestamp))

	sizeBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(sizeBuffer, uint16(len(rsaEncrypted)))

	publicKeyIDBytes := []byte(fmt.Sprintf("%d", publicKeyID))
	payload := append([]byte{0x01}, publicKeyIDBytes...)
	payload = append(payload, iv...)
	payload = append(payload, sizeBuffer...)
	payload = append(payload, rsaEncrypted...)
	payload = append(payload, taggedData...)

	encodedPayload := base64.StdEncoding.EncodeToString(payload)

	return fmt.Sprintf("#PWD_INSTAGRAM:4:%s:%s", timestamp, encodedPayload)
}

// PasswordPublicKeys fetches the password encryption public keys from Instagram.
func (c *Client) getPasswordPublicKey() (uint64, string) {
	resp, err := c.PublicClient.Get("https://i.instagram.com/api/v1/qe/sync/")
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()

	keyIDStr := resp.Header.Get("ig-set-password-encryption-key-id")
	publicKey := resp.Header.Get("ig-set-password-encryption-pub-key")

	var keyID uint64
	fmt.Sscanf(keyIDStr, "%d", &keyID)

	return keyID, publicKey
}

// GenerateTOTPSeed generates a TOTP seed for 2FA.
func (c *Client) GenerateTOTPSeed() (string, error) {
	data := map[string]any{
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	result, err := c.privateRequest(
		"accounts/generate_two_factor_totp_key/",
		data,
		nil,
	)
	if err != nil {
		return "", err
	}

	seed, _ := result["totp_seed"].(string)
	return seed, nil
}

// EnableTOTP enables TOTP 2FA.
func (c *Client) EnableTOTP(verificationCode string) ([]string, error) {
	data := map[string]any{
		"verification_code": verificationCode,
		"_uid":              strconv.FormatInt(c.UserID, 10),
		"_uuid":             c.UUID,
	}

	result, err := c.privateRequest(
		"accounts/enable_totp_two_factor/",
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var backupCodes []string
	codesArr := navigateJSON(result, "backup_codes")
	if codesList, ok := codesArr.([]any); ok {
		for _, code := range codesList {
			if s, ok := code.(string); ok {
				backupCodes = append(backupCodes, s)
			}
		}
	}

	return backupCodes, nil
}

// DisableTOTP disables TOTP 2FA.
func (c *Client) DisableTOTP() error {
	data := map[string]any{
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	result, err := c.privateRequest(
		"accounts/disable_totp_two_factor/",
		data,
		nil,
	)
	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: "Failed to disable TOTP"}
	}

	return nil
}

// GenerateTOTPCode generates a TOTP code from a seed.
func GenerateTOTPCode(seed string) (string, error) {
	decodedSeed, err := base64.StdEncoding.DecodeString(seed)
	if err != nil {
		return "", err
	}

	now := time.Now()
	timestep := uint64(now.Unix()) / 30

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], timestep)

	h := sha256.Sum256(append(decodedSeed, buf[:]...))

	var code int64
	for _, b := range h[:] {
		code = (code << 8) | int64(b)
	}
	code = code % 1000000

	return fmt.Sprintf("%06d", code), nil
}
