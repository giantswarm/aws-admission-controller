package admission

type PatchOperation struct {
	Operation string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value"`
}

func PatchReplace(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Operation: "replace",
		Path:      path,
		Value:     value,
	}
}

func PatchAdd(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Operation: "add",
		Path:      path,
		Value:     value,
	}
}
