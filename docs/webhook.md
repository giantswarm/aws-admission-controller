
## Add a new webhook

Make sure you update the [webhook configuration](../helm/aws-admission-controller/templates/webhook.yaml) to add the object which needs to be mutated or validated.

Initiate it in `main.go` and use a optional configuration.

Example:

```go
myAdmitter , err := myadmitter.NewAdmitter(&cfg.MyConfig)
	if err != nil {
		log.Fatalf("Unable to create Pod admitter: %v", err)
	}
    ...
	handler.Handle("/myendpoint", admission.Handler(myAdmitter))
```

The URL path has to match with service path in the [webhook configuration](../helm/aws-admission-controller/templates/webhook.yaml).

To satisfy the `Admitter` interface you need to add an `Admit` method.

Example:

```go
func (admitter *Admitter) Admit(request *v1beta1.AdmissionRequest) ([]admission.PatchOperation, error) {
	if request.Resource != exampleResource {
		log.Errorf("invalid resource: %s (expected %s)", request.Resource, exampleResource)
		return nil, admission.InternalError
	}

	example := examplev1.Example{}
	if _, _, err := admission.Deserializer.Decode(request.Object.Raw, nil, &job); err != nil {
		log.Errorf("unable to parse example: %v", err)
		return nil, admission.InternalError
	}

	var result []admission.PatchOperation
	if job.Spec.TTLSecondsAfterFinished == nil {
		result = append(result, admission.PatchAdd("/spec/something", admitter.DefaultSomething)
	}
	return result, nil
}
```

It's important to know `PatchOperation` only support `PatchAdd` or `PatchReplace`, see [patch.go](../aws-admission-controller/pkg/admission/patch.go).
