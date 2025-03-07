# Controllers

This document describes the controllers used by Lifecycle Manager.

## Kyma Controller

[Kyma Controller](../../internal/controller/kyma_controller.go) is dealing with the introspection, interpretation and status update of the [Kyma custom resource (CR)](/api/v1beta2/kyma_types.go).

Its main responsibilities are:

1. Interpret the `.spec.modules` list and use the correct [ModuleTemplate CR](/api/v1beta2/moduletemplate_types.go) for a module.
2. Translate the ModuleTemplate CR into a [Manifest CR](/api/v1beta2/manifest_types.go) and create it with an OwnerReference to the Kyma CR where the module was listed.
3. Propagate changes from ModuleTemplate CR updates (e.g. updates to the Module Layers contained in the OCI Descriptor) into the correct Manifest CR and process upgrades, but prohibit downgrades.
4. Track all created Manifest CRs and aggregate the status into a `State`, that reflects the integrity of the Kyma installation managed by Lifecycle Manager.
5. Synchronize all the above changes to the Kyma CR Status as well as available ModuleTemplate CRs into a remote cluster.
To determine the cluster to sync to, fields in **.spec.remote** are evaluated.
This allows the use of ModuleTemplate CRs in a cluster managed by Lifecycle Manager
while Kyma Control Plane is in a different cluster.

### Remote Synchronization

In order to synchronize remote clusters, the Kyma controller uses the concept of a _virtual_ resource.
The virtual resource is a superset of the specification of the control plane and runtime data of a module.
The synchronization of these is kept up-to-date with every reconciliation,
and will only be triggered if `operator.kyma-project.io/sync=true` label is set to Kyma CR.

In this case, a so called `SyncContext` is initialized.
Every time the Kyma on the control plane is enqueued for synchronization,
it's spec is merged with the remote specification through our [custom synchronization handlers](/pkg/remote/kyma_synchronization_context.go).
These are not only able to synchronize the Kyma resource in the remote,
but they also replace the specification for all further parts of the reconciliation as a _virtual_ Kyma.
For more information, checkout the `ReplaceWithVirtualKyma` function.

## Mandatory Modules Controllers

Lifecycle Manager uses two Mandatory Modules Controllers:
- [Mandatory Modules Installation Controller](../../internal/controller/mandatory_modules_installation_controller.go) deals with the reconciliation of mandatory modules
- [Mandatory Modules Deletion Controller](../../internal/controller/mandatory_modules_deletion_controller.go) deals with the deletion of mandatory modules

Since the channel concept does not apply to mandatory modules, the Mandatory Modules Installation Controller fetches all the Mandatory ModuleTemplate CRs without any channel filtering. It then translates the ModuleTemplate CR for the mandatory module to a [Manifest CR](/api/v1beta2/manifest_types.go) with an OwnerReference to the Kyma CR. Similarly to the [Kyma Controller](/internal/controller/kyma_controller.go),
it propagates changes from the ModuleTemplate CR to the Manifest CR. The mandatory ModuleTemplate CR is not synchronized to the remote cluster and the module status does not appear in the Kyma CR status. If a mandatory module needs to be removed from all clusters, the corresponding ModuleTemplate CR needs to be deleted. The Mandatory Module Deletion Controller picks this event up and marks all associated Manifest CRs for deletion. To ensure that the ModuleTemplate CR is not removed immediately, the controller adds a finalizer to the ModuleTemplate CR. Once all associated Manifest CRs are deleted, the finalizer is removed and the ModuleTemplate CR is deleted.


## Manifest Controller

[Manifest Controller](../../internal/controller/manifest_controller.go) deals with the reconciliation and installation of data desired through a Manifest CR, a representation of a single module desired in a cluster.
Since it mainly is a delegation to the [declarative reconciliation library](/internal/declarative/README.md) with certain [internal implementation additions](/internal/manifest/README.md) please look at the respective documentation for these parts to understand them more.

## Watcher Controller

[Watcher Controller](../../internal/controller/watcher_controller.go) deals with the update of VirtualService rules derived from the [Watcher CR](/api/v1beta2/watcher_types.go). This is then used to initialize the Watcher CR from the Kyma Controller in each runtime, a small component initialized to propagate changes from the runtime(remote) clusters back to react to changes that can affect the Manifest CR integrity.
