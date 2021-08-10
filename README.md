[![CircleCI](https://circleci.com/gh/giantswarm/aws-admission-controller.svg?style=shield)](https://circleci.com/gh/giantswarm/aws-admission-controller)

# AWS Admission Controller

Giant Swarm AWS Management Cluster admission controller that implements the following rules:

Mutating Webhook:

- In an `AWSCluster` resource, the AWS Operator Version is defaulted based on the `Release` CR if it is not set. 
- In an `AWSCluster` resource, the Release Version is defaulted based on the `Cluster` CR if it is not set. 
- In an `AWSCluster` resource, the Credential Secret is defaulted if it is not set. 
- In an `AWSCluster` resource, the Region is defaulted if it is not set. 
- In an `AWSCluster` resource, the Description is defaulted if it is not set. 
- In an `AWSCluster` resource, the DNS Domain is defaulted if it is not set. 
- In an `AWSCluster` resource, the Pod CIDR is defaulted if it is not set. 
- In an `AWSCluster` resource, in a pre-HA version, the Master attribute is defaulted if it is not set.

- In a `Cluster` resource, the Release Version is defaulted to the newest active production version if it is not set. 
- In a `Cluster` resource, the Cluster Operator Version is defaulted based on the `Release` CR if it is not set. 
- In a `Cluster` resource, the Cluster Operator Version is defaulted based on the new release version during an upgrade. 

- In a `G8sControlplane` resource, the Cluster Operator Version is defaulted based on the `Cluster` CR if it is not set. 
- In a `G8sControlplane` resource, the Release Version is defaulted based on the `Cluster` CR if it is not set. 
- In a `G8sControlPlane` resource, when the `.spec.replicas` is changed from 1 to 3, the Availability Zones of the according `AWSControlPlane` will be defaulted if needed.
- In a `G8sControlPlane` resource, the replicas attribute will be defaulted if it is not defined.
  - For HA-Versions, in case the matching `AWSControlPlane` already exists, the number of AZs determines the value of `replicas`.
    In case no such `AWSControlPlane` exists, the default number of AZs is assigned. 
  - For pre-HA versions, replicas is always set to 1 for a single master cluster.
- In a `G8sControlPlane` resource, the infrastructure reference will be set to point to the matching `AWSControlPlane`.
- In a `G8sControlPlane` resource, the control-plane label will be defaulted to its name if it is not set.

- In an `AWSControlplane` resource, the AWS Operator Version is defaulted based on the `AWSCluster` CR if it is not set. 
- In an `AWSControlplane` resource, the Release Version is defaulted based on the `Cluster` CR if it is not set. 
- In an `AWSControlPlane` resource, the Availability Zones will be defaulted if they are `nil`. 
  - For HA-Versions, in case the matching `G8sControlPlane` already exists, the number of AZs is determined by the number of `replicas` defined there. 
    In case no such `G8sControlPlane` exists, the default number of AZs is assigned. 
  - For Pre-HA-Versions, in case the matching `AWSCluster` already exists, the AZ is taken from there. 
- In an `AWSControlPlane` resource, the Instance Type will be defaulted if it is not defined. 
  - For HA-Versions, the default Instance Type is chosen. 
  - For Pre-HA-Versions, in case the matching `AWSCluster` already exists, the Instance Type is taken from there. 
- In a `AWSControlPlane` resource, the control-plane label will be defaulted to its name if it is not set.

- In an `AWSMachinedeployment` resource, the Availability Zones will be defaulted if they are `nil`. The default number of   
  AZs is assigned based on the master AZs taken from the `AWSControlPlane` CR.
- In an `AWSMachinedeployment` resource, the AWS Operator Version is defaulted based on the `AWSCluster` CR if it is not set. 
- When a new `AWSMachineDeployment` is created, details are logged.
- In an `AWSMachinedeployment` resource, the Release Version is defaulted based on the `Cluster` CR if it is not set. 

- In a `Machinedeployment` resource, the Release Version is defaulted based on the `Cluster` CR if it is not set. 
- In a `Machinedeployment` resource, the Cluster Operator Version is defaulted based on the `Cluster` CR if it is not set. 

Validating Webhook:

- In a `G8sControlPlane` resource, it validates the Master Node Replicas are a valid count (Right now either 1 or 3).
- In a `G8sControlPlane` resource, it validates the Master Node Replicas are matching the number of Availability Zones in the `AWSControlPlane` resource.
- In an `G8sControlPlane` resource, it validates that the control-plane label is set.

- In an `AWSControlPlane` resource, it validates the Master Instance Type is a valid Instance Type for the installation.
- In an `AWSControlPlane` resource, it validates that the order of Master Node Availability Zones does not change on update.
- In an `AWSControlPlane` resource, it validates that the number of distinct Master Node Availability Zones is maximal.
- In an `AWSControlPlane` resource, it validates the Master Node Availability Zones are valid AZs for the installation.
- In an `AWSControlPlane` resource, it validates the Master Node Availability Zones are a valid count (Right now either 1 or 3).
- In an `AWSControlPlane` resource, it validates the Master Node Availability Zones are matching the number of Replicas in the `G8sControlPlane` resource.
- In an `AWSControlPlane` resource, it validates that the control-plane label is set.

- In an `AWSMachineDeployment` resource, it validates the worker node instance type.
- In an `AWSMachineDeployment` resource, it validates the worker node availability zones.
- In an `AWSMachineDeployment` resource, it validates the Machine Deployment ID is matching against `MachineDeployment` resource.
- In an `AWSMachineDeployment` resource, on creation it validates that the `Cluster` is not deleted.
- In an `AWSMachinedeployment` resource, it validates that the `max` number of nodes is greater or equal to `min`.

- In a `Cluster` resource, the  release version label can only be changed to an existing and non-deprecated release by admin users and users in restricted groups. 
- In a `Cluster` resource, the  release version label can only be changed to a major version that is greater than the current one   
- In a `Cluster` resource, the  release version label can only be changed if the cluster is in a transitioned condition. ("updated" or "created")
  but does not skip major versions by admin users and users in restricted groups. 
- In a `Cluster` resource, the non-version label values are not allowed to be deleted or renamed by admin users and users in restricted groups. 
- In a `Cluster` resource, the `giantswarm.io` label keys are not allowed to be deleted or renamed by admin users and users in restricted groups. 

- In a `MachineDeployment` resource, on creation it validates that the `Cluster` is not deleted.

- In a `NetworkPool` resource, it validates the .Spec.CIDRBlock from other NetworkPools and also checks if there's overlapping from Docker CIDR, Kubernetes cluster IP range or tenant cluster CIDR.

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
