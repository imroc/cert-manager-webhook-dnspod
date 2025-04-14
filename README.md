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
### Use Kubectl

Use `kubectl apply` to install:

```bash
kubectl apply -f https://raw.githubusercontent.com/imroc/cert-manager-webhook-dnspod/master/bundle.yaml
```

## Usage

### Cridentials

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

### Create Issuer

Before you can issue a certificate, you need to create a `Issuer` or `ClusterIssuer`.

> If you use helm and only need a global `ClusterIssuer`, you can add `--set clusterIssuer.enabled=true --set clusterIssuer.secretId=xxx --set clusterIssuer.secretKey=xxx` to create the `ClusterIssuer`.

Create a `ClusterIssuer` referring the secret:

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
            groupName: acme.dnspod.com
            solverName: dnspod
```

1. `secretId` and `secretKey` is the SecretId and SecretKey of your TencentCloud account.
2. `groupName` is the the groupName that specified in your cert-manager-webhook-dnspod installation, defaults to `acme.dnspod.com`.
3. `solverName` must be `dnspod`.
4. `ttl` is the optional ttl of dns TXT record that created by webhook.
5. `recordLine` is the optional recordLine parameter of the dnspod.
6. `email` is the optional email address. When the domain is about to expire, a notification will be sent to this email address.

### Create Certificate

You can issue the certificate by creating `Certificate` that referring the dnspod `ClusterIssuer` or `Issuer`:

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
## Upgrade 1.4.0 to 1.5.x

Version `1.5.x` has the following changes

- Enhanced logging capabilities: Added JSON-structured logging with richer context for easier troubleshooting, and supports custom log levels (default: `info`).
- Improved security: Store `secretId` in Secrets instead of Helm values (consistent with `secretKey` handling).
- Enhanced code maintainability: Refactored codebase by splitting logic into multiple Go files for better organization.
- Optimized `Present` implementation: 
  - Removed redundant DNS SOA queries to resolve zones (The `ResolvedZone` sent by cert-manager is the zone already queried through SOA).
  - Eliminated domain lookup via DNSPod API (DNSPod API `CreateRecord` can accepts `Domain` directly without requiring `DomainID`).
- Changed default `groupName` from `acme.imroc.cc` to `acme.dnspod.com`.
- Added support for gitHub pages as helm repository.
- Added optional `recordLine` in Issuer's webhook config for custom DNS record lines.

If you upgrade from 1.4.0 to 1.5.x, and created `Issuer` or `ClusterIssuer` manually (`clusterIssuer.enabled=false`), you need to add `secretIdRef` to Issuer's webhook config, also add `secretId` in your corresponding `Secret`.

