# OpenVox View
[![Build and Release](https://github.com/voxpupuli/openvoxview/actions/workflows/ci.yml/badge.svg)](https://github.com/voxpupuli/openvoxview/actions/workflows/ci.yml)
[![Apache-2 License](https://img.shields.io/github/license/voxpupuli/openvoxview.svg)](LICENSE)

<img src="./ui/public/logo.png" alt="Logo" width="256" height="256">

## Introduction
OpenVox View is a viewer for openvoxdb/puppetdb, inspired by [Puppetboard](https://github.com/voxpupuli/puppetboard).

## Features
- Overview of reports
- Overview of facts
- Overview of nodes
- Predefined views
- Ability to perform multiple queries
- Query history
- Predefined queries
- Puppet CA web interface

## Container
You can build a container with the Containerfile

```bash
podman build -t openvoxview .
```

or for Docker
```bash
docker build -t openvoxview -f Containerfile .
```

## Running OpenVox view via systemd
We provide an example systemd unit including sandboxing in [openvoxview.service.example](./openvoxview.service.example).

Your overall environment will vary, so this is mainly thought as a starting point for your own unit file. Make sure to have a closer look at `ExecStart`, `ReadWritePaths` and `WorkingDirectory`.

## Configuration
See [CONFIGURATION.md](./CONFIGURATION.md)


## Screenshots
### Reports Overview
![Reports Overview](./screenshots/reports.png)

### Node Detail
![Node Detail](./screenshots/node_detail.png)

### Query Execution
![Query Execution](./screenshots/query_execution.png)

### Query History
![Query History](./screenshots/query_history.png)

## Contribution
We welcome you to create issues or submit pull requests. Most important be excellent to each other.

For more infos, see [DEVELOPMENT.md](./DEVELOPMENT.md)

## OpenVox/Puppet Module
There is also a openvox module for deployment of openvoxview see [puppet-openvoxview](https://github.com/voxpupuli/puppet-openvoxview)


## Special Thanks
We extend our gratitude for the remarkable work on [Puppetboard](https://github.com/voxpupuli/puppetboard).
