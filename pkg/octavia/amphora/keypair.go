package amphora

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func newKeypairSecret(instance *openstackv1beta1.Octavia) (*corev1.Secret, error) {
	labels := template.AppLabels(instance.Name, "octavia")
	name := template.Combine(instance.Name, "amphora-ssh")

	secret := template.GenericSecret(name, instance.Namespace, labels)

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
