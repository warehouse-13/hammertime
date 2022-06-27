## Hammertime

A very basic CLI tool for interacting with [flintlock](https://github.com/weaveworks/flintlock) servers.

(Those of you who know your archaic gun mechanisms will understand the name and no doubt
find it hilarious.) (You are welcome.)

This started off as a toy I made quickly to test live flintlock servers.
I have kept it around because it makes working with flintlock very straightforward.

It is currently being heavily refactored.


<!--
To update the TOC, install https://github.com/kubernetes-sigs/mdtoc
and run: mdtoc -inplace README.md
-->

<!-- toc -->
- [Flintlock?](#flintlock)
- [Versioning](#versioning)
- [Installation](#installation)
- [Usage](#usage)
- [Development](#development)
  - [Testing](#testing)
<!-- /toc -->


### Flintlock?

[Flintlock](https://github.com/weaveworks/flintlock) is a service to manage MicroVMs
on bare-metal.

MicroVMs are, as they sound, smaller VMs. Unlike regular VMs, which generally must
be prepared to run any kernel, OS, environment with any number of features that a user
may end up needing, MicroVMs are stripped down for a purpose. They provide a smaller
subset of virtualisation tailored for a specific task (in the case of Flintlock, this is to
run [Kubernetes](https://kubernetes.io/) nodes). This means they are smaller and "lighter"
to run. In a best of both worlds thing: they provide the speed and lower resource allocation
of container, and the security of full VMs.

### Versioning

Check the release notes for each release to find Flintlock compatibility.
Both Flintlock and this tool are in alpha development, thus the API is likely
to change often until v1.

### Installation

1. Build from source:
   ```bash
   git clone <this or your fork>
   cd hammertime
   make build
   ```

2. Get a [released binary](https://github.com/warehouse-13/hammertime/releases)

3. Install with go: `go install github.com/warehouse-13/hammertime/releases@latest`

Alias to `ht` if you like.

### Usage

4 commands, very few configuration options. Each command simply spits out the response
as JSON so you can pipe to `jq` or whatever as you like.

```bash
# see all options
hammertime --help

# create 'mvm0' in 'ns0' (take note of the UID after creation)
hammertime create

# get 'mvm0' in 'ns0'
hammertime get

# get just the state of 'mvm0' in 'ns0' see below
hammertime get -s

# get
hammertime get -i <UUID>

# get all mvms in `ns0`
hammertime list --namespace ns0

# delete 'bar' from 'foo'
hammertime delete --namespace foo --name --bar

# delete
hammertime delete -i <UID>
```

The name and namespace are configurable, as is the GRPC address.
There is the option to create with an SSH key.
You can also pass a full json configfile to `create`, `get` and `delete` if you want to override
everything (see [example.json](example.json)).

Run `hammertime --help` for all options.

### Development

For a list of all make commands, run `make help`.

You will need Go version >= `1.18`.

#### Testing

Our tests use [`ginkgo` v2](https://onsi.github.io/ginkgo/). To install v2 run:

```bash
go install github.com/onsi/ginkgo/v2/ginkgo
ginkgo version //should print out "Ginkgo Version 2.0.0"
```

Tests can then be run with `make test`.

All new code must be submitted with at least unit tests. All new or changed core features
must come with a matching or updated integration test.
