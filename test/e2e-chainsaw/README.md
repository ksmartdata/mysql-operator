# chainsaw E2E (version-compatibility gate)

Goal: every operator PR must pass the five scenarios — create, my.cnf golden,
replication, failover, config update — on 5.7.44 / 8.0.37 / (8.4.9 once 8.4
support lands), replacing the deprecated `test/e2e/` (ginkgo) suite.
Version-matrix rationale: the operator code paths only have three equivalence
classes — `Major == 5`, `8.0.x (< 8.4)` and `>= 8.4` — so we pick one
production version per class; 8.0.31 and 8.0.37 take the same branch, so
8.0.31 stays out of the PR gate (optional nightly).

## Layout

```
test/e2e-chainsaw/
├── config/chainsaw-configuration.yaml   # global chainsaw config (serial, timeouts, global failure diagnostics)
├── values/mysql-<version>.yaml          # per-version parameters (version/image/golden path), injected via --values
├── golden/                              # my.cnf golden files (generated on the baseline in step 0, see below)
└── tests/
    ├── _shared/                         # shared bits: CR/assert templates, create-cluster step template, sh helpers
    ├── 01-create-replication/           # scenarios 1+2+3: create, my.cnf golden, replication
    ├── 02-failover/                     # scenario 4: force-kill master (crash failover) -> promotion -> old master rejoins
    └── 03-config-update/                # scenario 5: patch mysqlConf -> rollout -> Ready again and setting live
```

Three chainsaw Tests cover the five scenarios of the 06 design doc
(scenarios 1/2/3 share one cluster to save CI time).

## Running locally

```bash
# 1. Build the operator image (must build from the local checkout:
#    arm64/images/mysql-operator/Dockerfile re-downloads the main
#    branch sources from GitHub, so it would not test local changes)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/mysql-operator_linux_amd64 ./cmd/mysql-operator
docker build -t mysql-operator:e2e -f hack/development/Dockerfile.operator .

# 2. Build the orchestrator image (Percona orchestrator fork built from
#    source; the gate must test the orchestrator shipped by the branch)
docker build -t mysql-operator-orchestrator:e2e -f images/mysql-operator-orchestrator/Dockerfile .

# 3. kind cluster + deploy the operator (same script as CI; override images
#    with OPERATOR_IMAGE_* / ORCHESTRATOR_IMAGE_* env vars)
kind create cluster --name chainsaw
kind load docker-image mysql-operator:e2e --name chainsaw
kind load docker-image mysql-operator-orchestrator:e2e --name chainsaw
./hack/e2e-chainsaw-setup.sh

# 4. Run a single version
chainsaw test --test-dir test/e2e-chainsaw/tests \
  --config test/e2e-chainsaw/config/chainsaw-configuration.yaml \
  --values test/e2e-chainsaw/values/mysql-5.7.44.yaml
```

## Step 0: generate goldens (must happen before code changes)

The goldens are the mechanical enforcement of the "zero my.cnf diff for
existing versions" rule: the 5.7.44 / 8.0.37 goldens must come from the
**pre-change baseline**. This branch contains no operator code changes, so the
CI build output is the baseline and step 0 can be completed straight from CI:
when a golden is missing, test 01 prints the full actual my.cnf between the
`-----BEGIN ACTUAL MY.CNF-----` markers in the log — capture it from the CI
log and commit it:

```bash
# Or generate locally (after creating the same cluster as
# tests/_shared/cluster.yaml in a baseline environment):
kubectl get cm e2e-mysql -n <ns> -o jsonpath='{.data.my\.cnf}' \
  > test/e2e-chainsaw/golden/my.cnf-5.7.44.cnf   # same for 8.0.37
```

Only after the goldens are committed does the golden step of test 01 pass.
**The 8.4.9 golden is generated and reviewed when the 8.4 work lands.**

## Notes

- **Test SQL uses the `sys_operator` account** (password fetched fresh from
  the `e2e-mysql-operated` secret), not root; test data goes into the
  `sys_operator` schema (the account has no INSERT/DELETE on `*.*` but ALL on
  its own schema). Reason: on machines with slow disks the `mysql:5.7` image
  can fail its first-boot init — the entrypoint's health-check client cannot
  reach the actually-ready temporary server (CI run 29088678923: 30 straight
  failures -> `Unable to start server` -> the restarted container skips init),
  so the root password is never set (while the empty password happens to
  work). This is a known issue class of upstream docker-library/mysql; 5.7.44
  is the last 5.7 image (EOL 2023-10) and will never be fixed. The newer 8.0+
  entrypoint does not reproduce it. `sys_operator` is recreated on every start
  via init_file and is unaffected. Production nodes with decent disks are
  usually fine, but spot-checking root usability on existing 5.7.44 instances
  is worthwhile.

- The CR template has the same shape as what mcamel deploys: explicit
  `spec.mysqlVersion` (full version) + `spec.image` (mcamel production uses
  the `library/mysql:{Version}` community image; CI pulls docker.io directly).
- Tests use emptyDir storage (we test operator behavior, not persistence);
  mcamel production uses PVCs.
- The JMESPath expressions in asserts (filtering conditions by type) may need
  tweaking on the first real run; trust chainsaw's actual error output.
- Test 02 force-kills the master with `--force --grace-period=0`: skipping the
  preStop `graceful-master-takeover-auto` hook is what exercises orchestrator's
  DeadMaster crash-recovery path; graceful deletion takes the planned-takeover
  code path instead, which is a different thing.

## Known coverage gaps (vs the deprecated ginkgo suite)

This suite currently covers only the five scenarios of the 06 design doc.
Things the old `test/e2e/` (ginkgo) suite covered that are not covered here:

- backup/restore (MysqlBackup CR, init from backup);
- scaling (replicas 2→3→1) and invariants such as PVC retention after
  scale-in;
- removal of unhealthy nodes from service endpoints (traffic-only scenarios
  that do not trigger failover);
- `spec.readOnly: true` read-only clusters;
- PVC-backed persistence paths (this suite is all emptyDir);
- GTID / read_only variable-level assertions (the old suite asserted
  @@gtid_mode etc. directly).

When adding scenarios, prefer reusing the step template and scripts under
`tests/_shared/`.
