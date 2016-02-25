# Security Overview

Kocho uses the following golang libraries to interact with certain APIs:

- [aws/aws-sdk-go](https://godoc.org/github.com/aws/aws-sdk-go) to interact with the AWS API
- [nlopes/slack](https://godoc.org/github.com/nlopes/slack) library to interact with the Slack API
- [crackcomm/cloudflare](https://godoc.org/github.com/crackcomm/cloudflare) to interact with the Cloudflare API

To interact with the Etcd discovery, plain HTTP requests are issued.

## Risks

We recommend to run `kocho` within a separate secured network
with limited access by non-authorized users. This way the lack of
encryption as well as the general protocol issues are less critical.
