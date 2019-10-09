#!/usr/bin/python
#
# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
# Needed for guerrilla patching
from urllib.parse import  parse_qsl, urlencode, urlparse, urlunparse
import flask
import requests
from typing import Tuple, List
# Dont need a req.txt for this as they are in mghp
from cryptography import x509
from cryptography.hazmat import backends
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives.asymmetric import padding
import google.auth.crypt
# Manually installed
from google.cloud import storage
from google.cloud import kms_v1
# Magic Proxy
from magicproxy import proxy
from magicproxy import magictoken
from magicproxy import queries
from magicproxy import headers as headers_utils

_BACKEND = backends.default_backend()


def decrypt_symmetric(project_id, location_id, key_ring_id, crypto_key_id,
                      ciphertext) -> str:
    """Decrypts input ciphertext using the provided symmetric CryptoKey."""

    # Creates an API client for the KMS API.
    client = kms_v1.KeyManagementServiceClient()

    # The resource name of the CryptoKey.
    name = client.crypto_key_path_path(project_id, location_id, key_ring_id,
                                       crypto_key_id)
    # Use the KMS API to decrypt the data.
    response = client.decrypt(name, ciphertext)
    return response.plaintext

def download_blob(bucket_name, source_blob_name) -> str:
    """Downloads a blob from the bucket."""
    storage_client = storage.Client()
    bucket = storage_client.get_bucket(bucket_name)
    blob = bucket.blob(source_blob_name)

    return blob.download_as_string()

def keys_from_strings(pri, cer) -> magictoken.Keys:
    private_key = serialization.load_pem_private_key(
                pri, password=None, backend=_BACKEND
            )
    private_key_signer = google.auth.crypt.RSASigner.from_string(
                pri
            )
    certificate = x509.load_pem_x509_certificate(cer, _BACKEND)
    public_key = certificate.public_key()

    key = magictoken.Keys( private_key=private_key,
            private_key_signer=private_key_signer,
            public_key=public_key,
            certificate=certificate,
            certificate_pem=cer,
        )

    return key

# This is to add pull request #5 support: https://github.com/theacodes/magic-github-proxy/pull/5
def guerrilla_clean_path_queries(query_params_to_clean, path) -> str:
    parts = urlparse(path)
    if not parts.query:
        return path
    queries = parse_qsl(parts.query, keep_blank_values=True, strict_parsing=True)
    cln = [q for q in queries if q[0] not in query_params_to_clean]
    return urlunparse((parts.scheme, parts.netloc, parts.path, parts.params, urlencode(cln), parts.fragment))

# This fixes an issue with headers being both the name of a variable and the module. (PR incoming)
def guerrilla_proxy_request(
    request: flask.Request, url: str, headers=None, **kwargs
) -> Tuple[bytes, int, dict]:
    clean_headers = headers_utils.clean_request_headers(request.headers, proxy.custom_request_headers_to_clean)

    if headers:
        clean_headers.update(headers)


    # Make the GitHub request
    resp = requests.request(
        url=url,
        method=request.method,
        headers=clean_headers,
        params=dict(request.args),
        data=request.data,
        **kwargs,
    )

    response_headers = headers_utils.clean_response_headers(resp.headers)

    return resp.content, resp.status_code, response_headers

if __name__ == "__main__":
    parser = argparse.ArgumentParser("Magic GitHub Proxy")
    parser.add_argument('--port', metavar='P', type=int, nargs='?', default=5000,
                   help='the port to listen on')
    parser.add_argument('--debug',metavar='D', type=bool, nargs='?', default=False,
                        help='run in debug mode or not')
    parser.add_argument('--project-id', type=str, required=True, help='the gcp project id')
    parser.add_argument('--kms-location', type=str, required=True, help='kms key location')
    parser.add_argument('--kms-key-ring', type=str, required=True, help='kms key ring')
    parser.add_argument('--kms-key', type=str, required=True, help='kms key name')
    parser.add_argument('--bucket-name', type=str, required=True, help='bucket name')
    parser.add_argument('--pri', type=str, required=True, help='name of private key')
    parser.add_argument('--cer', type=str, required=True, help='name of certificate')

    args = parser.parse_args()

    pri_enc = download_blob(args.bucket_name, args.pri)
    cer_enc = download_blob(args.bucket_name, args.cer)

    pri_dec = decrypt_symmetric(args.project_id, args.kms_location, args.kms_key_ring, args.kms_key, pri_enc)
    cer_dec = decrypt_symmetric(args.project_id, args.kms_location, args.kms_key_ring, args.kms_key, cer_enc)

    # guerrilla Patching
    queries.clean_path_queries = guerrilla_clean_path_queries
    proxy._proxy_request = guerrilla_proxy_request
    # End guerrilla Patching

    proxy.keys = keys_from_strings(pri_dec, cer_dec)

    proxy.query_params_to_clean = ['key']

    proxy.app.run(host='0.0.0.0',port=args.port,debug=args.debug)

