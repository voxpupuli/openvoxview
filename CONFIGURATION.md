# Configuration

Configuration can be done by config yaml file environment variable (without predefined queries/views)

Default it will look for a config.yaml in the current directory, 
but you can pass the -config parameter to define the location of the config file

## Options
| Option              | Environment Variable | Default   | Type   | Description                                          |
|---------------------|----------------------|-----------|--------|------------------------------------------------------|
| listen              | LISTEN               | localhost | string | Listen host/ip                                       |
| port                | PORT                 | 5000      | int    | Listen to port                                       |
| puppetdb.host       | PUPPETDB_HOST        | localhost | string | Address of puppetdb                                  |
| puppetdb.port       | PUPPETDB_PORT        | 8080      | int    | Port of puppetdb                                     |
| puppetdb.tls        | PUPPETDB_TLS         | false     | bool   | Communicate over tls with puppetdb                   |
| puppetdb.tls_ignore | PUPPETDB_TLS_IGNORE  | false     | bool   | Ignore validation of tls certificate                 |
| puppetdb.tls_ca     | PUPPETDB_TLS_CA      |           | string | Path to ca cert file for puppetdb                    |
| puppetdb.tls_key    | PUPPETDB_TLS_KEY     |           | string | Path to client key file for puppetdb                 |
| puppetdb.tls_crt    | PUPPETDB_TLS_CERT    |           | string | Path to client cert file for puppetdb                |
| queries             |                      |           | array  | predefined queries (see query table)                 |
| views               |                      |           | array  | predefined views (see view table)                    |
| trusted_proxies     | TRUSTED_PROXIES      |           | array  | List of trusted proxies (env var is space seperated) |


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
```
