# Configuration

Configuration can be done by config yaml file environment variable (without predefined queries/views)

Default it will look for a config.yaml in the current directory,
but you can pass the -config parameter to define the location of the config file

## Options
| Option                    | Environment Variable      | Default   | Type   | Description                                          |
|---------------------------|---------------------------|-----------|--------|------------------------------------------------------|
| listen                    | LISTEN                    | localhost | string | Listen host/ip                                       |
| port                      | PORT                      | 5000      | int    | Listen to port                                       |
| puppetdb.host             | PUPPETDB_HOST             | localhost | string | Address of puppetdb                                  |
| puppetdb.port             | PUPPETDB_PORT             | 8080      | int    | Port of puppetdb                                     |
| puppetdb.tls              | PUPPETDB_TLS              | false     | bool   | Communicate over tls with puppetdb                   |
| puppetdb.tls_ignore       | PUPPETDB_TLS_IGNORE       | false     | bool   | Ignore validation of tls certificate                 |
| puppetdb.tls_ca           | PUPPETDB_TLS_CA           |           | string | Path to ca cert file for puppetdb                    |
| puppetdb.tls_key          | PUPPETDB_TLS_KEY          |           | string | Path to client key file for puppetdb                 |
| puppetdb.tls_crt          | PUPPETDB_TLS_CERT         |           | string | Path to client cert file for puppetdb                |
| queries                   |                           |           | array  | predefined queries (see query table)                 |
| views                     |                           |           | array  | predefined views (see view table)                    |
| trusted_proxies           | TRUSTED_PROXIES           |           | array  | List of trusted proxies (env var is space seperated) |
| puppetca.host             | PUPPETCA_HOST             |           | string | Address of Puppet CA server (optional)               |
| puppetca.port             | PUPPETCA_PORT             | 8140      | int    | Port of Puppet CA server                             |
| puppetca.tls              | PUPPETCA_TLS              | true      | bool   | Use TLS for Puppet CA communications                 |
| puppetca.tls_ignore       | PUPPETCA_TLS_IGNORE       | false     | bool   | Ignore validation of TLS certificate                 |
| puppetca.tls_ca           | PUPPETCA_TLS_CA           |           | string | Path to CA cert file for Puppet CA                   |
| puppetca.tls_key          | PUPPETCA_TLS_KEY          |           | string | Path to client key file for Puppet CA                |
| puppetca.tls_crt          | PUPPETCA_TLS_CERT         |           | string | Path to client cert file for Puppet CA               |
| puppetca.readonly         | PUPPETCA_READONLY         | true      | bool   | Whether to allow signing / revoking / cleaning certs |
| puppetca.deactivate_nodes | PUPPETCA_DEACTIVATE_NODES | false     | bool   | Also deactivate node in PuppetDB with revoke / clean |


### predefined Queries
| Option      | Type   | Description               |
|-------------|--------|---------------------------|
| description | string | description of your query |
| query       | string | PQL query string          |
|             |        |                           |

### predefined Views
| Option                | Type   | Description                                       |
|-----------------------|--------|---------------------------------------------------|
| name                  | string | description of your query                         |
| facts                 | array  | facts that should be shown (see view facts table) |
| default_rows_per_page | number | how many rows should be default shown in ui       |


### predefined Views - Facts
| Option   | Type   | Description                                                                         |
|----------|--------|-------------------------------------------------------------------------------------|
| name     | string | column name that should be shown                                                    |
| fact     | string | which fact should be shown (can be . seperated for lower level (like networking.ip) |
| renderer | string | (optional) there are some renderer like hostname, certname, or os_name              |

### Puppet CA

The Puppet CA query and management functionality is enabled by configuring a `puppetca.host`; if this variable is left empty the
functionality will not be available. Furthermore, the CA functionality is read-only by default, unless `puppetca.readonly` is set to `false.
This means that the list of requested, signed, and revoked certificates can be viewed from the web interface when enabled, but no changes
can be made via the web interface unless the configuration setting is changed.

To use the CA functionality you will need a key and certificate with the `pp_cli_auth` extension. This can be obtained using the
`--ca-client` option of the `puppetserver ca generate`, for example. A standard node certificate lacks the required extension in order to
use the Puppet CA API within OpenVox Server / Puppet Server. The same key and certificate can be used for both PuppetDB and Puppet CA,
however.

There is also a function (enabled by setting `puppetca.deactivate_nodes` to true) which will send PuppetDB a "deactivate node" request when
revoking or cleaning a node certificate. This performs the equivalent of `puppet node deactivate $CERTNAME` after the certificate is revoked
or cleaned, which may be useful in environments where the CLI tools are not easily available.

## YAML Example

```yaml
---
listen: 127.0.0.1
port: 5000
trusted_proxies:
 - 127.0.0.1

puppetdb:
  host: localhost
  port: 8081
  tls: true
  tls_ignore: false
  tls_ca: /path/to/ca.crt
  tls_key: /path/to/cert.key
  tls_cert: /path/to/cert.crt

queries:
  - description: Inactive Nodes
    query: nodes[certname] { node_state = "inactive" }

views:
  - name: 'Inventory'
    facts:
      - name: 'Hostname'
        fact: 'trusted'
        renderer: 'hostname'
      - name: 'IP Address'
        fact: 'networking.ip'
      - name: 'OS'
        fact: 'os'
        renderer: 'os_name'
      - name: 'Kernel Version'
        fact: 'kernelrelease'
      - name: 'Puppet Version'
        fact: 'puppetversion'

puppetca:
  host: localhost
  port: 8140
  tls: true
  tls_ignore: false
  tls_ca: /path/to/ca.crt
  tls_key: /path/to/cert.key
  tls_cert: /path/to/cert.crt
  readonly: true
  deactivate_nodes: false
```
