# DEPRECATED

The ginkgo e2e suite in this directory (inherited from upstream bitpoke) is
deprecated: it is no longer maintained and will not be adapted to new MySQL
versions:

- it has bit-rotted, and its cases use legacy replication syntax that MySQL
  8.4 removed, so it cannot cover new versions;
- it does not actually run in any CI (the "e2e testing" pipeline referencing
  it in the repo-root `.drone.yml` is an upstream bitpoke leftover that
  depends on their GCP secrets and drone environment, and is inert on this
  fork).

Replacement: `test/e2e-chainsaw/` (declarative E2E with kyverno chainsaw), run
by `.github/workflows/e2e-chainsaw.yml` on every PR across the version matrix.
See [test/e2e-chainsaw/README.md](../e2e-chainsaw/README.md).
