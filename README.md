# SimpleCM - A Distributed Configuration Management System

This is an exercise and by no means a real-world tool.

## Caveats, Limitations and Known Issues

### Master-Worker Communication

The master communicates with the workers using [Go's RPC library][1]. This library is [frozen][3]
and is meant to be replaced by technologies such as [gRPC][2]. The library is suffering from
problems such as no TLS support and therefore - cleartext communication over the wire. **However**,
due to the requirement to use only standard library packages, `net/rpc` was chosen as it provides a
simple RPC interface without having to build REST APIs etc.

[1]: https://golang.org/pkg/net/rpc/
[2]: https://grpc.io/
[3]: https://github.com/golang/go/issues/16844