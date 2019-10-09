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

# magic-github-proxy

This application uses the wonderful
[magic-github-proxy](https://github.com/theacodes/magic-github-proxy)
originally authored by [Thea Flowers](https://github.com/theacodes) to add an
extra ACL layer in front of the GitHub API by using Asymmetric Cryptography and
JSON Web Tokens. 

# Setup

## Making the keys

magic-github-proxy needs the use of JSON Web Tokens and Aysymmetric
Cryptography in order to function. The public key, private key and 
certificate are all stored in Google Cloud Storage and are generated 
by running:

``` bash
$ openssl req -x509 -nodes -newkey rsa:4096 -keyout private.pem -out public.x509.cer
$ openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

The first command creates the public key and the x509 certificate and the 
second command creates a public key from the private key.

If these keys are lost, rotated, or deleted, the users (the ones who have
created Json Web Tokens) **must** regenerate their tokens, as
magic-github-proxy will no longer be able to decrypt the tokens.

## Encrypting the Keys with CloudKMS

### Create a Cloud KMS Keyring

```bash
gcloud kms keyrings create magic-github-proxy --location global
```

### Create the Symmetric key

```bash
gcloud kms keys create enc-at-rest --location global \
  --keyring magic-github-proxy --purpose encryption
```

This uses a symmetric key as the Asymmetric Keys cannot encrypt the
private key & certificate (they are too big).

### Encrypt the files

First, encrypt the certificate

```bash
gcloud alpha kms encrypt \
  --key enc-at-rest \
  --keyring magic-github-proxy \
  --location global \
  --plaintext-file ./public.x509.cer \
  --ciphertext-file ./public.x509.cer.enc
```

Then the private key

```bash
gcloud alpha kms encrypt \
  --key enc-at-rest \
  --keyring magic-github-proxy \
  --location global \
  --plaintext-file ./private.pem \
  --ciphertext-file ./private.pem.enc
```

### Upload the files to a Cloud Bucket

1. Create  a cloud storage bucket named magic-github-proxy-dev
1. Upload `public.x509.cer.enc` and `private.pem.enc` to the bucket.

## Giving the Kubernetes Deployment Appropriate Permissions

As the Deployment in the cluster will read from Cloud Storage, and perform an
asymmetric decrypt operation, it will need to run as a special Service Account
which has permissions to read from the bucket, as well as perform decryption
operations on the key we just created

1. Create a Service Account for magic-github-proxy
1. Give it Roles KMS Decrypt
1. Go to the Key we wish to use to Decrypt
1. Give the service account permission to decrypt using that key
1. Go to the magic-github-proxy-dev storage bucket and give the 
Service Account appropriate permissions. (Storage Legacy Bucket Reader, 
Storage Legacy Object Reader)

### Create the Kubernetes Secret for the Service Account

```bash
kubectl create secret generic service-account-magic-github-proxy
--from-file=key.json
```

## Deploying

Once you have done this pre-work, run `make deploy`, confirm the variables are
set appropriately (and override them with environment variables if need be),
and hit "enter".
