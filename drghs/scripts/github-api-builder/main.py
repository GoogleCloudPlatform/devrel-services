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

import yaml
import requests
import json
from typing import Dict

def generate():
    paths = create_paths()

    template_yaml = {}
    with open('magic-github-proxy.yaml', 'r') as f:
        template_yaml = yaml.load(f, Loader=yaml.FullLoader)

    template_yaml['paths'].update(paths)

    # Dump the yaml out to stdout
    # setting default_flow_style=False endables the pretty
    # block syntax as opposed to the {} style syntax
    print(yaml.dump(template_yaml, default_flow_style=False))

def create_paths() -> Dict:
    paths = {}
    for i in range(1, 16):
        key = ''.join(["/*" for x in range(0, i)])
        paths[key] = make_verbs(i)

    return paths

def make_verbs(path: str) -> Dict:
    d = {}
    for verb in ['get','head','patch','put','post','delete']:
        d[verb] = make_operation(path, verb)
    return d

def make_operation(path: str, oper: str) -> Dict:
    return {
        'operationId': '{}:{}'.format(path, oper),
        'produces': [ 'application/json' ],
        'security': [ { 'api_key': [] } ],
        # The responses block is REQUIRED, but as far as I
        # can tell, there is no effect on the API or proxy.
        # e.g. if your user credentials are unauthorized,
        # at GitHub the proxy will still return 401 even
        # though this spec says it can only return 200.
        'responses': {
            '200':{
                'description': 'A proxy response from GitHub'
            }
        }
    }

if __name__ == '__main__':
    generate()
