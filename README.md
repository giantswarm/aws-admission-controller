# G8S Admission Controller

Giant Swarm Control Plane admission controller that implements the following rules:

- When the G8SControlPlane replicas is changed from 1 to 3 the Availavility Zones of the AWSControlPlane will be defaulted if needed.

The certificates for the webhook are created with CertManager and injected through the CA Injector.

## Ownership

Firecracker Team

## Local development

    kind create cluster
    CGO_ENABLED=0 go build .
    docker build . -t admission-controller:dev
    kind load docker-image admission-controller:dev
    opsctl ensure crds -k "$(kind get kubeconfig)" -p aws
    kubectl apply --context kind-kind -f local_dev/certmanager.yml
    ## Wait until certmanaget is up
    kubectl apply --context kind-kind -f local_dev/clusterissuer.yml
    helm template admission-controller -f helm/admission-controller/ci/default-values.yaml helm/admission-controller > local_dev/deploy.yaml
    ## Replace image name with admission-controller:dev
    kubectl apply --context kind-kind -f local_dev/deploy.yaml
    kind delete cluster
