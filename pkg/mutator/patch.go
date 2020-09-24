package mutator

// PatchOperation specifies one JSONPatch operation.
// See [RFC6902](https://tools.ietf.org/html/rfc6902) for details.
type PatchOperation struct {
	Operation string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value"`
}

// PatchReplace creates a patch operation of type "replace".
func PatchReplace(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Operation: "replace",
		Path:      path,
		Value:     value,
	}
}

// PatchAdd creates a patch operation of type "add".
//
// The "add" operation performs one of the following functions,
// depending upon what the target location references:
//
//  -  If the target location specifies an array index, a new value is
//     inserted into the array at the specified index.
//  -  If the target location specifies an object member that does not
//     already exist, a new member is added to the object.
//  -  If the target location specifies an object member that does exist,
//     that member's value is replaced.
//
func PatchAdd(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Operation: "add",
		Path:      path,
		Value:     value,
	}
}
