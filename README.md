# G8S Admission Controller

Giant Swarm Control Plane admission controller that implements the following rules:

- When the G8SControlPlane replicas is changed from 1 to 3 the Availavility Zones of the AWSControlPlane will be defaulted if needed.

The certificates for the webhook are created with CertManager and injected through the CA Injector.

## Ownership

Firecracker Team

## Local development

    kind create cluster
    CGO_ENABLED=0 go build .
    docker build . -t g8s-admission-controller:dev
    kind load docker-image g8s-admission-controller:dev
    opsctl ensure crds -k "$(kind get kubeconfig)" -p aws
    kubectl apply --context kind-kind -f deploy/certmanager.yml
    # Wait until certmanaget is up
    kubectl apply --context kind-kind -f deploy/clusterissuer.yml
    kubectl apply --context kind-kind -f deploy/deploy.yaml
    kind delete cluster
