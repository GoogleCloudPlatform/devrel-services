# maint-butcket-migrate

This is a utility tool to migrate a set of mutation logs from one bucket to 
another.

This works by reading a file describing the set of repositories to migrate,
then calculating the "from" buckets and "to" buckets.

This then uses the kubernetes api to stop the running "to" maintner instances,
then it clears out the "to" bucket and copies the contents of the "from" bucket
to the "to" bucket.

Once all repositories have been migrated, the "supervisor" pod is restarted
in order to restart all the deployments we deleted during the migration.

> NOTE: The "from" and "to" buckets can be in different projects

## Usage

`maint-bucket-migrate --file=migrate_repos.json --from-prefix="mtr-b-"
--to-prefix="mtr-p-"`
