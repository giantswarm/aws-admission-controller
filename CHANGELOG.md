# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added validating webhook to validate `replicas` in the `G8sControlPlane` CR.
- aws-admission-controller metrics
- Validation for control-plane label
- Validation for machine-deployment label

### Changed

- Update k8s.io dependencies to 0.18.9

## [2.0.1] - 2020-09-02

### Changed

- Rename admission-controller to aws-admission-controller

## [2.0.0] - 2020-08-27

### Changed

- Update Kubernetes dependencies to v1.18

## [1.6.0] - 2020-08-21

### Added

- Add NetworkPolicy and security context matching Pod Security Policy.

## [1.5.2] - 2020-08-18

### Fixed

- Fixed label selector for PodDisruptionBudget.

## [1.5.1] - 2020-08-18

### Added

- Change the replicas to 3 and add a PodDisruptionBudget.

## [1.5.0] - 2020-08-14

### Added
- Added mutating webhook to default `availabilityZones` and `instanceType` in the `AWSControlPlane` CR.
- Added mutating webhook to default `replicas` and `infrastructureRef` in the `G8sControlPlane` CR.
- Added unit tests for `AWSControlPlane` and `G8sControlPlane` admitters

## [1.4.0] - 2020-08-10

- If a request has the dry-run flag, update of AWSControlPlane will not be triggered.

### Removed

- Removed Azure-related endpoints.

## [1.3.0] - 2020-07-23

### Changed

- When parsing the release version during Azure upgrades, we are now more tolerant when parsing the versions string so it works as well with leading `v` versions, like `v1.2.3`.

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

[Unreleased]: https://github.com/giantswarm/aws-admission-controller/compare/v2.0.1...HEAD
[2.0.1]: https://github.com/giantswarm/aws-admission-controller/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.6.0...v2.0.0
[1.6.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.5.2...v1.6.0
[1.5.2]: https://github.com/giantswarm/aws-admission-controller/compare/v1.5.1...v1.5.2
[1.5.1]: https://github.com/giantswarm/aws-admission-controller/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/giantswarm/aws-admission-controller/compare/v1.0.0...v0.0.1
[0.0.1]: https://github.com/giantswarm/aws-admission-controller/releases/tag/v0.0.1
