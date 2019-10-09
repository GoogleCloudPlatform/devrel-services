Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
# GitHub API Builder

This project came about for a number of reasons. 

## 1. Cloud Endpoints (and OpenAPI) does not allow wildcards in their path specification that include "/". 

Consider the following pseudo specification:

```yaml
/magictoken:
    get:
    post:
/*:
    get:
    put:
    post:
    delete:
```

Making a call to:

*  "/magictoken" 
* "/foo"
* "/bar"
* "/baz"

will succeed.

But making a call to
* "/foo/bar"

Will fail with a "Method Does Not Exist" error.


## 2. GitHub does not provide an OpenAPI-compliant specification for their APIs and Services.

## What does this do then?

This takes our base specification `magic-github-proxy.yaml.template` and appends 15 paths. 
Each of those paths are "/*" repeated n times.

```yaml
/*:
    get:
    put:
    post:
    delete:
/*/*:
    get:
    put:
    post:
    delete:
/*/*/*:
    get:
    put:
    post:
    delete:
# etc.
```
This bypasses the wildcard problem at the cost of losing resolution in our Endpoints portal in Pantheon.
We would know that someone called a github api that was 12 layers deep, but not what that call was.

## Past options

Previously, we were running this script to generate a custom OpenAPI spec based off of 
a community driven effort to document the GitHub api (found at apis.guru). 

I could not find the source code that scraped the GitHub API to create the OpenAPI 
specification, and it appears as though the specification has not been updated
in the past two years, so we migrated to this.

# To Run

From this directory:

```bash
 docker build -t github-builder . && docker run -it github-builder > out.yaml
 ```