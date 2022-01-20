module github.com/giantswarm/aws-admission-controller/v3

go 1.15

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/dylanmei/iso8601 v0.1.0
	github.com/dyson/certman v0.2.1
	github.com/giantswarm/apiextensions/v3 v3.39.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/k8smetadata v0.7.1
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/ruleengine v0.2.0
	github.com/giantswarm/to v0.4.0
	github.com/google/go-cmp v0.5.7
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/prometheus/client_golang v1.12.0
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.20.15
	k8s.io/apiextensions-apiserver v0.20.15
	k8s.io/apimachinery v0.20.15
	k8s.io/client-go v0.20.15
	k8s.io/utils v0.0.0-20210709001253-0e1f9d693477 // indirect
	sigs.k8s.io/cluster-api v0.99.99
	sigs.k8s.io/controller-runtime v0.6.5
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)
