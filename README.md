# MySQL Operator

> **Provenance**: This repository was originally forked from [bitpoke/mysql-operator](https://github.com/bitpoke/mysql-operator) (formerly presslabs/mysql-operator) and was detached into a standalone repository on 2026-07-17, as upstream development had been dormant. It maintains its own release line (v0.7.x, v0.8.x) under the original Apache 2.0 license. All credit for the original design and implementation goes to [Bitpoke](https://www.bitpoke.io/) and the upstream contributors.

MySQL Operator enables highly available MySQL on Kubernetes. It manages the full lifecycle of MySQL clusters — with automated failover via [orchestrator](https://github.com/openark/orchestrator) — and provides out-of-the-box backups, scheduled and on demand, point-in-time recovery, and cloning within and across clusters.

Supported MySQL versions: **5.7** and **8.0** (Percona Server), and **8.4** (MySQL Community) — 8.4 support is maintained in this repository and is not available upstream.

## Deploy the operator

```shell
git clone https://github.com/ksmartdata/mysql-operator.git
cd mysql-operator
helm install mysql-operator deploy/charts/mysql-operator
```

The chart deploys the controller together with an orchestrator cluster. See the chart [README](deploy/charts/mysql-operator/README.md) for the available values.

## Deploy a MySQL cluster

```shell
kubectl apply -f https://raw.githubusercontent.com/ksmartdata/mysql-operator/main/examples/example-cluster-secret.yaml
kubectl apply -f https://raw.githubusercontent.com/ksmartdata/mysql-operator/main/examples/example-cluster.yaml
```

More examples live in the [examples](examples) directory.

## Documentation

The upstream documentation still applies to this operator:

* [Getting started](https://www.bitpoke.io/docs/mysql-operator/getting-started/)
* [Deploy a MySQL cluster](https://www.bitpoke.io/docs/mysql-operator/deploy-mysql-cluster/)
* [Configure backups](https://www.bitpoke.io/docs/mysql-operator/backups/)
* [Recurrent backups](https://www.bitpoke.io/docs/mysql-operator/cluster-recurrent-backups/)
* [Restore a cluster](https://www.bitpoke.io/docs/mysql-operator/backups/#initialize-a-cluster-from-a-backup)
* [Orchestrator access](https://www.bitpoke.io/docs/mysql-operator/orchestrator/)

## Contributing

Issues and pull requests are welcome in this repository.

## License

This project is licensed under the [Apache 2.0](LICENSE) license.
