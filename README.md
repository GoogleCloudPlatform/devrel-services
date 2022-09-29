<p style="color: red;">
  UPDATE 2022-09-29: This repo is no longer being maintained or used.
</p>

# devrel-github-services

DevRel GitHub Services is a collection of tools and services aimed
at helping DevRel **do** DevRel. These services in particular focus on the
data aggregation portion of the DevRel process.
Each service will have it's own README with a deeper description.

## drghs

This is a common directory which defines specifications for
[Cloud Endpoints](https://cloud.google.com/endpoints), as well as the
[Protocol Buffer](https://developers.google.com/protocol-buffers/) and
[gRPC](https://grpc.io) definitions.

## drghs-worker

This folder contains an application that is a small fork of the
[maintner](https://github.com/golang/build/tree/master/maintner) service
written by the [Go](https://golang.org) team. The purpose of this service is
to read a list of repositories, synchronize all the Issues and Pull Requests
for each repository and expose an API to query these issues.
It can be thought of as a giant in-memory cache.

## leif

This service takes a set of GitHub repositories and scans them for service-level objective (SLO) rules and exposes an API to query them. It is designed to be deployed to a
[Kubernetes](https://kubernetes.io) cluster.

## samplr

This service takes a set of GitHub repositories and scans them for code
snippets and exposes an API to query them. It is designed to be deployed to a
[Kubernetes](https://kubernetes.io) cluster.
