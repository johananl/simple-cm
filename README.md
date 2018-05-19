# SimpleCM - A Distributed Configuration Management System

**NOTE: This is a proof of concept and by no means a production-ready tool.**

The idea behind this exercise is to create a simple configuration management tool with built-in
scalability. The system should allow managing the configuration of a large number of remote hosts.

The solution is implemented in Go and uses the standard library almost entirely. The only
exceptions are:

- [github.com/gocql/gocql][5] - used as the database driver.
- [golang.org/x/crypto/ssh][4] - this package is part of the "extended" standard library that is
maintained by the Go community but is not part of the core standard library.

## Design

### General

The high-level design of the system uses **agent-less push mode** configuration management. This
means that configuration changes are sent **from the system to the hosts** without the need to run
an agent on the hosts. A "pull mode" also exists, in which the hosts query the configuration
management system periodically using an agent that runs on each host, then perform configuration
changes locally as needed.

### Components

The solution utilizes two main components: **master** and **worker**.

A master reads the inventory information as well as the desired configuration for each host from
the database, and then sends tasks to one or more workers to be executed against the remote hosts.
It records the results received from the workers and stores them persistently in the database.

A worker receives commands, or "operations", from the master, executes them against the remote host
and reports back the results.

The system currently supports *one master* and an *arbitrary number of workers*.

### Communication

Each worker exposes an **RPC interface** over HTTP, which the master uses to send work to the
workers. During transport, the payload is serialized using `encoding/gob` which is binary-based and
gives very good performance comparing to alternatives.

The *direction* of communication between the master and the workers is *master -> worker*. In terms
of scalability it makes more sense to do it the other way around, that is - to make the workers
"register" with the master and signal that they are ready for work. This is because this way the
master doesn't need to know the network identities of all the workers, which makes the system's
configuration simpler. **However**, unfortunately the [Go RPC library][1] currently does not
support sending requests from the HTTP server to the HTTP client, and using an alternative such as
gRPC was out of the question since it isn't a part of the standard library.

The workers communicate with the remote hosts over **SSH**. SSH allows secure communication over
unsecured networks, and in addition allows interacting with remote hosts easily using shell
commands.

### Workload Distribution

The master uses a simple round-robin algorithm to distribute operations across workers. For every
host, it chooses a worker and sends *all* the operations for that host to the worker. The worker
in turn executes all the operations serially and returns the results synchronously back to the
master.

### Modules and Extensibility

The system can run any operation that can be described using a shell script. This allows a lot of
flexibility when defining new types of operations, or *modules*. Two sample modules were included
in the [modules][6] directory: `file_exists` and `file_contains`. To add new modules, simply add
new scripts to the directory that is referenced by the `--modules-dir` argument of the workers
(default is `/etc/simple-cm/modules`).

Operations typically require *attributes* which allow the user to control the operation's behavior.
Therefore, the modules are stored as Go [templates][7]. The attributes' values are read from the
database for each instance of an operation, and the template is then rendered with the actual
values before the script is executed by a worker on the remote host.

An sample module:

    #!/bin/bash

    if ! grep -q "{{.text}}" {{.path}}; then
        echo "{{.text}}" >> {{.path}}
    fi

This template expects `.text` and `.path` to be interpolated. The rendered script will then check
if the file at `.path` contains the text `.text`, and if not - it will append the text to the file.

NOTE: Operations need to be **idempotent**. That is - they don't need to perform anything if the
relevant resource is already in the desired state. It is the responsibility of the operation's
writer to ensure this is indeed the case.

### Data Model

TODO

## Running

Bring up the DB hosts:

    docker-compose up -d db1 db2 db3

Seed the DB with dummy data (might have to wait a few seconds for the cluster to initialize):

    docker-compose exec db1 cqlsh -e "SOURCE '/tmp/seed.cql'"

Bring up the dummy hosts:

    docker-compose up -d host1 host2 host3 host4 host5

Bring up the workers:

    docker-compose up -d worker1 worker2 worker3

Finally, run the master:

    docker-compose up master


## Caveats, Limitations and Known Issues

### Master-Worker Communication

The master communicates with the workers using [Go's RPC library][1]. This library is [frozen][3]
and is meant to be replaced by technologies such as [gRPC][2]. The library is suffering from
problems such as no TLS support and therefore - cleartext communication over the wire. **However**,
due to the requirement to use only standard library packages, `net/rpc` was chosen as it provides a
simple RPC interface that is more suitable to the task than, say, a REST API (because the
application is operation-oriented rather than resource-oriented).

### Supported Operating Systems

The system is designed to operate against Linux hosts. Other Unix-based operating systems are
likely to work as well, though these haven't been tested.

## Possible Enhancements

- Support multiple masters. This could greatly increase the maximum scale of the system.
- Use an encrypted transport protocol for master-worker communication.
- Paging of results from the database.

[1]: https://golang.org/pkg/net/rpc/
[2]: https://grpc.io/
[3]: https://github.com/golang/go/issues/16844
[4]: https://godoc.org/golang.org/x/crypto/ssh
[5]: https://github.com/gocql/gocql
[6]: modules
[7]: https://golang.org/pkg/text/template/