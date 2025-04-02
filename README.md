# cert-manager-webhook-dnspod

This is a cert-manager webhook solver for [DNSPod](https://www.dnspod.cn).

## Prerequisites

* [cert-manager](https://github.com/cert-manager/cert-manager) >= 1.13.0

## Installation

### Use Helm

First, generate `SecretId` and `SecretKey` in [Cloud API](https://console.cloud.tencent.com/cam/capi)

You can install chart from git repo:

```bash
# Firstly add cert-manager-webhook-dnspod charts repository if you haven't do this
helm repo add cert-manager-webhook-dnspod https://imroc.github.io/cert-manager-webhook-dnspod
# Install the latest version.
helm upgrade --install --namespace cert-manager \
  cert-manager-webhook-dnspod cert-manager-webhook-dnspod/cert-manager-webhook-dnspod
```
## Use Kubectl

Use `kubectl apply` to install:

```bash
kubectl apply -f https://raw.githubusercontent.com/imroc/cert-manager-webhook-dnspod/master/bundle.yaml
```

## Usage

### Prepare Issuer

Before you can issue a certificate, you need to create a `Issuer` or `ClusterIssuer`.

> If you use helm and only need a global `ClusterIssuer`, you can add `--set clusterIssuer.enabled=true --set clusterIssuer.secretId=xxx --set clusterIssuer.secretKey=xxx` to create the `ClusterIssuer`.

Firstly, create a secret that contains TencentCloud account's `SecretId` and `SecretKey`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dnspod-secret
  namespace: cert-manager
type: Opaque
stringData:
  secretId: xxx
  secretKey: xxx
```

> Base64 is not needed in `stringData`.

Then you can create a `ClusterIssuer` referring the secret:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: dnspod
spec:
  acme:
    email: roc@imroc.cc
    privateKeySecretRef:
      name: dnspod-letsencrypt
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      - dns01:
          webhook:
            config:
              secretIdRef:
                key: secretId
                name: dnspod-secret
              secretKeyRef:
                key: secretKey
                name: dnspod-secret
              ttl: 600
              recordLine: ""
            groupName: acme.imroc.cc
            solverName: dnspod
```

1. `secretId` and `secretKey` is the SecretId and SecretKey of your TencentCloud account.
2. `groupName` is the the groupName that specified in your cert-manager-webhook-dnspod installation, defaults to `acme.imroc.cc`.
3. `solverName` must be `dnspod`.
4. `ttl` is the optional ttl of dns TXT record that created by webhook.
5. `recordLine` is the optional recordLine parameter of the dnspod.
6. `email` is the optional email address. When the domain is about to expire, a notification will be sent to this email address.

### Issue Certificate

You can issue the certificate by creating `Certificate` that referring the dnspod `ClusterIssuer`:

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
