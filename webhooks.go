package promptjuggler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// defaultWebhookTolerance is the replay window (seconds) for VerifyWebhookSignature.
const defaultWebhookTolerance = 300

// VerifyWebhookSignature reports whether signatureHeader is a valid signature for payload, using
// the default 300-second replay tolerance and the current time.
//
// payload must be the raw request body, exactly as received (verify before JSON parsing).
// signatureHeader is the PromptJuggler-Signature header value (t=<ts>,v1=<hmac>); an empty header
// verifies as false.
func VerifyWebhookSignature(payload, signatureHeader, secret string) bool {
	return VerifyWebhookSignatureAt(
		payload,
		signatureHeader,
		secret,
		defaultWebhookTolerance,
		time.Now().Unix(),
	)
}

// VerifyWebhookSignatureAt is VerifyWebhookSignature with an explicit tolerance (seconds) and
// clock (now, a Unix timestamp). It recomputes the HMAC-SHA256 of "<timestamp>.<payload>",
// compares it in constant time, and rejects deliveries whose timestamp falls outside tolerance
// seconds of now.
func VerifyWebhookSignatureAt(payload, signatureHeader, secret string, tolerance, now int64) bool {
	var timestamp int64
	var signature string
	haveTimestamp := false
	for _, part := range strings.Split(signatureHeader, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "t":
			ts, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err != nil {
				return false
			}
			timestamp, haveTimestamp = ts, true
		case "v1":
			signature = strings.TrimSpace(value)
		}
	}
	if !haveTimestamp || signature == "" {
		return false
	}
	if delta := now - timestamp; delta > tolerance || delta < -tolerance {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10) + "." + payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
