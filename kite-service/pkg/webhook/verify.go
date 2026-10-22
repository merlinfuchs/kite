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
	"io"
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

const (
	PresetCustom           = "custom"
	PresetERLC             = "erlc"
	PresetGitHub           = "github"
	PresetStripe           = "stripe"
	PresetSlack            = "slack"
	PresetDiscord          = "discord"
	PresetStandardWebhooks = "standard_webhooks"
)

// MaxSignatureAge is how old the timestamp of a request may be for presets
// that sign one, so a request that got out can't be sent again later.
const MaxSignatureAge = 5 * time.Minute

// erlcPublicKey verifies the event webhooks of ER:LC, see
// https://apidocs.erlc.gg/event-webhooks.
var erlcPublicKey = func() ed25519.PublicKey {
	key, err := parseEd25519PublicKey("MCowBQYDK2VwAyEAjSICb9pp0kHizGQtdG8ySWsDChfGqi+gyFCttigBNOA=")
	if err != nil {
		panic(err)
	}
	return key
}()

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

	if algorithm == AlgorithmEd25519 {
		publicKey, keyErr := parseEd25519PublicKey(key)
		if keyErr != nil {
			return false, keyErr
		}
		return err == nil && verifyEd25519(publicKey, sig, message), nil
	}

	h, hashErr := hmacHash(algorithm)
	if hashErr != nil {
		return false, hashErr
	}
	if key == "" {
		return false, errors.New("the key is empty")
	}
	return err == nil && hmac.Equal(sign(h, []byte(key), message), sig), nil
}

// VerifyPreset verifies the signature of event the way the service named by
// preset signs it.
func VerifyPreset(preset string, key string, event *Event) (bool, error) {
	now := event.ReceivedAt
	if key == "" && preset != PresetERLC {
		return false, errors.New("the key is empty")
	}
	header := func(name string) string {
		return strings.TrimSpace(event.Headers[name])
	}
	body := event.Body

	switch preset {
	case PresetERLC, PresetDiscord:
		publicKey := erlcPublicKey
		if preset == PresetDiscord {
			var err error
			if publicKey, err = parseEd25519PublicKey(key); err != nil {
				return false, err
			}
		}
		sig, err := hex.DecodeString(header("x-signature-ed25519"))
		timestamp := header("x-signature-timestamp")
		return err == nil && recent(timestamp, now) && verifyEd25519(publicKey, sig, timestamp, body), nil

	case PresetGitHub:
		sig, ok := strings.CutPrefix(header("x-hub-signature-256"), "sha256=")
		return ok && hexEqual(sign(sha256.New, []byte(key), body), sig), nil

	case PresetSlack:
		timestamp := header("x-slack-request-timestamp")
		sig, ok := strings.CutPrefix(header("x-slack-signature"), "v0=")
		return ok && recent(timestamp, now) &&
			hexEqual(sign(sha256.New, []byte(key), "v0:", timestamp, ":", body), sig), nil

	case PresetStripe:
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
		expected := sign(sha256.New, []byte(key), timestamp, ".", body)
		for _, sig := range sigs {
			if hexEqual(expected, sig) {
				return true, nil
			}
		}
		return false, nil

	case PresetStandardWebhooks:
		// Svix, which many services use, sends the same headers with its name.
		id, timestamp, signatures := header("webhook-id"), header("webhook-timestamp"), header("webhook-signature")
		if id == "" {
			id, timestamp, signatures = header("svix-id"), header("svix-timestamp"), header("svix-signature")
		}

		// A secret for HMAC signatures (v1), or a public key for Ed25519
		// signatures (v1a).
		key = strings.TrimSpace(key)
		var version string
		var verify func(sig []byte) bool
		if publicKey, ok := strings.CutPrefix(key, "whpk_"); ok {
			raw, err := base64.StdEncoding.DecodeString(publicKey)
			if err != nil || len(raw) != ed25519.PublicKeySize {
				return false, errors.New("the public key starting with whpk_ isn't one")
			}
			version = "v1a"
			verify = func(sig []byte) bool {
				return verifyEd25519(raw, sig, id, ".", timestamp, ".", body)
			}
		} else {
			secret, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(key, "whsec_"))
			if err != nil || len(secret) == 0 {
				return false, errors.New("the key must be the secret starting with whsec_ or the public key starting with whpk_")
			}
			expected := sign(sha256.New, secret, id, ".", timestamp, ".", body)
			version = "v1"
			verify = func(sig []byte) bool { return hmac.Equal(expected, sig) }
		}

		if id == "" || !recent(timestamp, now) {
			return false, nil
		}
		// version,signature separated by spaces, with one for each key while
		// one is rolled.
		for _, part := range strings.Fields(signatures) {
			v, value, _ := strings.Cut(part, ",")
			if v != version {
				continue
			}
			sig, err := base64.StdEncoding.DecodeString(value)
			if err == nil && verify(sig) {
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

// sign returns the HMAC of the parts written one after the other, without
// copying the body into a message first.
func sign(h func() hash.Hash, key []byte, parts ...string) []byte {
	mac := hmac.New(h, key)
	for _, part := range parts {
		io.WriteString(mac, part)
	}
	return mac.Sum(nil)
}

func hexEqual(expected []byte, signature string) bool {
	sig, err := hex.DecodeString(signature)
	return err == nil && hmac.Equal(expected, sig)
}

func verifyEd25519(publicKey ed25519.PublicKey, sig []byte, parts ...string) bool {
	// A signature of the wrong size is just invalid.
	if len(sig) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(publicKey, []byte(strings.Join(parts, "")), sig)
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
	if publicKey, ok := parsed.(ed25519.PublicKey); err == nil && ok {
		return publicKey, nil
	}
	return nil, errors.New("the public key isn't an Ed25519 public key")
}
