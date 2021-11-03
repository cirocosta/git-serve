package pkg

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	gossh "golang.org/x/crypto/ssh"
)

func GenSSHKeyPair() ([]byte, []byte, error) {
	var privateKeyBuf bytes.Buffer

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(&privateKeyBuf, privateKeyPEM); err != nil {
		return nil, nil, err
	}

	publicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privateKeyBuf.Bytes(), gossh.MarshalAuthorizedKey(publicKey), err
}
