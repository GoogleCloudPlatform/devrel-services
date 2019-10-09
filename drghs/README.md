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
# Regenerating the Go proto files

1.  Follow the setup instructions:

https://grpc.io/docs/quickstart/go.html#before-you-begin 

1.  Generate the Go proto files:
	```
    protoc --go_out=plugins=grpc:v1/ --proto_path=v1/ --proto_path=../../googleapis --descriptor_set_out=v1/api_descriptor.pb v1/*.proto
    ```
