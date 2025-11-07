# obsolete-packages

[![Release](https://img.shields.io/github/release/omilevskyi/obsolete-packages.svg)](https://github.com/omilevskyi/obsolete-packages/releases/latest)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://raw.githubusercontent.com/omilevskyi/obsolete-packages/refs/heads/main/LICENSE)
[![Build](https://github.com/omilevskyi/obsolete-packages/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/omilevskyi/obsolete-packages/actions/workflows/build.yml)
[![Powered By: GoReleaser](https://img.shields.io/badge/Powered%20by-GoReleaser-blue.svg)](https://goreleaser.com/)

This tool finds obsolete FreeBSD local packages by making
as accurate as possible port version, revision, and epoch comparisons.
It may optionally delete findings. If the packages directory is not specified,
then it is determined by a series of "make" utility runs.

## Verification of the authenticity

```sh
export VERSION=0.0.5
cosign verify-blob \
  --certificate-identity https://github.com/omilevskyi/obsolete-packages/.github/workflows/release.yml@refs/heads/main \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate https://github.com/omilevskyi/obsolete-packages/releases/download/v${VERSION}/obsolete-packages_${VERSION}.sha256.pem \
  --signature https://github.com/omilevskyi/obsolete-packages/releases/download/v${VERSION}/obsolete-packages_${VERSION}.sha256.sig \
  https://github.com/omilevskyi/obsolete-packages/releases/download/v${VERSION}/obsolete-packages_${VERSION}.sha256
Verified OK
```
