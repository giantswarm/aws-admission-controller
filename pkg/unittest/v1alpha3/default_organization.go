package unittest

import (
	securityv1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DefaultOrganization() *securityv1alpha1.Organization {
	organization := &securityv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-organization",
		},
		Spec: securityv1alpha1.OrganizationSpec{},
	}
	return organization
}
