# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Adding validation for worker instance types in `AWSMachinedeployment` CR.
- Prevent max number of nodes in `AWSMachinedeployment` CR to be 0 or smaller than min number of nodes.
- Prevent upgrades (changing release version label on `Cluster` CR) if the Cluster has not transitioned yet.

## [2.9.1] - 2021-02-03

### Added

- Only allow customers to change the major release version in the `Cluster` CR to a version that is greater than the current one,
  but does not skip major release versions. 

## [2.9.0] - 2021-02-01

### Added

- Prevent creation of `AWSMachinedeployment` CR if the related cluster is deleted.
- Add Validator for `Machinedeployment` CRs and prevent their creation if the related cluster is deleted.

## [2.8.0] - 2021-01-21

- Adding validation for AWS CNI annotations for `AWSCluster` CR.

## [2.7.0] - 2021-01-19

### Added

- Adding label value validation for `Cluster` CR for non-version labels.
- Adding label key validation for `Cluster` CR for `giantswarm.io` labels.
- Adding label value validation for `Cluster` CR for version labels.

## [2.6.0] - 2020-12-07

### Added

- Default the Operator Version Label in `Cluster` to match the new release during upgrade.
- Default the Release Version Label in `Cluster` to the newest active production release.

## [2.5.0] - 2020-12-01

### Added

- Default the Availability Zones in `AWSMachinedeployment` based on `AWSControlplane` CR.

## [2.4.1] - 2020-11-24

- Check all patches for a release version

## [2.4.0] - 2020-11-24

### Changed

- Changed defaulting of the Infrastructure reference in the `G8sControlPlane` to not require `AWSControlPlane`to already exist.

## [2.3.3] - 2020-11-19

### Added

- Default the Cluster Operator Version Label in `Cluster` from `Release` CR.
- Default the AWS Operator Version Label in `AWSCluster` from `Release` CR.
- Default the AWS Operator Version Label in `AWSControlplane`, `AWSMachinedeployment` Mutators and add generic label defaulting from AWSCluster CR.
- Default the Cluster Operator Version Label in `G8sControlplane`, `Machinedeployment` Mutators and add generic label defaulting from cluster CR.
- Default the Master attributes in the `AWSCluster` Mutator for pre-HA versions.
- Default the Release Version Label and refactor the `G8sControlplane` and `AWSControlPlane` Mutators.
- Default the Release Version Label in the `AWSCluster`, `MachineDeployment` and `AWSMachineDeployment` CRs based on the `Cluster`CR

## [2.2.3] - 2020-11-18

- Set `400` status code in the validator response if a request is invalid.

## [2.2.2] - 2020-11-10

### Added

- Added defaulting for the Cluster credential secret in the `AWSCluster` CR

### Fixed

- Fix validation of `alpha.aws.giantswarm.io/update-pause-time` to allow maximum value of 1 hour.

## [2.2.1] - 2020-11-05

### Fixed

- Auto refresh certificate when renewed

## [2.2.0] - 2020-11-04

### Changed

- Use cert-manager v1 API

### Added

- Added defaulting for the Cluster region in the `AWSCluster` CR
- Added defaulting for the Cluster description in the `AWSCluster` CR
- Added defaulting for the Cluster DNS domain in the `AWSCluster` CR
- Added validation for `alpha.aws.giantswarm.io/update-max-batch-size` annotation on the `AWSCluster` CR.
- Added validation for `alpha.aws.giantswarm.io/update-pause-time` annotation on the `AWSCluster` CR.
- Added validation for `alpha.aws.giantswarm.io/update-max-batch-size` annotation on the `AWSMachineDeployment` CR.
- Added validation for `alpha.aws.giantswarm.io/update-pause-time` annotation on the `AWSMachineDeployment` CR.

## [2.1.0] - 2020-10-29

### Added

- Added defaulting for the Pod CIDR in the `AWSCluster` CR
- Added rules to validate `instanceType` in the `AWSControlPlane` CR.
- Added rules to validate `availabilityZones` in the `AWSControlPlane` CR.
- Added validating webhook to validate `replicas` in the `G8sControlPlane` CR.
- aws-admission-controller metrics
- Validation for control-plane label
- Validation for machine-deployment label
- Validation for NetworkPools

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

[Unreleased]: https://github.com/giantswarm/aws-admission-controller/compare/v2.9.1...HEAD
[2.9.1]: https://github.com/giantswarm/aws-admission-controller/compare/v2.9.0...v2.9.1
[2.9.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.8.0...v2.9.0
[2.8.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.7.0...v2.8.0
[2.7.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.6.0...v2.7.0
[2.6.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.5.0...v2.6.0
[2.5.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.4.1...v2.5.0
[2.4.1]: https://github.com/giantswarm/aws-admission-controller/compare/v2.4.0...v2.4.1
[2.4.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.3.3...v2.4.0
[2.3.3]: https://github.com/giantswarm/aws-admission-controller/compare/v2.2.3...v2.3.3
[2.2.3]: https://github.com/giantswarm/aws-admission-controller/compare/v2.2.2...v2.2.3
[2.2.2]: https://github.com/giantswarm/aws-admission-controller/compare/v2.2.1...v2.2.2
[2.2.1]: https://github.com/giantswarm/aws-admission-controller/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.0.1...v2.1.0
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
