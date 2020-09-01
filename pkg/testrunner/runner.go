package testrunner

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-admission-controller/pkg/admission"
)

type Runner struct {
	CreateAdmitter func(config []byte) (admission.Admitter, error)
	Resource       metav1.GroupVersionResource
	NewElement     func() runtime.Object
}

type testcaseDefinition struct {
	Config    interface{}       `json:"config"`
	Object    interface{}       `json:"object"`
	Expected  interface{}       `json:"expected"`
	Operation v1beta1.Operation `json:"operation"`
	Error     string            `json:"error"`
	Namespace string            `json:"namespace"`
}

func (runner *Runner) parseObject(t *testing.T, from string) runtime.Object {
	object := runner.NewElement()
	_, _, err := admission.Deserializer.Decode([]byte(from), nil, object)
	require.NoError(t, err)
	return object
}

func (runner *Runner) objectToRequest(t *testing.T, object runtime.Object, namespace string, operation v1beta1.Operation) *v1beta1.AdmissionRequest {
	serialized, err := json.Marshal(object)
	require.NoError(t, err)
	gvk := object.GetObjectKind().GroupVersionKind()

	accessor := meta.NewAccessor()

	name, err := accessor.Name(object)
	require.NoError(t, err)

	return &v1beta1.AdmissionRequest{
		UID:       "example",
		Kind:      metav1.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind},
		Resource:  runner.Resource,
		Name:      name,
		Namespace: namespace,
		Operation: operation,
		Object: runtime.RawExtension{
			Raw: serialized,
		},
	}
}

func (runner *Runner) applyPatch(t *testing.T, object runtime.Object, patch []admission.PatchOperation) runtime.Object {
	serializedObject, err := json.Marshal(object)
	require.NoError(t, err)

	serializedPatch, err := json.Marshal(patch)
	require.NoError(t, err)

	jPatch, err := jsonpatch.DecodePatch(serializedPatch)
	require.NoError(t, err)

	updated, err := jPatch.Apply([]byte(serializedObject))
	require.NoError(t, err)

	return runner.parseObject(t, string(updated))
}

func asYaml(t *testing.T, obj interface{}) string {
	serialized, err := json.Marshal(obj)
	require.NoError(t, err)

	simpleYaml, err := yaml.JSONToYAML(serialized)
	require.NoError(t, err)

	return string(simpleYaml)
}

func (runner *Runner) runTestcase(t *testing.T, filename string) {
	data, err := ioutil.ReadFile(filename)
	require.NoError(t, err)

	var testcase testcaseDefinition

	err = yaml.Unmarshal(data, &testcase)
	require.NoError(t, err)

	objectYaml, err := yaml.Marshal(testcase.Object)
	require.NoError(t, err)

	namespace := testcase.Namespace
	if namespace == "" {
		namespace = "default"
	}

	sourceObject := runner.parseObject(t, string(objectYaml))
	request := runner.objectToRequest(t, sourceObject, namespace, testcase.Operation)

	config, err := yaml.Marshal(testcase.Config)
	require.NoError(t, err)

	admitter, err := runner.CreateAdmitter(config)
	require.NoError(t, err)

	patch, err := admitter.Admit(request)
	if testcase.Error != "" {
		require.Error(t, err)
		require.Equal(t, string(testcase.Error), err.Error())
	} else {
		require.NoError(t, err)
		updatedObject := runner.applyPatch(t, sourceObject, patch)
		require.Equal(t, asYaml(t, testcase.Expected), asYaml(t, updatedObject))

	}
}

func (runner *Runner) RunTestcases(t *testing.T) {
	files, err := ioutil.ReadDir("testcases")
	require.NoError(t, err)

	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			runner.runTestcase(t, path.Join("testcases", file.Name()))
		})
	}
}
