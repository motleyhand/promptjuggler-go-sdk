package promptjuggler_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"

	promptjuggler "go.promptjuggler.com/sdk"
)

const (
	webhookSecret  = "whsec_test"
	webhookPayload = `{"event":"promptrun.finished","id":"run1"}`
	webhookTS      = int64(1_700_000_000)
)

func signHeader(payload, secret string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "." + payload))
	return "t=" + strconv.FormatInt(ts, 10) + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyAcceptsCorrectSignature(t *testing.T) {
	if !promptjuggler.VerifyWebhookSignatureAt(
		webhookPayload,
		signHeader(webhookPayload, webhookSecret, webhookTS),
		webhookSecret,
		300,
		webhookTS,
	) {
		t.Error("expected valid signature to verify")
	}
}

func TestVerifyRejectsTamperedPayload(t *testing.T) {
	if promptjuggler.VerifyWebhookSignatureAt(
		webhookPayload+" ",
		signHeader(webhookPayload, webhookSecret, webhookTS),
		webhookSecret,
		300,
		webhookTS,
	) {
		t.Error("expected tampered payload to be rejected")
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	if promptjuggler.VerifyWebhookSignatureAt(
		webhookPayload,
		signHeader(webhookPayload, webhookSecret, webhookTS),
		"whsec_wrong",
		300,
		webhookTS,
	) {
		t.Error("expected wrong secret to be rejected")
	}
}

func TestVerifyRejectsExpiredTimestamp(t *testing.T) {
	if promptjuggler.VerifyWebhookSignatureAt(
		webhookPayload,
		signHeader(webhookPayload, webhookSecret, webhookTS),
		webhookSecret,
		300,
		webhookTS+301,
	) {
		t.Error("expected expired timestamp to be rejected")
	}
}

func TestVerifyAcceptsTimestampAtEdge(t *testing.T) {
	if !promptjuggler.VerifyWebhookSignatureAt(
		webhookPayload,
		signHeader(webhookPayload, webhookSecret, webhookTS),
		webhookSecret,
		300,
		webhookTS+300,
	) {
		t.Error("expected timestamp at the tolerance edge to verify")
	}
}

func TestVerifyRejectsMalformedHeader(t *testing.T) {
	if promptjuggler.VerifyWebhookSignatureAt(
		webhookPayload,
		"not-a-signature",
		webhookSecret,
		300,
		webhookTS,
	) {
		t.Error("expected malformed header to be rejected")
	}
}

func TestVerifyRejectsEmptyHeader(t *testing.T) {
	if promptjuggler.VerifyWebhookSignature(webhookPayload, "", webhookSecret) {
		t.Error("expected empty header to be rejected")
	}
}
