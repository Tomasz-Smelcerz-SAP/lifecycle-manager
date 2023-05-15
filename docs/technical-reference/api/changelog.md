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


