module github.com/giantswarm/aws-admission-controller/v4

go 1.21

toolchain go1.22.3

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/dylanmei/iso8601 v0.1.0
	github.com/dyson/certman v0.2.1
	github.com/fluxcd/kustomize-controller/api v0.32.0
	github.com/giantswarm/apiextensions/v6 v6.6.0
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/k8sclient/v7 v7.2.0
	github.com/giantswarm/k8smetadata v0.24.0
	github.com/giantswarm/microerror v0.4.1
	github.com/giantswarm/micrologger v1.1.1
	github.com/giantswarm/organization-operator v1.6.3
	github.com/giantswarm/release-operator/v4 v4.2.0
	github.com/giantswarm/ruleengine v0.2.0
	github.com/giantswarm/to v0.4.0
	github.com/google/go-cmp v0.6.0
	github.com/prometheus/client_golang v1.19.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.28.4
	k8s.io/apiextensions-apiserver v0.28.4
	k8s.io/apimachinery v0.28.4
	k8s.io/client-go v0.28.4
	sigs.k8s.io/cluster-api v1.1.4
	sigs.k8s.io/controller-runtime v0.16.3
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fluxcd/pkg/apis/kustomize v0.7.0 // indirect
	github.com/fluxcd/pkg/apis/meta v0.18.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/oauth2 v0.16.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/term v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/component-base v0.28.4 // indirect
	k8s.io/klog/v2 v2.110.1 // indirect
	k8s.io/kube-openapi v0.0.0-20231129212854-f0671cc7e66a // indirect
	k8s.io/utils v0.0.0-20231127182322-b307cd553661 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
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
	github.com/getsentry/sentry-go => github.com/getsentry/sentry-go v0.13.0
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.8.1
	github.com/go-ldap/ldap/v3 => github.com/go-ldap/ldap/v3 v3.4.2
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
	github.com/kataras/iris/v12 => github.com/kataras/iris/v12 v12.2.0-beta3
	github.com/labstack/echo/v4 => github.com/labstack/echo/v4 v4.9.0
	github.com/microcosm-cc/bluemonday => github.com/microcosm-cc/bluemonday v1.0.18
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.8.4
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.16.0
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/pkg/sftp => github.com/pkg/sftp v1.13.5
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.37.0
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.4
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.12.1
)
