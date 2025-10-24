package ssh

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"math/big"
)

const (
	// magic word openSSH keys start with
	magic = "openssh-key-v1\x00"
)

// PrivateKeyToOpenSSH returns a string OpenSSH PEM encoded private key from a PKCS#8 PEM encoded private key
func PrivateKeyToOpenSSH(pemBytes []byte, uuid string) (string, error) {
	// Decode and get the PEM Private key block.
	pemBlock, rest := pem.Decode(pemBytes)
	if pemBlock == nil {
		return "", errors.New("invalid PEM private key passed in, decoding did not find a key")
	}
	if len(rest) > 0 {
		return "", errors.New("PEM block contains more than just private key")
	}

	// Confirm we got a supported block type
	var key any
	var err error
	switch pemBlock.Type {
	// The 'PRIVATE KEY' PEM block header is specific to the PKCS#8 standard.
	case "PRIVATE KEY":
		// Get the private key from PKCS#8 encoding
		key, err = x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return "", fmt.Errorf("error during parsing from PKCS#8, invalid PEM private key passed in: %w", err)
		}
	// The 'RSA PRIVATE KEY' PEM block header is specific to the older PKCS#1 format - which is only specific to the RSA algorithm.
	case "RSA PRIVATE KEY":
		// Get the private key from PKCS#1 encoding
		key, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		if err != nil {
			return "", fmt.Errorf("error during parsing from PCKS#1, invalid PEM private key passed in: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported key type %q passed with the PEM", pemBlock.Type)
	}

	// Marshal serialized private key to OpenSSH PEM.
	openSSHPemBlock, err := marshalOpenSSHPrivateKey(key, uuid)
	if err != nil {
		return "", fmt.Errorf("marshaling to OpenSSH format failed: %w", err)
	}

	encodedOpenSSHPrivateKey := pem.EncodeToMemory(openSSHPemBlock)
	return openSSHLineBreaker(encodedOpenSSHPrivateKey), nil
}

// marshalOpenSSHPrivateKey marshals an ed25519 or rsa private key into an OpenSSH Pem Block.
// Reverse engineered from ssh.parseOpenSSHPrivateKey and inspired from patch: https://go-review.googlesource.com/c/crypto/+/218620/.
func marshalOpenSSHPrivateKey(key crypto.PrivateKey, uuid string) (*pem.Block, error) {
	var privateKey struct {
		Check1  uint32
		Check2  uint32
		KeyType string
		Rest    []byte `ssh:"rest"`
	}

	var openSSHKey struct {
		CipherName      string
		KdfName         string
		KdfOpts         string
		NumKeys         uint32
		PubKey          []byte
		PrivateKeyBlock []byte
	}

	// Generate some check-bytes deterministically using the item uuid. This should match the approach taken on core.
	var check uint32
	hasher := sha256.New()
	hasher.Write([]byte(uuid))
	hashedUUID := hasher.Sum(nil)
	check = binary.BigEndian.Uint32(hashedUUID)

	privateKey.Check1 = check
	privateKey.Check2 = check
	openSSHKey.NumKeys = 1

	switch k := key.(type) {
	case *rsa.PrivateKey:
		E := new(big.Int).SetInt64(int64(k.PublicKey.E))
		// Marshal public key:
		// E and N are in reversed order in the public and private key.
		pubKey := struct {
			KeyType string
			E       *big.Int
			N       *big.Int
		}{
			ssh.KeyAlgoRSA,
			E, k.PublicKey.N,
		}
		openSSHKey.PubKey = ssh.Marshal(pubKey)

		// Marshal private key.
		prvKey := struct {
			N       *big.Int
			E       *big.Int
			D       *big.Int
			Iqmp    *big.Int
			P       *big.Int
			Q       *big.Int
			Comment string
		}{
			k.PublicKey.N, E,
			k.D, k.Precomputed.Qinv, k.Primes[0], k.Primes[1],
			"",
		}
		privateKey.KeyType = ssh.KeyAlgoRSA
		privateKey.Rest = ssh.Marshal(prvKey)
	case ed25519.PrivateKey:
		pub := make([]byte, ed25519.PublicKeySize)
		priv := make([]byte, ed25519.PrivateKeySize)
		copy(pub, k[ed25519.PublicKeySize:])
		copy(priv, k)

		// Marshal public key.
		pubKey := struct {
			KeyType string
			Pub     []byte
		}{
			ssh.KeyAlgoED25519, pub,
		}
		openSSHKey.PubKey = ssh.Marshal(pubKey)

		// Marshal private key.
		prvKey := struct {
			Pub     []byte
			Priv    []byte
			Comment string
		}{
			pub, priv,
			"",
		}
		privateKey.KeyType = ssh.KeyAlgoED25519
		privateKey.Rest = ssh.Marshal(prvKey)
	default:
		return nil, errors.New("unsupported key type provided")
	}

	// Add padding. No encryption necessary.
	padded := generateOpenSSHPadding(ssh.Marshal(privateKey), 8)
	openSSHKey.PrivateKeyBlock = padded
	openSSHKey.CipherName = "none"
	openSSHKey.KdfName = "none"
	openSSHKey.KdfOpts = ""

	SSHWireFormatKey := ssh.Marshal(openSSHKey)
	block := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: append([]byte(magic), SSHWireFormatKey...),
	}
	return block, nil
}

func generateOpenSSHPadding(block []byte, blockSize int) []byte {
	for i, l := 0, len(block); (l+i)%blockSize != 0; i++ {
		block = append(block, byte(i+1))
	}
	return block
}

// openSSHLineBreaker makes sure the base64 encoded OpenSSH key has line breaks once every 70 ASCII chars
func openSSHLineBreaker(privateKeyBytes []byte) string {
	const lineFeedByte = byte('\n')
	const lineBreaker = 70
	var prefix = []byte("-----BEGIN OPENSSH PRIVATE KEY-----")
	var suffix = []byte("-----END OPENSSH PRIVATE KEY-----")

	// new line character was appended once every 64 bytes, so remove those
	keyWithoutLineBreaks := bytes.ReplaceAll(privateKeyBytes, []byte{lineFeedByte}, nil)

	keyWithoutPrefix := bytes.TrimPrefix(keyWithoutLineBreaks, prefix)
	keyWithoutPrefixAndSuffix := bytes.TrimSuffix(keyWithoutPrefix, suffix)

	// append the new line character once every 70 bytes instead according to OpenSSH standard
	openSSHBytes := bytes.NewBuffer(prefix)

	for i, b := range keyWithoutPrefixAndSuffix {
		if i%lineBreaker == 0 {
			openSSHBytes.WriteByte(lineFeedByte)
		}
		openSSHBytes.WriteByte(b)
	}

	openSSHBytes.WriteByte(lineFeedByte)
	openSSHBytes.Write(suffix)
	openSSHBytes.WriteByte(lineFeedByte)

	return openSSHBytes.String()
}
