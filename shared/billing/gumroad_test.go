package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyWebhookHMAC(t *testing.T) {
	secret := "shh"
	body := []byte("resource_name=sale&email=a@b.com")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	g := NewGumroadHandler(secret, "", nil)
	if !g.verifyWebhook(body, sig) {
		t.Error("valid signature rejected")
	}
	if g.verifyWebhook(body, "deadbeef") {
		t.Error("bad signature accepted")
	}
	if g.verifyWebhook(body, "") {
		t.Error("missing signature accepted")
	}
}

func TestVerifyWebhookSellerFallback(t *testing.T) {
	g := NewGumroadHandler("", "seller123", nil)
	if !g.verifyWebhook([]byte("seller_id=seller123"), "") {
		t.Error("valid seller_id rejected")
	}
	if g.verifyWebhook([]byte("seller_id=other"), "") {
		t.Error("wrong seller_id accepted")
	}
}

func TestVerifyWebhookFailsClosed(t *testing.T) {
	g := NewGumroadHandler("", "", nil)
	if g.verifyWebhook([]byte("x=1"), "anything") {
		t.Error("no secret + no seller should fail closed")
	}
}
