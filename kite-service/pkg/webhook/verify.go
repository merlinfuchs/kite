package webhook

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"
)

// Errors are only returned for settings that are wrong, like a key that isn't
// one. Anything wrong with the request, which whoever sends it controls, only
// makes its signature invalid.

const (
	AlgorithmHMACSHA256 = "hmac_sha256"
	AlgorithmHMACSHA1   = "hmac_sha1"
	AlgorithmHMACSHA512 = "hmac_sha512"
	AlgorithmEd25519    = "ed25519"
)

const (
	EncodingHex    = "hex"
	EncodingBase64 = "base64"
)

// MaxSignatureAge is how old the timestamp of a request may be for presets
// that sign one, so a request that got out can't be sent again later.
const MaxSignatureAge = 5 * time.Minute

// erlcPublicKey verifies the event webhooks of ER:LC, see
// https://apidocs.erlc.gg/event-webhooks.
const erlcPublicKey = "MCowBQYDK2VwAyEAjSICb9pp0kHizGQtdG8ySWsDChfGqi+gyFCttigBNOA="

// VerifySignature reports whether signature is the signature of message under
// key with algorithm. HMAC keys are text, Ed25519 keys are public keys as hex,
// base64 or PEM.
func VerifySignature(algorithm string, key string, message string, signature string, encoding string) (bool, error) {
	var sig []byte
	var err error
	switch encoding {
	case "", EncodingHex:
		sig, err = hex.DecodeString(strings.TrimSpace(signature))
	case EncodingBase64:
		sig, err = base64.StdEncoding.DecodeString(strings.TrimSpace(signature))
	default:
		return false, fmt.Errorf("unknown signature encoding %q", encoding)
	}
	if err != nil {
		return false, nil
	}

	if algorithm == AlgorithmEd25519 {
		publicKey, err := parseEd25519PublicKey(key)
		if err != nil {
			return false, err
		}
		return verifyEd25519(publicKey, []byte(message), sig), nil
	}

	h, err := hmacHash(algorithm)
	if err != nil {
		return false, err
	}
	if key == "" {
		return false, errors.New("the key is empty")
	}
	return hmac.Equal(sign(h, []byte(key), []byte(message)), sig), nil
}

// VerifyPreset verifies the signature of event the way the service named by
// preset signs it.
func VerifyPreset(preset string, key string, event *Event, now time.Time) (bool, error) {
	header := func(name string) string {
		return strings.TrimSpace(event.Headers[name])
	}

	switch preset {
	case "erlc", "discord":
		if preset == "erlc" {
			key = erlcPublicKey
		}
		publicKey, err := parseEd25519PublicKey(key)
		if err != nil {
			return false, err
		}
		sig, err := hex.DecodeString(header("x-signature-ed25519"))
		if err != nil {
			return false, nil
		}
		timestamp := header("x-signature-timestamp")
		if !recent(timestamp, now) {
			return false, nil
		}
		return verifyEd25519(publicKey, []byte(timestamp+event.Body), sig), nil

	case "github":
		if key == "" {
			return false, errors.New("the key is empty")
		}
		sig, ok := strings.CutPrefix(header("x-hub-signature-256"), "sha256=")
		if !ok {
			return false, nil
		}
		return hmacHexEqual(sha256.New, key, event.Body, sig), nil

	case "slack":
		if key == "" {
			return false, errors.New("the key is empty")
		}
		timestamp := header("x-slack-request-timestamp")
		sig, ok := strings.CutPrefix(header("x-slack-signature"), "v0=")
		if !ok || !recent(timestamp, now) {
			return false, nil
		}
		return hmacHexEqual(sha256.New, key, "v0:"+timestamp+":"+event.Body, sig), nil

	case "stripe":
		if key == "" {
			return false, errors.New("the key is empty")
		}
		// t=timestamp,v1=signature, with a v1 for each secret while one is
		// rolled.
		var timestamp string
		var sigs []string
		for _, part := range strings.Split(header("stripe-signature"), ",") {
			name, value, _ := strings.Cut(strings.TrimSpace(part), "=")
			switch name {
			case "t":
				timestamp = value
			case "v1":
				sigs = append(sigs, value)
			}
		}
		if !recent(timestamp, now) {
			return false, nil
		}
		for _, sig := range sigs {
			if hmacHexEqual(sha256.New, key, timestamp+"."+event.Body, sig) {
				return true, nil
			}
		}
		return false, nil

	case "standard_webhooks":
		secret, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(strings.TrimSpace(key), "whsec_"))
		if err != nil || len(secret) == 0 {
			return false, errors.New("the key must be the secret starting with whsec_")
		}
		// Svix, which many services use, sends the same headers with its name.
		id, timestamp, signatures := header("webhook-id"), header("webhook-timestamp"), header("webhook-signature")
		if id == "" {
			id, timestamp, signatures = header("svix-id"), header("svix-timestamp"), header("svix-signature")
		}
		if id == "" || !recent(timestamp, now) {
			return false, nil
		}
		expected := sign(sha256.New, secret, []byte(id+"."+timestamp+"."+event.Body))
		// v1,signature separated by spaces, with one for each secret while
		// one is rolled.
		for _, part := range strings.Fields(signatures) {
			version, value, _ := strings.Cut(part, ",")
			if version != "v1" {
				continue
			}
			sig, err := base64.StdEncoding.DecodeString(value)
			if err == nil && hmac.Equal(expected, sig) {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("unknown preset %q", preset)
}

func hmacHash(algorithm string) (func() hash.Hash, error) {
	switch algorithm {
	case "", AlgorithmHMACSHA256:
		return sha256.New, nil
	case AlgorithmHMACSHA1:
		return sha1.New, nil
	case AlgorithmHMACSHA512:
		return sha512.New, nil
	}
	return nil, fmt.Errorf("unknown algorithm %q", algorithm)
}

func sign(h func() hash.Hash, key []byte, message []byte) []byte {
	mac := hmac.New(h, key)
	mac.Write(message)
	return mac.Sum(nil)
}

func hmacHexEqual(h func() hash.Hash, key string, message string, signature string) bool {
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	return hmac.Equal(sign(h, []byte(key), []byte(message)), sig)
}

func verifyEd25519(publicKey ed25519.PublicKey, message []byte, sig []byte) bool {
	// ed25519.Verify panics on keys of the wrong size, not on signatures,
	// but a signature of the wrong size is just invalid.
	if len(sig) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(publicKey, message, sig)
}

// recent reports whether timestamp, in Unix seconds, is at most
// MaxSignatureAge away from now, in either direction for clocks that differ.
func recent(timestamp string, now time.Time) bool {
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	age := now.Sub(time.Unix(seconds, 0))
	return age <= MaxSignatureAge && age >= -MaxSignatureAge
}

// parseEd25519PublicKey reads a public key as hex, like Discord shows it, as
// PEM, or as base64 of the key itself or its SPKI encoding, like ER:LC's.
func parseEd25519PublicKey(key string) (ed25519.PublicKey, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("the public key is empty")
	}

	var raw []byte
	if block, _ := pem.Decode([]byte(key)); block != nil {
		raw = block.Bytes
	} else if b, err := hex.DecodeString(key); err == nil {
		raw = b
	} else if b, err := base64.StdEncoding.DecodeString(key); err == nil {
		raw = b
	} else {
		return nil, errors.New("the public key must be hex, base64 or PEM")
	}

	if len(raw) == ed25519.PublicKeySize {
		return ed25519.PublicKey(raw), nil
	}
	parsed, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		return nil, errors.New("the public key isn't an Ed25519 public key")
	}
	publicKey, ok := parsed.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("the public key isn't an Ed25519 public key")
	}
	return publicKey, nil
}
