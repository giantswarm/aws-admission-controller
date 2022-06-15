module github.com/giantswarm/aws-admission-controller/v4

go 1.15

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/dylanmei/iso8601 v0.1.0
	github.com/dyson/certman v0.2.1
	github.com/giantswarm/apiextensions/v6 v6.0.0
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/k8sclient/v7 v7.0.1
	github.com/giantswarm/k8smetadata v0.10.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/release-operator/v3 v3.2.0
	github.com/giantswarm/ruleengine v0.2.0
	github.com/giantswarm/to v0.4.0
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/google/go-cmp v0.5.7
	github.com/kr/pretty v0.3.0 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/prometheus/client_golang v1.12.0
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v3 v3.0.0 // indirect
	k8s.io/api v0.22.5
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v0.22.5
	k8s.io/component-base v0.22.5 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	sigs.k8s.io/cluster-api v1.0.4
	sigs.k8s.io/controller-runtime v0.10.3
)

replace (
	// v0.8.7 requires kubernetes 1.13 that triggers nancy alerts.
	github.com/Microsoft/hcsshim v0.8.7 => github.com/Microsoft/hcsshim v0.8.10
	// v3.3.10 is required by spf13/viper. Can remove this replace when updated.
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.4
)
