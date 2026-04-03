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
| strip_path_prefix         | STRIP_PATH_PREFIX         |           | string | Strip base paths from Puppet code locations          |
| puppetca.host             | PUPPETCA_HOST             |           | string | Address of Puppet CA server (optional)               |
| puppetca.port             | PUPPETCA_PORT             | 8140      | int    | Port of Puppet CA server                             |
| puppetca.tls              | PUPPETCA_TLS              | true      | bool   | Use TLS for Puppet CA communications                 |
| puppetca.tls_ignore       | PUPPETCA_TLS_IGNORE       | false     | bool   | Ignore validation of TLS certificate                 |
| puppetca.tls_ca           | PUPPETCA_TLS_CA           |           | string | Path to CA cert file for Puppet CA                   |
| puppetca.tls_key          | PUPPETCA_TLS_KEY          |           | string | Path to client key file for Puppet CA                |
| puppetca.tls_crt          | PUPPETCA_TLS_CERT         |           | string | Path to client cert file for Puppet CA               |
| puppetca.readonly         | PUPPETCA_READONLY         | true      | bool   | Whether to allow signing / revoking / cleaning certs |
| puppetca.deactivate_nodes | PUPPETCA_DEACTIVATE_NODES | false     | bool   | Also deactivate node in PuppetDB with revoke / clean |


### Authentication

| Option                       | Environment Variable                         | Default               | Type   | Description                                      |
|------------------------------|----------------------------------------------|-----------------------|--------|--------------------------------------------------|
| auth.enabled                 | OPENVOXVIEW_AUTH_ENABLED                     | false                 | bool   | Enable local user authentication                 |
| auth.jwt_secret              | OPENVOXVIEW_AUTH_JWT_SECRET                  |                       | string | Secret for signing JWT tokens (min 32 chars)     |
| auth.access_token_ttl_minutes| OPENVOXVIEW_AUTH_ACCESS_TOKEN_TTL_MINUTES    | 15                    | int    | Access token lifetime in minutes                 |
| auth.refresh_token_ttl_days  | OPENVOXVIEW_AUTH_REFRESH_TOKEN_TTL_DAYS      | 30                    | int    | Refresh token lifetime in days                   |
| auth.db_path                 | OPENVOXVIEW_AUTH_DB_PATH                     | data/openvoxview.db   | string | Path to SQLite database file                     |

When `auth.enabled` is `true`, all API endpoints (except `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/api/v1/version`, and `/api/v1/meta`) require a valid JWT bearer token. If no `jwt_secret` is configured, a random one is generated at startup (tokens will not survive restarts).

To create the first admin user, run:

```
openvoxview --create-admin
```

Users can also be managed via the API endpoints when authenticated:

| Method | Endpoint                    | Description            |
|--------|-----------------------------|------------------------|
| POST   | /api/v1/auth/login          | Login (returns tokens) |
| POST   | /api/v1/auth/refresh        | Refresh access token   |
| POST   | /api/v1/auth/logout         | Revoke refresh token   |
| GET    | /api/v1/auth/me             | Current user profile   |
| GET    | /api/v1/auth/users          | List all users         |
| POST   | /api/v1/auth/users          | Create user            |
| PUT    | /api/v1/auth/users/:id      | Update user            |
| DELETE | /api/v1/auth/users/:id      | Delete user            |

### SAML Authentication (EntraID / ADFS)

| Option                          | Environment Variable                         | Default                                                                    | Type   | Description                                     |
|---------------------------------|----------------------------------------------|----------------------------------------------------------------------------|--------|-------------------------------------------------|
| auth.saml.enabled               | OPENVOXVIEW_AUTH_SAML_ENABLED                | false                                                                      | bool   | Enable SAML 2.0 SSO authentication              |
| auth.saml.idp_metadata_url      | OPENVOXVIEW_AUTH_SAML_IDP_METADATA_URL       |                                                                            | string | URL to IdP federation metadata XML              |
| auth.saml.idp_metadata_file     | OPENVOXVIEW_AUTH_SAML_IDP_METADATA_FILE      |                                                                            | string | Path to local IdP metadata XML file (fallback)  |
| auth.saml.sp_entity_id          | OPENVOXVIEW_AUTH_SAML_SP_ENTITY_ID           |                                                                            | string | SP Entity ID (e.g. https://openvoxview.example.com) |
| auth.saml.sp_acs_url            | OPENVOXVIEW_AUTH_SAML_SP_ACS_URL             |                                                                            | string | Assertion Consumer Service URL                  |
| auth.saml.sp_cert_file          | OPENVOXVIEW_AUTH_SAML_SP_CERT_FILE           |                                                                            | string | Path to SP X.509 certificate (PEM)              |
| auth.saml.sp_key_file           | OPENVOXVIEW_AUTH_SAML_SP_KEY_FILE            |                                                                            | string | Path to SP private key (PEM)                    |
| auth.saml.attr_email            |                                              | http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress         | string | SAML attribute URI for email                    |
| auth.saml.attr_given_name       |                                              | http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname            | string | SAML attribute URI for given name               |
| auth.saml.attr_surname          |                                              | http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname              | string | SAML attribute URI for surname                  |
| auth.saml.attr_display_name     |                                              | http://schemas.microsoft.com/identity/claims/displayname                   | string | SAML attribute URI for display name             |

SAML requires `auth.enabled: true` as a prerequisite. Both local login and SAML SSO can be active simultaneously (recommended for break-glass admin access).

To generate a self-signed SP certificate for SAML:

```
openvoxview --generate-saml-cert
```

This creates `saml-sp.crt` and `saml-sp.key` in the current directory. Point `sp_cert_file` and `sp_key_file` to these files.

SAML API endpoints (public, no auth required):

| Method | Endpoint                         | Description                        |
|--------|----------------------------------|------------------------------------|
| GET    | /api/v1/auth/saml/metadata      | SP metadata XML (for IdP setup)    |
| GET    | /api/v1/auth/saml/login         | Initiates SAML SSO redirect to IdP |
| POST   | /api/v1/auth/saml/acs           | Assertion Consumer Service callback|

After the IdP returns a valid assertion, the user is auto-provisioned in the local database (with `auth_source = 'saml'`) and redirected to the frontend with JWT tokens.

#### EntraID Setup

1. Azure Portal > Enterprise Applications > New Application > Create your own (non-gallery)
2. Single Sign-On > SAML
3. Basic SAML Configuration:
   - Identifier (Entity ID): value of `sp_entity_id`
   - Reply URL (ACS): value of `sp_acs_url`
   - Sign on URL: `https://<host>/api/v1/auth/saml/login`
4. Copy the **App Federation Metadata Url** and use it as `idp_metadata_url`
5. Assign users/groups

#### ADFS Setup

1. ADFS Management > Relying Party Trusts > Add
2. Import from URL: `https://<host>/api/v1/auth/saml/metadata`
3. Add claim rules for email, given name, surname, and display name

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

auth:
  enabled: true
  jwt_secret: "change-me-to-a-long-random-string-min-32-chars"
  access_token_ttl_minutes: 15
  refresh_token_ttl_days: 30
  db_path: "data/openvoxview.db"

  saml:
    enabled: true
    idp_metadata_url: "https://login.microsoftonline.com/<tenant-id>/federationmetadata/2007-06/federationmetadata.xml"
    sp_entity_id: "https://openvoxview.example.com"
    sp_acs_url: "https://openvoxview.example.com/api/v1/auth/saml/acs"
    sp_cert_file: "/etc/openvoxview/saml-sp.crt"
    sp_key_file: "/etc/openvoxview/saml-sp.key"

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
