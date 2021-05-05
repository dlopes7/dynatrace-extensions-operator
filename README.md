# Dynatrace Extensions Operator

This is Kubernetes operator that can deploy [Dynatrace OneAgent extensions](https://www.dynatrace.com/support/help/shortlink/extensions-hub#oneagent-extensions) to a kubernetes cluster.

This is specially useful when you have a lot of nodes, or does not have access to the underlying host (GKE for instance)

> :warning: **Not supported**: This project is not supported by Dynatrace, it is a personal project, use at your own risk.

# Installation

Install the operator with

```shell
kubectl apply -f https://raw.githubusercontent.com/dlopes7/dynatrace-extensions-operator/master/config/crd/bases/extensions_crd.yaml
```

You can obtain a sample extensions definition [here](https://github.com/dlopes7/dynatrace-extensions-operator/blob/master/config/samples/extensions.yaml):

Example:

```shell
curl -o extensions.yaml https://raw.githubusercontent.com/dlopes7/dynatrace-extensions-operator/master/config/samples/extensions.yaml
```

Edit the file to include to the extensions you would like to deploy

```yaml
apiVersion: dynatrace.com/v1alpha1
kind: Extension
metadata:
  name: extensions-sample
  namespace: dt-extensions
spec:
  extensions:
    - name: rabbitmq
      downloadLink: http://192.168.15.101:5656/custom.python.rabbitmq_kubernetes.zip
    - name: redis
      downloadLink: http://192.168.15.101:5656/custom.python.redis_kubernetes.zip
```

Deploy the operator with:

```shell
kubectl apply -f extensions.yaml
```

This will create a DaemonSet that guarantees the extensions will be deployed to all nodes

### TODO

* Deploy only for nodes where the OneAgent already exists
* Maybe have a public repository with some extensions as an example





