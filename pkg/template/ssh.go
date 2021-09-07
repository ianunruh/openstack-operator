package template

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
)

func SSHKeypairSecret(name, namespace string, labels map[string]string) (*corev1.Secret, error) {
	secret := GenericSecret(name, namespace, labels)

	privateKey, authorizedKey, err := newSSHKeypair()
	if err != nil {
		return nil, err
	}

	secret.Data = map[string][]byte{
		"id_rsa":     privateKey,
		"id_rsa.pub": authorizedKey,
	}

	return secret, nil
}

func newSSHKeypair() (privateKey []byte, authorizedKey []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	privateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	publicKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return
	}
	authorizedKey = ssh.MarshalAuthorizedKey(publicKey)

	return
}
