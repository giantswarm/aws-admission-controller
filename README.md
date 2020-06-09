# Local development
    kind create cluster
    CGO_ENABLED=0 go build .
    docker build . -t admission-controller:dev
    kind load docker-image admission-controller:dev
    opsctl ensure crds -k "$(kind get kubeconfig)" -p aws
    kubectl apply --context kind-kind -f deploy/certmanager.yml
    # Wait until certmanaget is up
    kubectl apply --context kind-kind -f deploy/clusterissuer.yml
    kubectl apply --context kind-kind -f deploy/deploy.yaml

    kind delete cluster
# g8s-admission-controller
