package template

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

func RenderFile(app, filename string, values interface{}) (string, error) {
	basePath := os.Getenv("OPERATOR_TEMPLATES")
	if basePath == "" {
		basePath = "templates"
	}

	path := filepath.Join(basePath, app, filename)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	file := string(b)

	var buff bytes.Buffer
	tmpl, err := template.New("tmp").Parse(file)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buff, values)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

func MustRenderFile(app, filename string, values interface{}) string {
	out, err := RenderFile(app, filename, values)
	if err != nil {
		panic(err)
	}
	return out
}
