## v1beta1 to v1beta2 API changes

### `Kyma` Custom Resource:

 The main change to the `Kyma` Resource in the `v1beta2` API version is removal of the  `.spec.sync` attribute.
 The `v1beta1` `.spec.sync` sub-attributes handling is changed as described:
 
 - `.sync.enabled` - replaced by a label on the `Kyma` object. See "Kyma synchronization labels" (link TBD) for details
 - `.sync.moduleCatalog` - replaced by a combination of labels on the `Kyma` and `ModuleTemplate` objects. See "Kyma synchronization labels" (link TBD) for details.
 - `.sync.strategy` - replaced with `sync-strategy` annotation on the `Kyma` Resource.  By default, the sync strategy is always `local-secret`, the other values are used only for testing purposes.
 - `.sync.namespace` - replaced with a `sync-namespace` command-line flag for the Lifecycle-Manager. It means the namespace the `Kyma` is synchronized to is no longer user-configurable for every `Kyma` object, but it is the same for all objects for the given Lifecycle-Manager instance, and user's can't change it.
 - `.sync.noModuleCopy` - dropped. Currently the remote `Kyma` `.spec.modules[]` is always initialized as empty. 
 
 
### `ModuleTemplate` CustomResource

Changes in the `v1beta2` API version:
- `.sync.target` attribute is replaced with a `in-kcp-mode` command-line flag for the Lifecycle Manager.  It means the module template synchronization is no longer user-configurable for every `ModuleTemplate` object, but it is the same for all `ModuleTemplates` for the given Lifecycle-Manager instance and user's can't change it.


### Synchronization of `ModuleTemplate` catalog to remote clusters

(Note for reviewer: This section is not a changelog, this is part of `v1beta2` API docs)

Lifecycle Manager ensures that the Module Catalog is correctly synchronized to the users' runtimes.
The Module Catalog consists of all the modules that are available for the users. It may be different for different users.
The synchronization mechanism described below is essential to allow users to enable modules in their clusters.
The mechanism is controlled by the set of flags that are configured on the `Kyma` and `ModuleTemplate` CRs in the control plane.
The v1beta2 API introduces three groups of Modules:
- standard modules, synchronized by default.
- internal modules, synchronized per-cluster only if confgured explicitly on the corresponding `Kyma` CRs. You mark a `ModuleTemplate` CR as "internal" by setting a label `operator.kyma-project.io/internal` to `true`
- beta modules, synchronized per-cluster only if confgured explicitly on the corresponding `Kyma` CRs. You mark a `ModuleTemplate` CR as "beta" by setting a label `operator.kyma-project.io/beta` to `true`

By default, without any flags configured on the `Kyma` and `ModuleTemplate` CRs, the `ModuleTemplate` CR, as a "standard" one, is synchronized (copied) to the remote cluster. 


Synchronization labels available on the `Kyma` CR:

- `operator.kyma-project.io/sync` A boolean flag. If set to false, the Module Catalog synchronization is disabled for given `Kyma` CR, and so, for the related remote cluster (Managed Kyma Runtime). Defaults to true if not set explicitly.
- `operator.kyma-project.io/internal` A boolean flag. If set to true, the `ModuleTemplates` labelled with the same label - so called "internal modules" are also synchronized to the remote cluster. Defaults to false if not set explicitly.
- `operator.kyma-project.io/beta` A boolean flag. If set to true, the `ModuleTemplates` labelled with the same label - so called "beta modules" are also synchronized to the remote cluster. Defaults to false if not set explicitly.

Synchronization labels available on the `ModuleTemplate` CR:

- `operator.kyma-project.io/sync` A boolean flag. If set to false, this `ModuleTemplate` CR is not synchronized to any remote cluster. Defaults to true if not set explicitly.
- `operator.kyma-project.io/internal` A boolean flag. If set to true, marks the `ModuleTemplate` as an "internal" module. It is then synchronized only for these remote clusters which are managed by the `Kyma` CR having the same label (`operator.kyma-project.io/internal`) explicitly set to `true`. Defaults to false if not set explicitly.
- `operator.kyma-project.io/beta` A boolean flag. If set to true, marks the `ModuleTemplate` as a "beta" module. It is then synchronized only for these remote clusters which are managed by the `Kyma` CR having the same label (`operator.kyma-project.io/beta`) explicitly set to `true`. Defaults to false if not set explicitly.


Note: Disabling synchronization for already synchronized `ModuleTemplates` does not remove them from remote clusters! They remain as they are, but any subsequent changes to these `ModuleTemplates` in the control-plane are not synchronized.
