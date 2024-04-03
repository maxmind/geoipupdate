<!--- app-name: geoipupdate -->

# GeoIP Update

[GeoIP Update](https://github.com/maxmind/geoipupdate) Helm chart.

## TL;DR

```console
helm install my-release oci://registry-1.docker.io/konvergence/geoipupdate
```

## Introduction

%%INTRODUCTION%% (check existing examples)

## Prerequisites

- Kubernetes 1.23+
- Helm 3.8.0+
- PV provisioner support in the underlying infrastructure
- ReadWriteMany volumes to share geoip databases with other pods like ingress-nginx.

## Installing the Chart

To install the chart with the release name `my-release`:

```console
helm install my-release oci://REGISTRY_NAME/REPOSITORY_NAME/geoipupdate
```

> Note: You need to substitute the placeholders `REGISTRY_NAME` and `REPOSITORY_NAME` with a reference to your Helm chart registry and repository.

The command deploys geoipupdate on the Kubernetes cluster in the default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

### Common parameters

| Name      | Description             | Default        |
|:----------|:------------------------|:---------------|
| `kubeVersion` | Override Kubernetes version | `""` |
| `nameOverride` | String to partially override common.names.name | `""` |
| `fullnameOverride` | String to fully override common.names.fullname | `""` |
| `namespaceOverride` | String to fully override common.names.namespace | `""` |
| `commonLabels` | Labels to add to all deployed objects | `{}` |
| `commonAnnotations` | Annotations to add to all deployed objects | `{}` |
| `clusterDomain` | Kubernetes cluster domain name | `"cluster.local"` |
| `extraDeploy` | Array of extra objects to deploy with the release | `[]` |
| `diagnosticMode.enabled` | Enable diagnostic mode (all probes will be disabled and the command will be overridden) | `false` |
| `diagnosticMode.command` | Command to override all containers in the workload | `["sleep"]` |
| `diagnosticMode.args` | Args to override all containers in the workload | `["infinity"]` |

### GeoIP Update parameters

| Name      | Description             | Default        |
|:----------|:------------------------|:---------------|
| `image.registry` | GeoIP Update image registry | `ghcr.io` |
| `image.repository` | GeoIP Update image repository | `maxmind/geoipupdate` |
| `image.tag` | GeoIP Update image tag (immutable tags are recommended) | `v6.1` |
| `image.digest` | GeoIP Update image digest in the way sha256:aa.... Please note this parameter, if set, will override the tag image tag (immutable tags are recommended) | `""` |
| `image.pullPolicy` | Specify a imagePullPolicy. Defaults to 'Always' if image tag is 'latest', else set to 'IfNotPresent' | `IfNotPresent` |
| `image.pullSecrets` | Optionally specify an array of imagePullSecrets. Secrets must be manually created in the namespace | `[]` |
| `geoipupdate.config.existingSecret` | Name of existing secret to use for credentials | `""` |
| `geoipupdate.config.secretKeys.accountId` | Name of key in existing secret to use for accoundId. Only used when `geoipupdate.config.existingSecret` is set. | `""` |
| `geoipupdate.config.secretKeys.licenceKey` | Name of key in existing secret to use for licenceKey. Only used when `geoipupdate.config.existingSecret` is set. | `""` |
| `geoipupdate.config.secretKeys.userPassword` | Name of key in existing secret to use for userPassword. Only used when `geoipupdate.config.existingSecret` is set. | `""` |
| `geoipupdate.config.editionsIds` | GEOIPUPDATE_EDITION_IDS - List of space-separated database edition IDs. Edition IDs may consist of letters, digits, and dashes. For example, GeoIP2-City would download the GeoIP2 City database (GeoIP2-City) | `nil` |
| `geoipupdate.config.accoundId` | GEOIPUPDATE_ACCOUNT_ID - Your MaxMind account ID | `nil` |
| `geoipupdate.config.licenceKey` | GEOIPUPDATE_LICENSE_KEY - Your case-sensitive MaxMind license key | `nil` |
| `geoipupdate.config.frequency` | GEOIPUPDATE_FREQUENCY - The number of hours between geoipupdate runs. If this is not set or is set to 0, geoipupdate will run once and exit. Only used for `Deployment` workload | `nil` |
| `geoipupdate.config.host` | GEOIPUPDATE_HOST - The host name of the server to use. The default is `updates.maxmind.com` | `nil` |
| `geoipupdate.config.proxy` | GEOIPUPDATE_PROXY - The proxy host name or IP address. You may optionally specify a port number, e.g., 127.0.0.1:8888. If no port number is specified, 1080 will be used | `nil` |
| `geoipupdate.config.userPassword` | GEOIPUPDATE_PROXY_USER_PASSWORD - The proxy user name and password, separated by a colon. For instance, `username:password` | `nil` |
| `geoipupdate.config.preserveFilesTimes` | GEOIPUPDATE_PRESERVE_FILE_TIMES - Whether to preserve modification times of files downloaded from the server. This option is either 0 or 1. The default is 0 | `nil` |
| `geoipupdate.config.verbose` | GEOIPUPDATE_VERBOSE - Enable verbose mode. Prints out the steps that geoipupdate takes. Set to anything (e.g., 1) to enable | `nil` |
| `geoipupdate.kind` | Workload to use: `CronJob` or `Deployment` | `CronJob` |
| `geoipupdate.cronjob.schedule` | CronJob scheduling | `0 * * * *` |
| `geoipupdate.cronjob.timeZone` | CronJob time zone | `nil` |
| `geoipupdate.cronjob.restartPolicy` | CronJob restart policy | `OnFailure` |


### Persistence parameters

| Name      | Description             | Default        |
|:----------|:------------------------|:---------------|
| `persistence.enabled` | Enable persistence using Persistent Volume Claims | `true` |
| `persistence.mountPath` | Path to mount the volume at | `/usr/share/GeoIP` |
| `persistence.subPath` | The subdirectory of the volume to mount to, useful in dev environments and one PV for multiple services | `""` |
| `persistence.storageClass` | Storage class of backing PVC | `""` |
| `persistence.annotations` | Persistent Volume Claim annotations | `{}` |
| `persistence.accessModes` | Persistent Volume Access Modes | `["ReadWriteOnce"]` |
| `persistence.size` | Size of data volume | `1Gi` |
| `persistence.existingClaim` | The name of an existing PVC to use for persistence | `""` |
| `persistence.selector` | Selector to match an existing Persistent Volume | `{}` |
| `persistence.dataSource` | Custom PVC data source | `{}` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
helm install my-release \
  --set geoipupdate.config.editionsIds=GeoLite2-City \
  --set geoipupdate.config.accoundId=accoundId \
  --set geoipupdate.config.licenceKey=licenceKey \
  --set geoipupdate.kind=CronJob \
  --set geoipupdate.cronjob.schedule="* 0 0 0 0" \
    oci://REGISTRY_NAME/REPOSITORY_NAME/geoipupdate
```

> Note: You need to substitute the placeholders `REGISTRY_NAME` and `REPOSITORY_NAME` with a reference to your Helm chart registry and repository.

The above command deploy geoipupdate as a CronJob with credentials to update geoip databases every hour.

> NOTE: You can also deploy geoipupdate as a Deployment (see `geoipupdate.config.frequency` to configure update frequency).

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
helm install my-release -f values.yaml oci://REGISTRY_NAME/REPOSITORY_NAME/geoipupdate
```

> Note: You need to substitute the placeholders `REGISTRY_NAME` and `REPOSITORY_NAME` with a reference to your Helm chart registry and repository.
> **Tip**: You can use the default [values.yaml](./values.yaml)

## Configuration and installation details

### [Rolling VS Immutable tags](https://docs.bitnami.com/tutorials/understand-rolling-tags-containers)

It is strongly recommended to use immutable tags in a production environment. This ensures your deployment does not change automatically if the same tag is updated with a different image.

Bitnami will release a new chart updating its containers if a new version of the main container, significant changes, or critical vulnerabilities exist.

## Persistence

The [geoipupdate](https://github.com/maxmind/geoipupdate) image stores the geoipupdate data `/usr/share/GeoIP` path of the container. Persistent Volume Claims are used to keep the data across deployments.

If you encounter errors when working with persistent volumes, refer to our [troubleshooting guide for persistent volumes](https://docs.bitnami.com/kubernetes/faq/troubleshooting/troubleshooting-persistence-volumes/).

### Additional environment variables

In case you want to add extra environment variables (useful for advanced operations like custom init scripts), you can use the `extraEnvVars` property.

```yaml
geoipupdate:
  extraEnvVars:
    - name: LOG_LEVEL
      value: error
```

Alternatively, you can use a ConfigMap or a Secret with the environment variables. To do so, use the `extraEnvVarsCM` or the `extraEnvVarsSecret` values.

### Sidecars

If additional containers are needed in the same pod as geoipupdate (such as additional metrics or logging exporters), they can be defined using the `sidecars` parameter.

```yaml
sidecars:
- name: your-image-name
  image: your-image
  imagePullPolicy: Always
  ports:
  - name: portname
    containerPort: 1234
```

If these sidecars export extra ports, extra port definitions can be added using the `service.extraPorts` parameter (where available), as shown in the example below:

```yaml
service:
  extraPorts:
  - name: extraPort
    port: 11311
    targetPort: 11311
```

If additional init containers are needed in the same pod, they can be defined using the `initContainers` parameter. Here is an example:

```yaml
initContainers:
  - name: your-image-name
    image: your-image
    imagePullPolicy: Always
    ports:
      - name: portname
        containerPort: 1234
```

Learn more about [sidecar containers](https://kubernetes.io/docs/concepts/workloads/pods/) and [init containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/).

### Pod affinity

This chart allows you to set your custom affinity using the `affinity` parameter. Find more information about Pod affinity in the [kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity).

As an alternative, use one of the preset configurations for pod affinity, pod anti-affinity, and node affinity available at the [bitnami/common](https://github.com/bitnami/charts/tree/main/bitnami/common#affinities) chart. To do so, set the `podAffinityPreset`, `podAntiAffinityPreset`, or `nodeAffinityPreset` parameters.

## Troubleshooting

Find more information about how to deal with common errors related to Bitnami's Helm charts in [this troubleshooting guide](https://docs.bitnami.com/general/how-to/troubleshoot-helm-chart-issues).

## License

Copyright &copy; 2024 Broadcom. The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

<http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
