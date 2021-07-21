package unittest

import (
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	fakeg8s "github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned/fake"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v5/pkg/k8scrdclient"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck // v0.6.4 has a deprecation on pkg/client/fake that was removed in later versions
)

type fakeK8sClient struct {
	ctrlClient client.Client
	k8sClient  *fakek8s.Clientset
	g8sclient  *fakeg8s.Clientset
}

func FakeK8sClient() k8sclient.Interface {
	var err error

	var k8sClient k8sclient.Interface
	{
		scheme := runtime.NewScheme()
		err = v1alpha2.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = infrastructurev1alpha2.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = releasev1alpha1.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = securityv1alpha1.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		_ = fakek8s.AddToScheme(scheme)
		client := fakek8s.NewSimpleClientset()
		g8sclient := fakeg8s.NewSimpleClientset()

		k8sClient = &fakeK8sClient{
			ctrlClient: fake.NewFakeClientWithScheme(scheme),
			k8sClient:  client,
			g8sclient:  g8sclient,
		}
	}

	return k8sClient
}

func (f *fakeK8sClient) CRDClient() k8scrdclient.Interface {
	return nil
}

func (f *fakeK8sClient) CtrlClient() client.Client {
	return f.ctrlClient
}

func (f *fakeK8sClient) DynClient() dynamic.Interface {
	return nil
}

func (f *fakeK8sClient) ExtClient() apiextensionsclient.Interface {
	return nil
}

func (f *fakeK8sClient) G8sClient() versioned.Interface {
	return f.g8sclient
}

func (f *fakeK8sClient) K8sClient() kubernetes.Interface {
	return f.k8sClient
}

func (f *fakeK8sClient) RESTClient() rest.Interface {
	return nil
}

func (f *fakeK8sClient) RESTConfig() *rest.Config {
	return nil
}

func (f *fakeK8sClient) Scheme() *runtime.Scheme {
	return nil
}
