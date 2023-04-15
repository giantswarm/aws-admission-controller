module github.com/giantswarm/aws-admission-controller/v4

go 1.15

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/dylanmei/iso8601 v0.1.0
	github.com/dyson/certman v0.3.0
	github.com/giantswarm/apiextensions/v6 v6.0.0
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/k8sclient/v7 v7.0.1
	github.com/giantswarm/k8smetadata v0.14.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/micrologger v1.0.0
	github.com/giantswarm/release-operator/v3 v3.2.0
	github.com/giantswarm/ruleengine v0.2.0
	github.com/giantswarm/to v0.4.0
	github.com/google/go-cmp v0.5.9
	github.com/kr/pretty v0.3.1 // indirect
	github.com/prometheus/client_golang v1.15.0
	github.com/stretchr/testify v1.8.1 // indirect
	golang.org/x/net v0.8.0 // indirect
	google.golang.org/protobuf v1.29.1 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.3.2
	k8s.io/api v0.24.0
	k8s.io/apiextensions-apiserver v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	sigs.k8s.io/cluster-api v1.0.4
	sigs.k8s.io/controller-runtime v0.12.1
)

replace (
	// v0.8.7 requires kubernetes 1.13 that triggers nancy alerts.
	github.com/Microsoft/hcsshim v0.8.7 => github.com/Microsoft/hcsshim v0.8.10
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.44.244
	github.com/containerd/containerd => github.com/containerd/containerd v1.7.0
	github.com/containerd/imgcrypt => github.com/containerd/imgcrypt v1.1.7
	// v3.3.10 is required by spf13/viper. Can remove this replace when updated.
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/getsentry/sentry-go => github.com/getsentry/sentry-go v0.20.0
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.0
	github.com/go-ldap/ldap/v3 => github.com/go-ldap/ldap/v3 v3.4.4
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
	github.com/kataras/iris/v12 => github.com/kataras/iris/v12 v12.2.0
	github.com/labstack/echo/v4 => github.com/labstack/echo/v4 v4.10.2
	github.com/microcosm-cc/bluemonday => github.com/microcosm-cc/bluemonday v1.0.23
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.9.15
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.25.0
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/pkg/sftp => github.com/pkg/sftp v1.13.5
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.45.0
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.4
)
