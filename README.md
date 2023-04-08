# cert-manager-webhook-dnspod

This is a cert-manager webhook solver for [DNSPod](https://www.dnspod.cn).

## Prerequisites

* [cert-manager](https://github.com/jetstack/cert-manager) >= 1.6.0

## Installation

### Helm

Generate SecretId and SecretKey in [Cloud API](https://console.cloud.tencent.com/cam/capi)

```console
$ helm pull oci://registry-1.docker.io/imroc/cert-manager-webhook-dnspod --untar
$ helm upgrade --install cert-manager-webhook-dnspod ./cert-manager-webhook-dnspod \
    --namespace cert-manager \
    --set clusterIssuer.secretId=<SECRET_ID> \
    --set clusterIssuer.secretKey=<SECRET_KEY> 
```

Notice: **`secretId`, `secretKey` is not DNSPod secret, it's tencent cloud secret!**

Then create certificate referring auto-created ClusterIssuer:

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

## Kubectl Apply

Use `kubectl apply` to install:

```bash
kubectl apply -f https://raw.githubusercontent.com/imroc/cert-manager-webhook-dnspod/master/bundle.yaml
```

Create a secret that contains TencentCloud account's `SecretKey`:

```yaml
apiVersion: v1
stringData:
  secret-key: ******
kind: Secret
metadata:
  name: dnspod-secret
  namespace: cert-manager
type: Opaque
```

> base64 is not need in `stringData`.

Create a `ClusterIssuer` referring the secret:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: dnspod
spec:
  acme:
    email: roc@imroc.cc
    preferredChain: ""
    privateKeySecretRef:
      name: dnspod-letsencrypt
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
      - dns01:
          webhook:
            config:
              secretId: ************************************
              secretKeyRef:
                key: secret-key
                name: dnspod-secret
              ttl: 600
            groupName: acme.imroc.cc
            solverName: dnspod
```

> `secretId` is the SecretId of your TencentCloud account.

Create the `Certificate` you want:

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
