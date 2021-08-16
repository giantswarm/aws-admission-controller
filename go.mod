module github.com/giantswarm/aws-admission-controller/v3

go 1.15

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/dylanmei/iso8601 v0.1.0
	github.com/dyson/certman v0.2.1
	github.com/giantswarm/apiextensions/v3 v3.31.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/ruleengine v0.2.0
	github.com/giantswarm/to v0.3.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.1.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/prometheus/client_golang v1.11.0
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.19
	k8s.io/apiextensions-apiserver v0.18.19
	k8s.io/apimachinery v0.18.19
	k8s.io/client-go v0.18.19
	k8s.io/klog/v2 v2.1.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29 // indirect
	k8s.io/utils v0.0.0-20210709001253-0e1f9d693477 // indirect
	sigs.k8s.io/cluster-api v0.4.1
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)
