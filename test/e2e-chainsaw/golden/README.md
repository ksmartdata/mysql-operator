# my.cnf golden files

This directory holds the expected content of each version's my.cnf ConfigMap
(compared byte for byte). It is the mechanical enforcement of the "zero config
diff for existing versions" rule — if any operator change unintentionally
alters the my.cnf generated for 5.7/8.0 clusters, the golden step of test 01
goes red.

See "Step 0" in the parent README for how to generate them. **Goldens must be
taken from the pre-change baseline** — do not regenerate them on a feature
branch to "make the test green"; that bakes the regression into the expected
values. The only legitimate reason to update a golden: an intentional config
change, explained and reviewed in the MR.

**Goldens are coupled to the CR resource spec**: the operator derives
`innodb-buffer-pool-size` / `innodb-log-file-size` from the pod memory and
writes them into my.cnf (appended at the end of the file). Whenever the
resources in `tests/_shared/cluster.yaml` change, the goldens must be updated
in the same change (observed in CI: 512Mi→768Mi added a buffer-pool line).

- my.cnf-5.7.44.cnf — committed (baseline run 29083492185 + 768Mi resource fix)
- my.cnf-8.0.37.cnf — committed (same as above)
- my.cnf-8.4.9.cnf  — to be generated and reviewed when the 8.4 work lands
