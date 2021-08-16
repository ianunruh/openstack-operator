package template

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/go-logr/logr"
	"gopkg.in/ini.v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func init() {
	// pretty much all OpenStack configs need the explicit default section
	// header, and this package only has one way of doing that
	ini.DefaultHeader = true
}

func ReadFile(app, filename string) (string, error) {
	basePath := os.Getenv("OPERATOR_TEMPLATES")
	if basePath == "" {
		basePath = "templates"
	}

	path := filepath.Join(basePath, app, filename)

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func MustReadFile(app, filename string) string {
	out, err := ReadFile(app, filename)
	if err != nil {
		panic(err)
	}
	return out
}

func RenderFile(app, filename string, values interface{}) (string, error) {
	file, err := ReadFile(app, filename)
	if err != nil {
		return "", err
	}

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

func DecodeManifest(encoded string) (*unstructured.Unstructured, error) {
	resource := &unstructured.Unstructured{}
	_, _, err := decUnstructured.Decode([]byte(encoded), nil, resource)
	if err != nil {
		return nil, err
	}
	return resource, nil
}

func MustDecodeManifest(encoded string) *unstructured.Unstructured {
	resource, err := DecodeManifest(encoded)
	if err != nil {
		panic(err)
	}
	return resource
}

func EnsureResources(ctx context.Context, c client.Client, resources []*unstructured.Unstructured, log logr.Logger) error {
	for _, instance := range resources {
		if err := EnsureResource(ctx, c, instance, log); err != nil {
			return err
		}
	}

	return nil
}

func EnsureResource(ctx context.Context, c client.Client, instance *unstructured.Unstructured, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	fields := []interface{}{
		"Name", intended.GetName(),
		"Namespace", intended.GetNamespace(),
		"Kind", intended.GetObjectKind().GroupVersionKind(),
	}

	if err := c.Get(context.TODO(), client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHashUnstructured(instance, hash)

		log.Info("Creating Resource", fields...)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		intendedSpec, _, err := unstructured.NestedMap(intended.UnstructuredContent(), "spec")
		if err != nil {
			return err
		}

		unstructured.SetNestedMap(instance.UnstructuredContent(), intendedSpec, "spec")

		SetAppliedHashUnstructured(instance, hash)

		log.Info("Updating Resource", fields...)
		return c.Update(ctx, instance)
	}

	return nil
}

func MustParseINI(encoded string) *ini.File {
	file, err := ini.Load([]byte(encoded))
	if err != nil {
		panic(err)
	}
	return file
}

func MustLoadINITemplate(app, filename string, values interface{}) *ini.File {
	return MustParseINI(MustRenderFile(app, filename, values))
}

func MustLoadINI(app, filename string) *ini.File {
	return MustParseINI(MustReadFile(app, filename))
}

func MustOutputINI(file *ini.File) *bytes.Buffer {
	cfgOut := &bytes.Buffer{}
	if _, err := file.WriteTo(cfgOut); err != nil {
		panic(err)
	}
	return cfgOut
}
