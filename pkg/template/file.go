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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

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

		SetAppliedHash(instance, hash)

		log.Info("Creating Resource", fields...)
		return c.Create(ctx, instance)
	} else if !MatchesAppliedHash(instance, hash) {
		intendedSpec, _, err := unstructured.NestedMap(intended.UnstructuredContent(), "spec")
		if err != nil {
			return err
		}

		unstructured.SetNestedMap(instance.UnstructuredContent(), intendedSpec, "spec")
		SetAppliedHash(instance, hash)

		log.Info("Updating Resource", fields...)
		return c.Update(ctx, instance)
	}

	return nil
}
