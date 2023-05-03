<!--
Copyright The Shipwright Contributors

SPDX-License-Identifier: Apache-2.0
-->

# Running on development mode

The following document highlights how to deploy a Build controller locally for running on development mode.

**Before generating an instance of the Build controller, ensure the following:**

- Target your Kubernetes cluster. We recommend the usage of [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/) for development.

- On the cluster, ensure the Tekton Pipelines is deployed. Please follow the [official documentation](https://tekton.dev/docs/pipelines/install/).

---

Once the code have been modified, you can generate an instance of the Build controller running locally to validate your changes. For running the Build controller locally via the `local` target:

```sh
pushd $GOPATH/src/github.com/shipwright-io/build
  make local
popd
```

_Note_: The above target will uninstall/install all related CRDs and start an instance of the controller. All existing CRDs instances will be deleted.
