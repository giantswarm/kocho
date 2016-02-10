# Security Overview

Kocho uses the golang libraries to interact with certain APIs.

- AWS SDK to interact with the AWS API: https://godoc.org/github.com/aws/aws-sdk-go.
- Slack library to interact with the Slack API: https://godoc.org/github.com/nlopes/slack
- Cloudflare library to interact with the Cloudflare API: https://godoc.org/github.com/crackcomm/cloudflare

Interacting with the Etcd discovery, plain HTTP requests are done.

## Risks

We recommend to run `kocho` within a separate secured network
with limited access by non-authorized users. This way the lack of
encryption as well as the general protocol issues are less critical.
