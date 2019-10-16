# maint-bucket-reduce

This script takes a set of mutation logs from disparate Google Cloud Storage
buckets and moves them to a common bucket, but prefixed with the Owner and
Repository they are tracking

## Usage

```bash
./maint-bucket-reduce --settings-bucket=devrel-maintner-settings \
                      --file=tracked_repos.json \
                      --from-prefix=mtr-d \
                      --to-bucket=maintner-bucket
```


## Example

Consider the following buckets

|  Bucket Name | Repository Tracked |
| ------------ | ------------------ |
| mtr-foo      | org1/repo1         |
| mtr-bar      | org1/repo2         |
| mtr-baz      | org1/repo3         |
| mtr-biz      | org2/repo1         |

And a new bucket `mtr-combined`

When this script is run it will copy the mutation logs in each of those buckets
to a new bucket (specified by the user). The new Bucket, `mtr-combined` will
have the following structure:

```
mtr-combined
+-- org1
    +-- repo1
        + --mutation log files
    +-- repo2
        +-- mutation log files
    +-- repo3
        +-- mutation log files
+-- org2
    +-- repo1
        +-- mutation log files
```


