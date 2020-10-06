[![CircleCI](https://circleci.com/gh/giantswarm/aws-admission-controller.svg?style=shield)](https://circleci.com/gh/giantswarm/aws-admission-controller)

# G8S Admission Controller

Giant Swarm Control Plane admission controller that implements the following rules:

Mutating Webhook:

- In a `G8sControlPlane` resource, when the `.spec.replicas` is changed from 1 to 3, the Availability Zones of the according `AWSControlPlane` will be defaulted if needed.
- In a `G8sControlPlane` resource, the replicas attribute will be defaulted if it is not defined.
  - For HA-Versions, in case the matching `AWSControlPlane` already exists, the number of AZs determines the value of `replicas`.
    In case no such `AWSControlPlane` exists, the default number of AZs is assigned. 
  - For pre-HA versions, replicas is always set to 1 for a single master cluster.
- In a `G8sControlPlane` resource, the infrastructure reference will be set to point to the matching `AWSControlPlane` resource if it already exists.

- In an `AWSControlPlane` resource, the Availability Zones will be defaulted if they are `nil`. 
  - For HA-Versions, in case the matching `G8sControlPlane` already exists, the number of AZs is determined by the number of `replicas` defined there. 
    In case no such `G8sControlPlane` exists, the default number of AZs is assigned. 
  - For Pre-HA-Versions, in case the matching `AWSCluster` already exists, the AZ is taken from there. 
- In an `AWSControlPlane` resource, the Instance Type will be defaulted if it is not defined. 
  - For HA-Versions, the default Instance Type is chosen. 
  - For Pre-HA-Versions, in case the matching `AWSCluster` already exists, the Instance Type is taken from there. 
- On creation of an `AWSControlPlane` resource, the infrastructure reference of the according `G8sControlPlane` will be set if needed.

- When a new `AWSMachineDeployment` is created, details are logged.

Validating Webhook:

- In a `G8sControlPlane` resource, it validates the Master Node Replicas are a valid count (Right now either 1 or 3).
- In a `G8sControlPlane` resource, it validates the Master Node Replicas are matching the number of Availability Zones in the `AWSControlPlane` resource.

- In a `AWSControlPlane` resource, it validates the Control Plane ID is matching against `G8sControlPlane` resource.
- In a `AWSMachineDeployment` resource, it validates the Machine Deployment ID is matching against `MachineDeployment` resource.

The certificates for the webhook are created with CertManager and injected through the CA Injector.

## Ownership

Firecracker Team

### Local Development

Testing the aws-admission-controller in a kind cluster on your local machine:

```nohighlight
kind create cluster

# Build a linux image
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
docker build . -t aws-admission-controller:dev
kind load docker-image aws-admission-controller:dev

# Make sure the Custom Resource Definitions are in place
opsctl ensure crds -k "$(kind get kubeconfig)" -p aws

# Insert the certificate
kubectl apply --context kind-kind -f local_dev/certmanager.yml

## Wait until certmanager is up

kubectl apply --context kind-kind -f local_dev/clusterissuer.yml
helm template aws-admission-controller -f helm/aws-admission-controller/ci/default-values.yaml helm/aws-admission-controller > local_dev/deploy.yaml

## Replace image name with aws-admission-controller:dev
kubectl apply --context kind-kind -f local_dev/deploy.yaml
kind delete cluster
```

## Changelog

See [Releases](https://github.com/giantswarm/aws-admission-controller/releases)

## Contact

- Bugs: [issues](https://github.com/giantswarm/aws-admission-controller/issues)
- Please visit https://www.giantswarm.io/responsible-disclosure for information on reporting security issues.

## Contributing, reporting bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## Publishing a release

See [docs/Release.md](https://github.com/giantswarm/aws-admission-controller/blob/master/docs/release.md)

## Add a new webhook

See [docs/webhook.md](https://github.com/giantswarm/aws-admission-controller/blob/master/docs/webhook.md)

## Writing tests

See [docs/tests.md](https://github.com/giantswarm/aws-admission-controller/blob/master/docs/tests.md)
