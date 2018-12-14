# Mutating trace admission controller

[Mutating admission controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) that injects base64 encoded [OpenCensus span context](https://github.com/census-instrumentation/opencensus-specs/blob/master/trace/Span.md#spancontext) into the `trace.kubernetes.io/context` object annotation.

## Purpose

The trace context injected with this mutating controller can be used by Kubernetes components to export traces associated with object lifecycles. For more information on this effort, [please refer to the official KEP](https://github.com/kubernetes/enhancements/pull/650).


## Quick start

The structure of this mutating admission controller was informed by the [mutating admission webhook found here](https://github.com/morvencao/kube-mutating-webhook-tutorial). The basic idea is as follows:

1) Create an HTTPS-enabled server that takes Pod json from the API server, inserts encoded span context as an annotation, and returns it 
2) Run a deployment with this webhook server, and expose it as a service
3) Create a `MutatingWebhookConfiguration` which instructs the API server to send Pod objects to the aforementioned service upon creation

The included `Makefile` makes these steps straightforward and the available commands are as follows:

* `make docker`: build local Docker image
* `make cluster-up`: apply certificate configuration and deployment configuration to cluster for the mutating webhook
* `make cluster-down`: delete resources associated with the mutating webhook from the active cluster

There are example patches which can be used with `kustomize` to configure the deployment of this webhook into your cluster under `deploy/base/overlays/example`. This example custom configuration can be applied with:  

`kustomize build deploy/overlays/example | kubectl apply -f -`

This can be used, for example, to set different sampling policies between production and staging clusters.