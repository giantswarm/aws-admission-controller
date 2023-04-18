package key

import "sigs.k8s.io/controller-runtime/pkg/client"

const (
	fluxKustomizationNameLabel      = "kustomize.toolkit.fluxcd.io/name"
	fluxKustomizationNamespaceLabel = "kustomize.toolkit.fluxcd.io/namespace"
)

func FluxKustomizationObjectKey(getter LabelsGetter) *client.ObjectKey {
	fluxName := ""
	fluxNamespace := ""
	for k, v := range getter.GetLabels() {
		if k == fluxKustomizationNameLabel {
			fluxName = v
		} else if k == fluxKustomizationNamespaceLabel {
			fluxNamespace = v
		}
	}

	if fluxName != "" && fluxNamespace != "" {
		return &client.ObjectKey{
			Name:      fluxName,
			Namespace: fluxNamespace,
		}
	}

	return nil
}
