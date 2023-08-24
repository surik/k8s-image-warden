# Kubernetes Image Warden (KIW) Usage Guide

## Installation

Please read [Installation](./installation.md) guide.

## Configuration

As `kiw-agent` supports different CRI implementations when you deploy KIW via helm chart you have to specify where the CRI socket is located.
Please update [the value file](../chart/k8s-image-warden/values.yaml) section `agent.criEndpoint`. The following values are possible:

* `/run/containerd/containerd.sock`
* `/var/run/cri-dockerd.sock`
* `/run/crio/crio.sock`

The rules are configured as a YAML document and have to be provided in [the value file](../chart/k8s-image-warden/values.yaml) under `controller.rulesConfig`.

### Rules configuration examples

KIW rules engines support mutation and validation rules. Mutation rules run first and modify image references. After all mutations are applied engine will evaluate validation rules one by one until either one rule stops the validation pipeline or there will be no validation rules anymore.

#### Setting default registry

The following mutation pipeline contains a single rule. This will ensure that `docker.io` will be used if no registry is provided.

```yaml
    rules:
    - name: docker.io is the default registry
      mutate:
        type: DefaultRegistry
        registry: "docker.io"
```


#### Redefine registry

The following mutation pipeline extends the previous example with additional rules. That will change `quay.io` to `docker.io`.

```yaml
    rules:
    - name: docker.io is the default registry
      mutate:
        type: DefaultRegistry
        registry: "docker.io"
    - name: change quay.io to docker.io
      mutate:
        type: RewriteRegistry
        registry: "quay.io"
        newRegistry: "docker.io"
```

These are the example of pipeline evaluation:

```
nginx:latest -> docker.io/nginx:latest
docker/nginx:latest -> docker.io/nginx:latest
quay.io/nginx:latest -> docker.io/nginx:latest
ghcr.io/alpine:latest -> ghcr.io/nginx:latest
```

#### No latest tag is allowed

The following pipeline uses mutation and validation rules. Please remember that mutation rules are executed first.

```yaml
    rules:
    - name: docker.io is default registry
      mutate:
        type: DefaultRegistry
        registry: "docker.io"
    - name: no latests
      validate:
        type: Latest
        allow: false
```

These are the example of pipeline evaluation:

```
nginx:latest -> docker.io/nginx:latest -> not allowed
docker/nginx:latest -> docker.io/nginx:latest -> not allowed
quay.io/nginx:latest -> quay.io/nginx:latest -> not allowed
docker.io/nginx:1.25.1 -> docker.io/nginx:1.25.1 -> allowed
quay.io/nginx:1.25.1 -> quay.io/nginx:1.25.1 -> allowed
```

#### SemVer validation

The following pipeline uses three validation rules and ensures that:
1. no latest tag for any image is allowed
2. no dev  for docker.io/our-org is allowed
3. nginx is any 1.25.x (>=1.25.0 and <1.26.0)

```yaml
    rules:
    - name: no latests
      validate:
        type: Latest
        allow: false
    - name: no latest or dev tag
      validate:
        type: Lock
        allow: false
    - name: nginx 1.25
      validate:
        type: SemVer
        imageName: docker\.io/our-org/.*
        imageTag: "1.25.x"
        allow: true
```


#### Rolling tag validation

This type of validation is experimental and based on historical data collected by agents.
When evaluated it may reach out to the remote image to get manifest information.
**The current version of KIW doesn't **provide any guaranties that this** always works.**


```yaml
rules:
  - name: no rolling tags for our app is allowed
    validate:
      type: RollingTag
        imageName: docker\.io/our-org/app.*
      allow: false
```

For example. when attempting to deploy `docker.io/our-org/app:feature` happens KIW performs attestation of the image to ensure that the `feature` tag is immutable.
