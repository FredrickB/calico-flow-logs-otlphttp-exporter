# calico-flow-logs-otlphttp-exporter

![Version: 0.11.3](https://img.shields.io/badge/Version-0.11.3-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.13.6](https://img.shields.io/badge/AppVersion-0.13.6-informational?style=flat-square)

Export network flow logs from Calico using OTLP/HTTP

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| env.CA_CERT_PATH | string | `"/etc/goldmane/ca-cert/tigera-ca-bundle.crt"` | Path to CA certificate used for mTLS connection to Goldmane |
| env.GOLDMANE_HOST | string | `"goldmane:7443"` | Host and port of Goldmane, must be present as SAN in certificate used for mTLS |
| env.PRIVATE_KEY_PATH | string | `"/etc/goldmane/certs/tls.key"` | Path to private key used for mTLS connection to Goldmane |
| env.PUBLIC_CERT_PATH | string | `"/etc/goldmane/certs/tls.crt"` | Path to public certificate used for mTLS connection to Goldmane |
| env.RECONNECT_WAIT_TIME_IN_MILLISECONDS | string | `"5000"` | Amount of milliseconds to wait before attempting to reconnect to Goldmane in the event of connection error |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Image pullpolicy |
| image.repository | string | `"ghcr.io/fredrickb/calico-flow-logs-otlphttp-exporter"` | Image repository |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| replicaCount | int | `1` | Number of pods, this should not be more than 1, that would cause duplicate amounts of data to be sent to OTLP/HTTP endpoint |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automount | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |
| volumeMounts[0] | object | `{"mountPath":"/etc/goldmane/certs","name":"goldmane-certs","readOnly":true}` | VolumeMount for Goldmane mTLS certificates |
| volumeMounts[1] | object | `{"mountPath":"/etc/goldmane/ca-cert","name":"goldmane-ca-cert","readOnly":true}` | VolumeMount for Goldmane CA bundle |
| volumes[0] | object | `{"name":"goldmane-certs","secret":{"optional":false,"secretName":"goldmane-key-pair"}}` | Secret containing Goldmane mTLS certificates |
| volumes[1] | object | `{"configMap":{"items":[{"key":"tigera-ca-bundle.crt","path":"tigera-ca-bundle.crt"}],"name":"goldmane-ca-bundle"},"name":"goldmane-ca-cert"}` | ConfigMap containing Goldmane CA bundle |

