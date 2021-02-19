### Krius

Lightweight operator to sync Configmap and Secret in your cluster, across all namespaces.


### Functions

- Create a CR Konfig or Sekret with data and respective ConfigMap or Secret will be created
in the all namespaces.

- If the ownerref is set to our CR in the secret of configmap, direct change on those objects
wont make any sense.

- There should be a way to specify that we don't want those resources in the `system` namespaces.