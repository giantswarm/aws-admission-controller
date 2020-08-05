# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Added mutating webhook to default `availabilityZones` and `instanceType` in the `AWSControlPlane` CR.
- Added mutating webhook to default `replicas` and `infrastructureRef` in the `G8sControlPlane` CR.
- Added unit tests for `AWSControlPlane` and `G8sControlPlane` admitters

## [1.2.0] - 2020-07-20

### Added

- Validation Webhooks that check for valid upgrade paths for legacy Azure clusters.
- Added application to Azure app collection.

## [1.1.0] - 2020-07-16

### Added

- Handling of creation and updates to [`AWSMachineDeployment`](https://docs.giantswarm.io/reference/cp-k8s-api/awsmachinedeployments.infrastructure.giantswarm.io) (`awsmachinedeployments.infrastructure.giantswarm.io`) resources, with defaulting of the [`.spec.node_spec.aws.instanceDistribution.onDemandPercentageAboveBaseCapacity`](https://docs.giantswarm.io/reference/cp-k8s-api/awsmachinedeployments.infrastructure.giantswarm.io/#v1alpha2-.spec.provider.instanceDistribution.onDemandPercentageAboveBaseCapacity) attribute.

## [1.0.0] - 2020-06-15

- Several changes

## [0.1.0] - 2020-06-10

- First release.

[Unreleased]: https://github.com/giantswarm/admission-controller/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/giantswarm/admission-controller/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/giantswarm/admission-controller/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/giantswarm/admission-controller/compare/v1.0.0...v0.0.1
[0.0.1]: https://github.com/giantswarm/admission-controller/releases/tag/v0.0.1
