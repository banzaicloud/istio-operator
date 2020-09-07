# Istio upgrade

Starting from Istio 1.7 the recommended way to upgrade an Istio control plane is with the [canary upgrade workflow](canary-upgrade.md), because the data plane can be gradually moved to use the new control plane version and hence this upgrade model is safer than the original in-place upgrade process.

The [in-place upgrade](in-place-upgrade.md) model is still available, it is now only recommended for Istio patch versions when the data plane migration can be done in one step, with low risk for any traffic disruptions.
