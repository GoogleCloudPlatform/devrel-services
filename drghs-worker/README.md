# drghs-worker

`drghs-worker` is a soft fork of the [maintner](https://github.com/golang/build/tree/master/maintner) service
written by the [Go](https://golang.org) team.

Whereas `maintner` is a single, monlithic service (and therfore mutation log)
that records all Issues and Pull Requests for a set of repositories,
`drghs-worker` has a single-tenancy theory. Each repository is given
a single worker to read and check mutations. This allows the application to scale better and handle transient failures in a much
more graceful manner.

## Main Components

### maintner-sprvsr

This process is a "supervisor" to the rest of the cluster. It reads
a list of repositories to track from a [Cloud Storage](https://cloud.google.com/storage/) bucket
Then interacts with the Kubernetes API (within the cluster) to dynamically
add and delete `Deployment`s and `Service`s for each repository listed in the file.

> Note: because this service runs in the cluster the service account it runs as must have permissions to edit and delete Deployments and Services.

### maitntner-rtr

This process is a reverse proxy that takes the incoming request, parses out
the Owner and Repository from the request and proxies it to the `Service` in the cluster that is responsible for the repository

### cmd

This is the "main" process that leverages the `corpus` from `maintner` and syncrhonizes the Issues and Pull Requests from GitHub and exposes the API to query them.

## Other tools

### maintmigrate

This tool is used to take a mutation source in Cloud Storage, and create a subset
of it by reading the source into memory and applying filters to it.

> This was originally used to take our monolithic mutation source and split it to a single-tenancy model.

### maint-bucket-migrate

This process is used to do a "one off" migration of a set of mutation logs from one set of buckets to another.
