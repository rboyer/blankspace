blankspace
----------

Run:

```
$ docker run --name bs --net=host -d rboyer/blankspace -name my-name
```

gRPC:

```
$ grpcurl -plaintext localhost:8079 list
blankspace.Server
grpc.reflection.v1alpha.ServerReflection

$ grpcurl -plaintext localhost:8079 describe blankspace.Server
blankspace.Server is a service:
service Server {
  rpc Describe ( .blankspace.DescribeRequest ) returns ( .blankspace.DescribeResponse );
}

$ grpcurl -d '{}' -plaintext localhost:8079 blankspace.Server.Describe
{
  "name": "my-name"
}
```

TCP:

```
$ telnet localhost 8078
Trying ::1...
Connected to localhost.
Escape character is '^]'.
describe
my-name
describe
my-name
quit
Connection closed by foreign host.
```
