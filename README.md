# k8s-grpc-client-side-lb

This is a sample code of gRPC client side load balancing on kubernetes.

## Setup

1. Export environment variables.

```console
$ export PROJECT_ID=${YOUR_GCP_PROJECT_ID}
```

2. Create container cluster.

```console
$ make create-cluster
```

3. Install dependencies.

```console
$ dep ensure --vendor-only
```

4. Deploy sample applications.

```console
$ make deploy
```

## Cleanup

```console
$ make delete-cluster
```

## Architecture

- http app: call grpc app function via gRPC
- grpc app: return pod host name

```
           +-----------------------+
           | load balancer service |
           +----------+------------+
                      |
                 +----v-----+
                 | http app |
                 +----+-----+
     /cluster-ip      |        /headless
          +-----------+-----------+
          |                       |
+---------v----------+  +---------v--------+
| cluster IP service |  | headless service |
+---------+----------+  +---------+--------+
          |                       |
          +-----------+-----------+
                      |
       +--------------+---------------+
       |              |               |
  +----v-----+   +----v-----+   +-----v----+
  | gRPC app |   | gRPC app |   | gRPC app |
  +----------+   +----------+   +----------+
```

## How to check client side load balancing

### use HTTP API

#### via cluster IP service

##### without client side LB

`/cluster-ip` returns single host name.

```console
$ watch -n 1 -d curl -s `make external-ip`:8000/cluster-ip
```

##### with client side LB


`/cluster-ip/lb` returns single host name. (LB is not work)

```console
$ watch -n 1 -d curl -s `make external-ip`:8000/cluster-ip/lb
```

#### via headless service

##### without client side LB

`/headless` returns single host name.

```console
$ watch -n 1 -d curl -s `make external-ip`:8000/headless
```


##### with client side LB

`/headless/lb` returns a different host name for each request. (LB is work)

```console
$ watch -n 1 -d curl -s `make external-ip`:8000/headless/lb
```

### use [channelz](https://github.com/grpc/proposal/blob/master/A14-channelz.md) service with [evans](https://github.com/ktr0731/evans)

```console
$ echo '{"start_channel_id":0}' | evans --host `make external-ip` --port 9000 --package grpc.channelz.v1 --service Channelz --call GetTopChannels ../../grpc/grpc/src/proto/grpc/channelz/channelz.proto | jq '.channel[] | {ref, subchannelRef}'
```

You can check the connection information.
`dns:///headless-service:9000` has three connection.

```json
{
  "ref": {
    "channelId": 1,
    "name": "cluster-ip-service:9000"
  },
  "subchannelRef": [
    {
      "subchannelId": 7,
      "name": ""
    }
  ]
}
{
  "ref": {
    "channelId": 2,
    "name": "headless-service:9000"
  },
  "subchannelRef": [
    {
      "subchannelId": 8,
      "name": ""
    }
  ]
}
{
  "ref": {
    "channelId": 3,
    "name": "dns:///cluster-ip-service:9000"
  },
  "subchannelRef": [
    {
      "subchannelId": 11,
      "name": ""
    }
  ]
}
{
  "ref": {
    "channelId": 4,
    "name": "dns:///headless-service:9000"
  },
  "subchannelRef": [
    {
      "subchannelId": 15,
      "name": ""
    },
    {
      "subchannelId": 13,
      "name": ""
    },
    {
      "subchannelId": 14,
      "name": ""
    }
  ]
}
```
