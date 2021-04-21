# Testing

## End-to-end testing

The test uses KinD with MetalLB, so external traffic can be routed into
the cluster.

### How to use it

To run the tests locally, run the following commands:

1. set up a test env (creates a kind cluster with MetalLB and pre-loads docker images):

       make e2e-test-env

2. run the tests (builds the istio-operator docker image, installs it using helm and starts the tests):

       make e2e-test

Currently, the kind cluster needs to be deleted before the tests can be re-run.
The test itself can be re-run, but the istio-operator will not be updated, so code changes
won't have any effect on the tests. This can (and will) be fixed with some work, but
generally, the tests expect a clean cluster, so it's best to start with a clean one.

The tests expect a clean cluster, so after running the tests, it might be best to delete the
kind cluster. If the tests pass, the cluster will be left in a (relatively) clean state, so
re-running the tests should be ok. On the other hand, a failed test will leave the cluster
as-is to help with investigating the issue, and the leftover resources could interfere with
further test runs.

### Current issues

The MetalLB setup doesn't work on MacOS, yet.

The tests are tailored for running on a KinD cluster with preloaded docker images,
so running the tests on a real cluster will most probably fail with timeouts
because of the additional time it takes to pull the docker images.
