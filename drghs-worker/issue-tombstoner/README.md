# issue-tombstoner

## Why?

Sometimes the `sweeper` job can be behind, this provides quick one-off access to
quickly tombstone a set of issues on a repository

## How to Use

1. Open a terminal
1. Ensure you are using the correct cluster
1. Run `kubectl run tmp-shell --restart=Never --rm -i -tty --image
   gcr.io/${PROJECT_NAME}/issue-tombstoner`. This will get you a running shell
   inside the k8s cluster with the `issue-tombstoner` cli available to it
1. Run `issue-tombstoner --owner=<PROBLEMATIC OWNER> --repo=<PROBLEMATIC REPO>
   <ISSUE_1> <ISSUE_2> ... <ISSUE_N>`

This will make a one-off pod in the cluster named `tmp-shell` with a TTY
attached to it so you can interact with it. Then when you run the
`issue-tombstoner` command it will make a set of RPCs to the Pod responsible
for the owner/repository given in the flags to tell the Pod to tombstone each
issue given in the args list.

## Building
``
docker build -t issue-tombstoner -f ./Dockerfile ../../
``
## Deploying
``
docker tag issue-tombstoner gcr.io/${PROJECT_NAME}/issue-tombstoner

docker push gcr.io/${PROJECT_NAME}/issue-tombtoner
``

