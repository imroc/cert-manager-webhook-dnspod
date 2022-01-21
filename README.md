# cert-manager-webhook-dnspod

This is a cert-manager webhook solver for [DNSPod](https://www.dnspod.cn).

## Prerequisites

* [cert-manager](https://github.com/jetstack/cert-manager) >= 1.6.0

## Installation

Generate SecretId and SecretKey in [Cloud API](https://console.cloud.tencent.com/cam/capi)

```console
$ helm repo add roc https://charts.imroc.cc
$ helm upgrade --install cert-manager-webhook-dnspod roc/cert-manager-webhook-dnspod \
    --namespace <NAMESPACE> \
    --set clusterIssuer.secretId=<SECRET_ID> \
    --set clusterIssuer.secretKey=<SECRET_KEY> 
```

Notice: **`secretId`, `secretKey` is not DNSPod secret, it's tencent cloud secret!**

## Create Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-crt
spec:
  secretName: example-crt
  issuerRef:
    name: dnspod
    kind: ClusterIssuer
    group: cert-manager.io
  dnsNames:
  - "example.com"
  - "*.example.com"
```