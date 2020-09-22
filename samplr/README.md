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

## Testing

To test the samplr code `cd` into the samplr directory and run

```bash
go test -v -race ./...
```

This will run the tests in the `samplr` directory as well as all subdirectories in it, recursively.

### Integration Tests

The integration tests take a bit longer to run, and require network access, so they require a special
command line option to run: `-tags integration`

```bash
go test -v -tags integration ./...
```

This uses the Go `tag` feature, which is enabled by putting a comment in a file of form

```go
// +build TAGNAME
```

## Deploying

The `Makefile` has several variables that can be overridden via environment variables
in order to customize how the deployment occurs.

```makefile
# The (gcloud) test cluster that is being worked against
GCP_CLUSTER_NAME ?= devrel-services
GCP_CLUSTER_ZONE ?= us-central1-a
# The service account to run as
SERVICE_ACCOUNT_SECRET_NAME ?= service-account-maintnerd
# Bucket settings for Repositories
GCS_BUCKET_NAME ?= devrel-dev-settings
REPOS_FILE_NAME ?= public_repos.json
```

These defaults are largely the same for dev and prod, but there are noteable differences

### Deploy to DEV

```bash
export GCP_CLUSTER_NAME=devrel-dev-cluster
make deploy
```

### Deploy to PROD

```bash
export GCS_BUCKET_NAME=devrel-prod-settings
make deploy
```

### Notes

The Deployment process is done via Cloud Build.

The Deployment process is also done using the source code on your machine, as
such local, potentially uncommitted or unreviewed changes may be pushed.

## Debugging

### Finding Problematic Deployments

If you know that repository `bar` in organization `foo` is experiencing problems,
finding the deployment that is responsible for handling that Repository can be
done by running:

```bash
kubectl deployments list -l owner=foo,repository=bar,samplr-sprvsr-autogen=true
```

That should return the singular Deployment responsible for that Repository.

Head over to the GKE Cloud Console to do further inspection and analysis

### Restarting Problematic Pods

There are two ways to "restart" a problematic pod.

#### Scaling

This is the "more correct" way to restart them, which is to scale
the Deployment to 0, then re-scale it to one.

```bash
kubectl scale deployment --replicas=0 {DEPLOYMENT_NAME}
```

Wait for the scale to complete. Then run

```bash
kubectl scale deployment --replicas=1 {DEPLOYMENT_NAME}
```

#### Deleting Pods

Because the replicas for a deployment is set to 1, simply 
deleting the pod will cause Kubernetes to create a new one
to satisfy the desired state of "1 Replicas".

Get pod name

```bash
kubectl get pods -l owner=foo,repository=bar,samplr-sprvsr-autogen=true
```

Delete pod

```bash
kubectl delete pod {POD_NAME}
```

### Running Locally

When running locally, its usually best to run samplr in a container (though it is
possible to run it without it)

To build the images, `cd` into the samplr directory and run `make build` to build
the Docker images locally, then run

```bash
docker run -p 3009:3009 -it samplrd:dev samplrd --owner=foo --repository=bar
```

In order to run an instance of `samplr` in a container on your local machine.
This command also forwards port 3009 on your machine to 3009 on the container's
instance, which allows you to use tools such as BloomRPC in order to inspect the
state of the container.

If you are debugging issues with history parsing, it might be useful to mount
your `/tmp/samplr` directory to the `/tmp` directory in the container in order
to debug how the git repositories are being parsed.

```bash
docker run -p 3009:3009 -v /tmp/samplr:/tmp -it samplrd:dev samplrd --owner=foo --repository=bar
```

A useful tool to run while the container is running is the `docker stats` command.

```bash
docker stats
```

This will bring up a TUI which displays statistics over your running containers.
For samplr, the most interesting (and important) is the RAM and memory usage.

This is best run in a seperate window or `tmux` session.
