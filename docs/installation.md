# Installation

The application can be deployed as helm release to any existing Kubernetes cluster.

Install [cert-manager](https://artifacthub.io/packages/helm/cert-manager/cert-manager) CRDs:

    $ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.3/cert-manager.crds.yaml

And then install k8s-image-warden:

	$ helm --namespace kiw  upgrade -i --create-namespace prod chart/k8s-image-warden

