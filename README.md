# Kocho

[![Build Status](https://api.travis-ci.org/giantswarm/kocho.svg)](https://travis-ci.org/giantswarm/kocho) [![](https://godoc.org/github.com/giantswarm/kocho?status.svg)](http://godoc.org/github.com/giantswarm/kocho) [![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)

Kocho provides a set of mechanisms to bootstrap AWS nodes that must follow a
specific configuration with CoreOS. It sets up fleet meta-data, and patched
versions of fleet, etcd, and docker when using
[Yochu](https://github.com/giantswarm/yochu).

## Getting Kocho

Download the latest release: https://github.com/giantswarm/kocho/releases/latest

Clone the git repository: https://github.com/giantswarm/kocho.git

Download the latest docker image from here: https://hub.docker.com/r/giantswarm/kocho/

## Running Kocho

```
./kocho help
```

## Further Steps

Check more detailed documentation: [docs](docs)

Check code documentation: [godoc](https://godoc.org/github.com/giantswarm/kocho)

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/#!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/kocho/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

Kocho is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

## Origin of the Name

`kocho` (こちょう[蝴蝶] pronounced "ko-cho") is Japanese for butterfly.
