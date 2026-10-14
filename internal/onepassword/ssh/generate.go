package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// GeneratedKey holds the outputs of generating an SSH key pair. The private
// key is kept in PKCS#8 form because that is the encoding 1Password stores
// for SSH key items; the OpenSSH rendering is derived on read.
type GeneratedKey struct {
	PrivateKeyPKCS8 string // PEM-encoded PKCS#8 private key
	PublicKey       string // authorized_keys format, e.g. "ssh-ed25519 AAAA..."
	Fingerprint     string // e.g. "SHA256:abcdef..."
	KeyType         string // "Ed25519" or "RSA, 2048-bit"
}

// GenerateKeyPair generates an SSH key pair of the given type ("ed25519" or
// "rsa"). bits is only used for RSA keys.
func GenerateKeyPair(keyType string, bits int) (*GeneratedKey, error) {
	switch strings.ToLower(keyType) {
	case "ed25519":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generating ed25519 key: %w", err)
		}
		return newGeneratedKey(priv, pub)
	case "rsa":
		if bits < 2048 {
			return nil, fmt.Errorf("rsa key size must be at least 2048 bits, got %d", bits)
		}
		priv, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, fmt.Errorf("generating rsa key: %w", err)
		}
		return newGeneratedKey(priv, &priv.PublicKey)
	default:
		return nil, fmt.Errorf("unsupported key type %q: must be \"ed25519\" or \"rsa\"", keyType)
	}
}

func newGeneratedKey(priv any, pub any) (*GeneratedKey, error) {
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshaling private key to PKCS#8: %w", err)
	}
	pkcs8PEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("deriving SSH public key: %w", err)
	}

	return &GeneratedKey{
		PrivateKeyPKCS8: pkcs8PEM,
		PublicKey:       strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub))),
		Fingerprint:     ssh.FingerprintSHA256(sshPub),
		KeyType:         keyTypeLabel(sshPub),
	}, nil
}

// PublicKeyFromPrivateKey derives the authorized_keys-format public key
// from a PEM private key in PKCS#8, PKCS#1, or OpenSSH form.
func PublicKeyFromPrivateKey(privateKeyPEM string) (string, error) {
	key, err := parsePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return "", err
	}

	var pub any
	switch k := key.(type) {
	case *rsa.PrivateKey:
		pub = &k.PublicKey
	case ed25519.PrivateKey:
		pub = k.Public()
	case *ed25519.PrivateKey:
		pub = k.Public()
	default:
		return "", fmt.Errorf("unsupported private key type %T", key)
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("deriving SSH public key: %w", err)
	}

	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub))), nil
}

// parsePrivateKeyPEM decodes a PEM private key in PKCS#8, PKCS#1, or
// OpenSSH form.
func parsePrivateKeyPEM(privateKeyPEM string) (any, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM private key passed in, decoding did not find a key")
	}

	switch block.Type {
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "OPENSSH PRIVATE KEY":
		parsed, err := ssh.ParseRawPrivateKey([]byte(privateKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("error parsing OpenSSH private key: %w", err)
		}
		if k, ok := parsed.(*ed25519.PrivateKey); ok {
			return *k, nil
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("unsupported key type %q passed with the PEM", block.Type)
	}
}

// OpenSSHToPKCS8 converts an OpenSSH-format PEM private key back to PKCS#8
// PEM, the encoding 1Password stores. Used to preserve a generated key
// across resource updates.
func OpenSSHToPKCS8(openSSHPEM string) (string, error) {
	block, _ := pem.Decode([]byte(openSSHPEM))
	if block == nil {
		return "", fmt.Errorf("invalid PEM private key passed in, decoding did not find a key")
	}

	if block.Type == "PRIVATE KEY" {
		// Already PKCS#8.
		return openSSHPEM, nil
	}

	key, err := parsePrivateKeyPEM(openSSHPEM)
	if err != nil {
		return "", err
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("marshaling private key to PKCS#8: %w", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), nil
}

// PublicKeyFingerprint returns the SHA256 fingerprint of an authorized_keys
// format public key, e.g. "SHA256:abcdef...".
func PublicKeyFingerprint(authorizedKey string) (string, error) {
	pub, err := parseAuthorizedKey(authorizedKey)
	if err != nil {
		return "", err
	}
	return ssh.FingerprintSHA256(pub), nil
}

// KeyTypeFromPublicKey returns the 1Password key type label for an
// authorized_keys format public key: "Ed25519" or "RSA, <bits>-bit".
func KeyTypeFromPublicKey(authorizedKey string) (string, error) {
	pub, err := parseAuthorizedKey(authorizedKey)
	if err != nil {
		return "", err
	}
	return keyTypeLabel(pub), nil
}

func parseAuthorizedKey(authorizedKey string) (ssh.PublicKey, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(authorizedKey))
	if err != nil {
		return nil, fmt.Errorf("parsing public key %q: %w", authorizedKey, err)
	}
	return pub, nil
}

func keyTypeLabel(pub ssh.PublicKey) string {
	switch pub.Type() {
	case ssh.KeyAlgoED25519:
		return "Ed25519"
	case ssh.KeyAlgoRSA:
		if cryptoPub, ok := pub.(ssh.CryptoPublicKey); ok {
			if rsaPub, ok := cryptoPub.CryptoPublicKey().(*rsa.PublicKey); ok {
				return fmt.Sprintf("RSA, %d-bit", rsaPub.N.BitLen())
			}
		}
		return "RSA"
	default:
		return pub.Type()
	}
}
