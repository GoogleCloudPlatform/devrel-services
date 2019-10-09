# samplr

`samplr` is a service designed to take a set of GitHub repositories,
scan their history, and produce a set of `Snippets` which are
exposed over an API.

An example snippet:

``` csharp
// [ START helloworld_snippet ]
public static void Main(string[] args) {
    System.Console.WriteLine("Hello world!");
}
// [ END helloworld_snippet]
```

## Main Components

### samplrd

This service clones a GitHub repository, iterates through the commits on the `master` branch,
and exposes a [gRPC](https://grpc.io) API to query those samples. Once instantiated, it will
periodically check the repository for new commits and, in the event there are new commits,
will update it's set of snippets.

### samplr-sprvsr

This process is a "supervisor" to the rest of the cluster. It reads
a list of repositories to track from a [Cloud Storage](https://cloud.google.com/storage/) bucket
Then interacts with the Kubernetes API (within the cluster) to dynamically
add and delete `Deployment`s and `Service`s for each repository listed in the file.

> Note: because this service runs in the cluster the service account it runs as must have
> permissions to edit and delete Deployments and Services.

### samplr-rtr

This service is the "entrypoint" to the cluster. It is secured behind Cloud Endpoints,
and exposes a gRPC reverse-proxy which inspects the incoming request and forwards it
to the `Service` in the cluster which is responsible for handling that

## Other Tools

### samplrctl

This is a command line tool using [Cobra Commands](https://github.com/spf13/cobra) to
query a repository checked out on disk for `Snippets`

e.g.

```bash
samplrctl snippets list /tmp/local-repository
```
