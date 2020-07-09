[![CircleCI](https://circleci.com/gh/giantswarm/admission-controller.svg?style=svg)](https://circleci.com/gh/giantswarm/admission-controller)

# G8S Admission Controller

Giant Swarm Control Plane admission controller that implements the following rules:

- When the G8SControlPlane replicas is changed from 1 to 3 the Availavility Zones of the AWSControlPlane will be defaulted if needed.

The certificates for the webhook are created with CertManager and injected through the CA Injector.

## Ownership

Firecracker Team

## Getting the project

Clone the git repository: https://github.com/giantswarm/admission-controller

### How to build

Build it using the `make` command.

```bash
$ cd admission-controller
$ make
```

### Local Development

Testing the admission-controller in a kind cluster on your local machine:

```bash
kind create cluster
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
docker build . -t admission-controller:dev
kind load docker-image admission-controller:dev
opsctl ensure crds -k "$(kind get kubeconfig)" -p aws
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

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/admission-controller/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## Publishing a Release

See [docs/Release.md](https://github.com/giantswarm/admission-controller/blob/master/docs/release.md)

## Add a new webhook

See [docs/webhook.md](https://github.com/giantswarm/admission-controller/blob/master/docs/webhook.md)

## Writing Tests 

See [docs/tests.md](https://github.com/giantswarm/admission-controller/blob/master/docs/tests.md)
