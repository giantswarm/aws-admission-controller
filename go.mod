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
	github.com/google/go-cmp v0.5.8
	github.com/prometheus/client_golang v1.12.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
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
	github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.44.34
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.6
	github.com/containerd/imgcrypt => github.com/containerd/imgcrypt v1.1.6
	// v3.3.10 is required by spf13/viper. Can remove this replace when updated.
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.8.1
	github.com/go-ldap/ldap/v3 => github.com/go-ldap/ldap/v3 v3.4.2
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
	github.com/kataras/iris/v12 => github.com/kataras/iris/v12 v12.2.0-beta3
	github.com/labstack/echo/v4 => github.com/labstack/echo/v4 v4.7.2
	github.com/microcosm-cc/bluemonday => github.com/microcosm-cc/bluemonday v1.0.18
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.8.4
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/pkg/sftp => github.com/pkg/sftp v1.13.5
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.37.0
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.4
)
