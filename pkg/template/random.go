package template

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/sethvargo/go-password/password"
)

var generator *password.Generator

func init() {
	var err error
	generator, err = password.NewGenerator(nil)
	if err != nil {
		panic(err)
	}
}

func MustGenerateFernetKey() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func MustGeneratePassword() string {
	return generator.MustGenerate(32, 10, 0, false, false)
}
