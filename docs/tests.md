## How to write tests

In general when the admission controller receives an admission request it will be send to the right admission handler depending on the [webhook configuration](../helm/aws-admission-controller/templates/webhook.yaml).

Each webhook endpoint is created in `pkg`, e.g. `g8scontrolplane`.

```
└─ pkg
    └─ example-webhook
        ├─ admit_example.go
        ├─ admit_example_test.go
        └─ testcases
            ├─ test1.yaml
            └─ test2.yaml
```

Example for the go test file:

```go
package example

import (
	"testing"

	"github.com/giantswarm/aws-admission-controller/pkg/admission"
	"github.com/giantswarm/aws-admission-controller/pkg/testrunner"
	"github.com/ghodss/yaml"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func newExampleAdmitter(config []byte) (admission.Admitter, error) {
	conf := &AdmitterConfig{}

	err := yaml.Unmarshal(config, &conf)
	if err != nil {
		return nil, err
	}

	return NewAdmitter(conf)
}

func TestPodAdmission(t *testing.T) {
	runner := &testrunner.Runner{
		CreateAdmitter: newExampleAdmitter,
		Resource:       exampleResource,
		NewElement:     func() runtime.Object { return &v1.Example{} },
	}
	runner.RunTestcases(t)
}
```

This will run all tests inside the `testcases` directory.

Example for a testcase:

```yaml
config: nil # Config can be set in case NewAdmitter uses admitter config
object:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: example
  ...
expected:
  apiVersion: example.com/v1
  kind: Example
  metadata:
    name: example
    injected: new
  ...
```
