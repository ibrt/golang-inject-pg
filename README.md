# golang-inject-pg
[![Go Reference](https://pkg.go.dev/badge/github.com/ibrt/golang-inject-pg.svg)](https://pkg.go.dev/github.com/ibrt/golang-inject-pg)
![CI](https://github.com/ibrt/golang-inject-pg/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/ibrt/golang-inject-pg/branch/main/graph/badge.svg?token=BQVP881F9Z)](https://codecov.io/gh/ibrt/golang-inject-pg)

Postgres module for the [golang-inject](https://github.com/ibrt/golang-inject) framework.

### Developers

Contributions are welcome, please check in on proposed implementation before sending a PR. You can validate your changes
using the `./test.sh` script.

```bash
POSTGRES_VERSION='14.2'  POSTGRES_PORT='5433' ./test.sh
```