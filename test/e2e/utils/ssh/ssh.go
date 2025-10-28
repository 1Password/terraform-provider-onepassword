package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// ValidateSSHKeys confirms that the public and private SSH keys are both valid.
func ValidateSSHKeys(publicKey string, privateKey string) error {

	if publicKey == "" {
		return fmt.Errorf("ssh_public_key attribute is not set")
	}

	if privateKey == "" {
		return fmt.Errorf("ssh_private_key attribute is not set")
	}

	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return fmt.Errorf("invalid ssh public key: %s", err)
	}

	_, err = ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return fmt.Errorf("invalid ssh private key: %s", err)
	}

	return nil
}
