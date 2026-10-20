package webhook

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"hash"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var verifyNow = time.Unix(1_800_000_000, 0)

func hmacHex(h func() hash.Hash, key []byte, message string) string {
	mac := hmac.New(h, key)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignatureHMAC(t *testing.T) {
	sig := hmacHex(sha256.New, []byte("key"), "message")

	valid, err := VerifySignature(AlgorithmHMACSHA256, "key", "message", sig, EncodingHex)
	require.NoError(t, err)
	assert.True(t, valid)

	for name, args := range map[string][3]string{
		"wrong key":       {"other", "message", sig},
		"wrong message":   {"key", "message ", sig},
		"not hex":         {"key", "message", "zz"},
		"empty signature": {"key", "message", ""},
	} {
		valid, err := VerifySignature(AlgorithmHMACSHA256, args[0], args[1], args[2], EncodingHex)
		require.NoError(t, err, name)
		assert.False(t, valid, name)
	}

	sha1Sig := hmac.New(sha1.New, []byte("key"))
	sha1Sig.Write([]byte("message"))
	valid, err = VerifySignature(AlgorithmHMACSHA1, "key", "message", base64.StdEncoding.EncodeToString(sha1Sig.Sum(nil)), EncodingBase64)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifySignatureEd25519(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sig := hex.EncodeToString(ed25519.Sign(privateKey, []byte("message")))

	spki, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)
	keys := map[string]string{
		"hex":         hex.EncodeToString(publicKey),
		"base64":      base64.StdEncoding.EncodeToString(publicKey),
		"spki base64": base64.StdEncoding.EncodeToString(spki),
		"pem":         string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: spki})),
	}
	for name, key := range keys {
		valid, err := VerifySignature(AlgorithmEd25519, key, "message", sig, EncodingHex)
		require.NoError(t, err, name)
		assert.True(t, valid, name)

		valid, err = VerifySignature(AlgorithmEd25519, key, "other", sig, EncodingHex)
		require.NoError(t, err, name)
		assert.False(t, valid, name)

		// A signature of the wrong size doesn't panic.
		valid, err = VerifySignature(AlgorithmEd25519, key, "message", "abcd", EncodingHex)
		require.NoError(t, err, name)
		assert.False(t, valid, name)
	}
}

func TestVerifySignatureSettingErrors(t *testing.T) {
	_, err := VerifySignature("md5", "key", "message", "", EncodingHex)
	assert.Error(t, err)
	_, err = VerifySignature(AlgorithmHMACSHA256, "", "message", "", EncodingHex)
	assert.Error(t, err)
	_, err = VerifySignature(AlgorithmEd25519, "not a key", "message", "", EncodingHex)
	assert.Error(t, err)
	_, err = VerifySignature(AlgorithmHMACSHA256, "key", "message", "", "base32")
	assert.Error(t, err)
	_, err = VerifyPreset("unknown", "key", &Event{ReceivedAt: verifyNow})
	assert.Error(t, err)
}

func TestVerifyPresetGitHub(t *testing.T) {
	event := &Event{Body: `{"ok":true}`, Headers: map[string]string{
		"x-hub-signature-256": "sha256=" + hmacHex(sha256.New, []byte("secret"), `{"ok":true}`),
	}}
	assertPreset(t, "github", "secret", event)
}

func TestVerifyPresetSlack(t *testing.T) {
	ts := strconv.FormatInt(verifyNow.Unix(), 10)
	event := &Event{Body: "a=b", Headers: map[string]string{
		"x-slack-request-timestamp": ts,
		"x-slack-signature":         "v0=" + hmacHex(sha256.New, []byte("secret"), "v0:"+ts+":a=b"),
	}}
	assertPreset(t, "slack", "secret", event)
}

func TestVerifyPresetStripe(t *testing.T) {
	ts := strconv.FormatInt(verifyNow.Unix(), 10)
	sig := hmacHex(sha256.New, []byte("whsec_secret"), ts+".{}")
	event := &Event{Body: "{}", Headers: map[string]string{
		// The second v1 is for a rolled secret.
		"stripe-signature": fmt.Sprintf("t=%s,v1=%s,v1=%s", ts, hmacHex(sha256.New, []byte("old"), ts+".{}"), sig),
	}}
	assertPreset(t, "stripe", "whsec_secret", event)
}

func TestVerifyPresetStandardWebhooks(t *testing.T) {
	secret := []byte("0123456789abcdef")
	key := "whsec_" + base64.StdEncoding.EncodeToString(secret)
	ts := strconv.FormatInt(verifyNow.Unix(), 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("msg_1." + ts + ".{}"))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	for _, prefix := range []string{"webhook-", "svix-"} {
		event := &Event{Body: "{}", Headers: map[string]string{
			prefix + "id":        "msg_1",
			prefix + "timestamp": ts,
			prefix + "signature": "v1,b3RoZXI= v1," + sig,
		}}
		assertPreset(t, "standard_webhooks", key, event)
	}

	_, err := VerifyPreset("standard_webhooks", "whsec_!!", &Event{ReceivedAt: verifyNow})
	assert.Error(t, err)
}

func TestVerifyPresetStandardWebhooksEd25519(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	ts := strconv.FormatInt(verifyNow.Unix(), 10)
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, []byte("msg_1."+ts+".{}")))

	event := &Event{Body: "{}", Headers: map[string]string{
		"webhook-id":        "msg_1",
		"webhook-timestamp": ts,
		// A v1 signature with the same value isn't checked as Ed25519.
		"webhook-signature": "v1," + sig + " v1a,b3RoZXI= v1a," + sig,
	}}
	assertPreset(t, "standard_webhooks", "whpk_"+base64.StdEncoding.EncodeToString(publicKey), event)

	onlyV1 := receivedAt(event, verifyNow)
	onlyV1.Headers = map[string]string{"webhook-id": "msg_1", "webhook-timestamp": ts, "webhook-signature": "v1," + sig}
	valid, err := VerifyPreset("standard_webhooks", "whpk_"+base64.StdEncoding.EncodeToString(publicKey), onlyV1)
	require.NoError(t, err)
	assert.False(t, valid)

	_, err = VerifyPreset("standard_webhooks", "whpk_AAAA", &Event{ReceivedAt: verifyNow})
	assert.Error(t, err)
}

func TestVerifyPresetDiscord(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	ts := strconv.FormatInt(verifyNow.Unix(), 10)
	event := &Event{Body: "{}", Headers: map[string]string{
		"x-signature-timestamp": ts,
		"x-signature-ed25519":   hex.EncodeToString(ed25519.Sign(privateKey, []byte(ts+"{}"))),
	}}
	assertPreset(t, "discord", hex.EncodeToString(publicKey), event)
}

// ER:LC's key is built in, so only a request ER:LC didn't sign can be tested.
func TestVerifyPresetERLC(t *testing.T) {
	_, err := parseEd25519PublicKey(hex.EncodeToString(erlcPublicKey))
	require.NoError(t, err)

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	ts := strconv.FormatInt(verifyNow.Unix(), 10)
	event := &Event{Body: "{}", Headers: map[string]string{
		"x-signature-timestamp": ts,
		"x-signature-ed25519":   hex.EncodeToString(ed25519.Sign(privateKey, []byte(ts+"{}"))),
	}}
	valid, err := VerifyPreset("erlc", "", receivedAt(event, verifyNow))
	require.NoError(t, err)
	assert.False(t, valid)
}

// assertPreset checks that event is valid, and invalid once its body changes,
// it's too old, or its headers are missing.
func assertPreset(t *testing.T, preset string, key string, event *Event) {
	t.Helper()

	valid, err := VerifyPreset(preset, key, receivedAt(event, verifyNow))
	require.NoError(t, err)
	assert.True(t, valid, "valid")

	changed := *event
	changed.Body += " "
	valid, err = VerifyPreset(preset, key, receivedAt(&changed, verifyNow))
	require.NoError(t, err)
	assert.False(t, valid, "changed body")

	valid, err = VerifyPreset(preset, key, &Event{Body: event.Body, ReceivedAt: verifyNow})
	require.NoError(t, err)
	assert.False(t, valid, "no headers")

	if preset != "github" {
		valid, err = VerifyPreset(preset, key, receivedAt(event, verifyNow.Add(MaxSignatureAge+time.Second)))
		require.NoError(t, err)
		assert.False(t, valid, "too old")
	}
}

func receivedAt(event *Event, t time.Time) *Event {
	res := *event
	res.ReceivedAt = t
	return &res
}
