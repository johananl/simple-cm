# SimpleCM - A Distributed Configuration Management System

This is an exercise and by no means a real-world tool.

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

[1]: https://golang.org/pkg/net/rpc/
[2]: https://grpc.io/
[3]: https://github.com/golang/go/issues/16844