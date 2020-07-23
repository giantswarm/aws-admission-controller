[![CircleCI](https://circleci.com/gh/giantswarm/admission-controller.svg?style=svg)](https://circleci.com/gh/giantswarm/admission-controller)

# G8S Admission Controller

Giant Swarm Control Plane admission controller that implements the following rules:

- In a G8SControlPlane resource, when the `.spec.replicas` is changed from 1 to 3, the Availability Zones of the according AWSControlPlane will be defaulted if needed.
- In an AWSControlPlane resource, the Availability Zones will be defaulted if they are `nil`.
- When a new AWSMachineDeployment is created, details are logged.

The certificates for the webhook are created with CertManager and injected through the CA Injector.

## Ownership

Firecracker Team

### Local Development

Testing the admission-controller in a kind cluster on your local machine:

```nohighlight
kind create cluster

# Build a linux image
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
docker build . -t admission-controller:dev
kind load docker-image admission-controller:dev

# Make sure the Custom Resource Definitions are in place
opsctl ensure crds -k "$(kind get kubeconfig)" -p aws

# Insert the certificate
kubectl apply --context kind-kind -f local_dev/certmanager.yml

## Wait until certmanager is up

kubectl apply --context kind-kind -f local_dev/clusterissuer.yml
helm template admission-controller -f helm/admission-controller/ci/default-values.yaml helm/admission-controller > local_dev/deploy.yaml

## Replace image name with admission-controller:dev
kubectl apply --context kind-kind -f local_dev/deploy.yaml
kind delete cluster
```

## Changelog

See [Releases](https://github.com/giantswarm/admission-controller/releases)

## Contact

- Bugs: [issues](https://github.com/giantswarm/admission-controller/issues)
- Please visit https://www.giantswarm.io/responsible-disclosure for information on reporting security issues.

## Contributing, reporting bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## Publishing a release

See [docs/Release.md](https://github.com/giantswarm/admission-controller/blob/master/docs/release.md)

## Add a new webhook

See [docs/webhook.md](https://github.com/giantswarm/admission-controller/blob/master/docs/webhook.md)

## Writing tests

See [docs/tests.md](https://github.com/giantswarm/admission-controller/blob/master/docs/tests.md)
