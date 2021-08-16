package template

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	AppliedHashAnnotation = "openstack.k8s.ianunruh.com/applied-hash"
)

func MatchesAppliedHash(obj metav1.Object, expected string) bool {
	return AppliedHash(obj) == expected
}

func AppliedHash(obj metav1.Object) string {
	ann := obj.GetAnnotations()
	return ann[AppliedHashAnnotation]
}

func SetAppliedHash(obj metav1.Object, hash string) {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
		obj.SetAnnotations(ann)
	}
	ann[AppliedHashAnnotation] = hash
}

func SetAppliedHashUnstructured(instance *unstructured.Unstructured, hash string) {
	unstructured.SetNestedField(instance.UnstructuredContent(), hash, "metadata", "annotations", AppliedHashAnnotation)
}

// ObjectHash creates a deep object hash and return it as a safe encoded string
func ObjectHash(i interface{}) (string, error) {
	// Convert the hashSource to a byte slice so that it can be hashed
	hashBytes, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("unable to convert to JSON: %v", err)
	}
	hash := sha256.Sum256(hashBytes)
	return rand.SafeEncodeString(fmt.Sprint(hash)), nil
}
