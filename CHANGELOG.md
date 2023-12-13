# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Fix the requests duration metrics exposed by the controller, as the duration was always 0. 
### Added
- Add `global.podSecurityStandards.enforced` value for PSS migration.
### Changed

- Configure `gsoci.azurecr.io` as the default container image registry.

## [4.12.0] - 2023-08-23

### Changed

- Don't change pod CIDR during upgrade from v18 to v19 if eni ipam mode is enabled.

## [4.11.0] - 2023-07-14

### Fixed

- Fix deployment manifest indentation

## [4.10.0] - 2023-07-13

### Fixed

- Add required values for pss policies.

## [4.9.1] - 2023-06-28

### Fixed

- Check validity of `cilium.giantswarm.io/ipam-mode` on `Cluster` CR update as well.

## [4.9.0] - 2023-06-28

### Added

- Add validation for `cilium.giantswarm.io/ipam-mode` annotation on `Cluster` CR creation.

## [4.8.1] - 2023-06-07

### Fixed

- When upgrading to cilium, ensure defaulting of needed annotations even when the cilium CIDR annotation is already in place.

### Added

- Add service monitor to be scraped by Prometheus Agent.

## [4.8.0] - 2023-04-18

### Added

- Added the use of the runtime default seccomp profile
- Block upgrade from v18 to v19 if cluster is gitops managed and Kustomization is not suspended.

## [4.7.0] - 2023-01-04

### Added

- Validate if cluster already exists.

## [4.6.0] - 2022-10-10

### Added

- Add annotation to disable kube-proxy feature in cilium during migration from AWS-cni to cilium.

## [4.5.0] - 2022-09-21

### Changed

- Making mutation and validation `dryRun`-aware.

## [4.4.0] - 2022-08-24

#### Changed

- Move cilium to `v19.x.x` AWS release.

## [4.3.0] - 2022-08-09

### Added

- Default `cilium.giantswarm.io/pod-cidr` annotation on Cluster CR when upgrading from v17 to v18.

## [4.2.0] - 2022-08-04

### Added

- Validate `cilium.giantswarm.io/pod-cidr` annotation does not intersect with current CIDR's.
- Validate `cilium.giantswarm.io/pod-cidr` annotation is present and valid while upgrading from v17 to v18.

## [4.1.0] - 2022-07-22

## [4.0.3] - 2022-06-15

### Changed

- Ignore repeated AZs in China even if there are 3 available.

### Added

- Skip validation for `service-priority` label.

## [4.0.2] - 2022-04-04

### Fixed

- Bump go module version to v4.

## [4.0.1] - 2022-04-01

### Fixed

- Update HTTP handler endpoints to `v1beta1`.

## [4.0.0] - 2022-03-30

### Changed

- Update `apiextensions` to `v6.0.0`.
- Update `k8sclient` to `v7.0.1`.
- Update `cluster-api` to `v1.0.4`.
- Use `v1beta1` CAPI CRDs.

## [3.7.0] - 2022-03-21

### Added

- Add VerticalPodAutoscaler CR.
- Validate EBS volume iops and throughput settings.

## [3.6.3] - 2021-12-17

### Fixed

- Handle update patch with empty version string

## [3.6.2] - 2021-12-01

### Fixed

- Patch `infraRef` in case of updating g8scontrolplane.

## [3.6.1] - 2021-11-29

### Fixed

- Allow setting ASG max scaling to 0 if ASG min is also set to 0.

## [3.6.0] - 2021-11-26

### Added

- Validate if AWS CNI prefix can be enabled.

### Changed

- Use `k8smetadata` for annotations.

## [3.5.1] - 2021-11-12

### Fixed

- Fixed the release considered the first CAPI release.

## [3.5.0] - 2021-11-09

### Added

- Add `cluster.x-k8s.io/cluster-name` for all CRs.

## [3.4.0] - 2021-09-28

### Added

- Validate operator version on all CR's.

## [3.3.0] - 2021-09-10

- Added validation for `alpha.giantswarm.io/update-schedule-target-release` annotation.
- Added validation for `alpha.giantswarm.io/update-schedule-target-time` annotation.
- Added validation and tests for cluster CRs to be created in the organization-namespace from version `16.0.0`.
- Added tests for nodepool CRs to be created in the same namespace as the cluster.

## [3.2.1] - 2021-08-12

### Added

- Added rule for users to be allowed to modify/add/remove tags with the `tag.provider.giantswarm.io` format.

## [3.2.0] - 2021-08-10

### Added

- Added validation for the `availabilityZones` selectopm in the `AWSMachineDeployment.

## [3.1.0] - 2021-08-06

### Added

- Add `cluster.x-k8s.io/cluster-name` for `v1alpha3` CAPI CR's.

### Changed

- Default `clusterName` for `MachineDeployments` if empty.

### Fixed

- Fixed `infrastrutureRef` in /spec/template/spec for `MachineDeployments`.

## [3.0.0] - 2021-08-03

### Changed

- Bump CRD API version to `v1alpha3`.

## [2.12.0] - 2021-07-23

### Changed

- Prepare helm values to configuration management.
- Update architect-orb to v4.0.0.

## [2.11.0] - 2021-05-31

### Removed

- AWSMachinedeployment with maximum 0 nodes is no longer rejected, as long as minimum is also 0.

### Added

- Default the control-plane label on awscontrolplane and g8scontrolplane to the control-plane name if they are not set.
- Validate that control-plane label on awscontrolplane and g8scontrolplane are set
- Validate existing organization label when cluster is created.

## [2.10.0] - 2021-04-06

### Added

- Do not validate and mutate for CAPI controller release versions
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

- Handling of creation and updates to [`AWSMachineDeployment`](https://docs.giantswarm.io/ui-api/management-api/crd/awsmachinedeployments.infrastructure.giantswarm.io/) (`awsmachinedeployments.infrastructure.giantswarm.io`) resources, with defaulting of the [`.spec.node_spec.aws.instanceDistribution.onDemandPercentageAboveBaseCapacity`](https://docs.giantswarm.io/ui-api/management-api/crd/awsmachinedeployments.infrastructure.giantswarm.io/#v1alpha2-.spec.provider.instanceDistribution.onDemandPercentageAboveBaseCapacity) attribute.

## [1.0.0] - 2020-06-15

- Several changes

## [0.1.0] - 2020-06-10

- First release.

[Unreleased]: https://github.com/giantswarm/aws-admission-controller/compare/v4.12.0...HEAD
[4.12.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.11.0...v4.12.0
[4.11.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.10.0...v4.11.0
[4.10.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.9.1...v4.10.0
[4.9.1]: https://github.com/giantswarm/aws-admission-controller/compare/v4.9.0...v4.9.1
[4.9.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.8.1...v4.9.0
[4.8.1]: https://github.com/giantswarm/aws-admission-controller/compare/v4.8.0...v4.8.1
[4.8.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.7.0...v4.8.0
[4.7.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.6.0...v4.7.0
[4.6.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.5.0...v4.6.0
[4.5.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.4.0...v4.5.0
[4.4.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.3.0...v4.4.0
[4.3.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.2.0...v4.3.0
[4.2.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.1.0...v4.2.0
[4.1.0]: https://github.com/giantswarm/aws-admission-controller/compare/v4.0.3...v4.1.0
[4.0.3]: https://github.com/giantswarm/aws-admission-controller/compare/v4.0.2...v4.0.3
[4.0.2]: https://github.com/giantswarm/aws-admission-controller/compare/v4.0.1...v4.0.2
[4.0.1]: https://github.com/giantswarm/aws-admission-controller/compare/v4.0.0...v4.0.1
[4.0.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.7.0...v4.0.0
[3.7.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.6.3...v3.7.0
[3.6.3]: https://github.com/giantswarm/aws-admission-controller/compare/v3.6.2...v3.6.3
[3.6.2]: https://github.com/giantswarm/aws-admission-controller/compare/v3.6.1...v3.6.2
[3.6.1]: https://github.com/giantswarm/aws-admission-controller/compare/v3.6.0...v3.6.1
[3.6.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.5.1...v3.6.0
[3.5.1]: https://github.com/giantswarm/aws-admission-controller/compare/v3.5.0...v3.5.1
[3.5.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.4.0...v3.5.0
[3.4.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.3.0...v3.4.0
[3.3.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.2.1...v3.3.0
[3.2.1]: https://github.com/giantswarm/aws-admission-controller/compare/v3.2.0...v3.2.1
[3.2.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.1.0...v3.2.0
[3.1.0]: https://github.com/giantswarm/aws-admission-controller/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.12.0...v3.0.0
[2.12.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.11.0...v2.12.0
[2.11.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.10.0...v2.11.0
[2.10.0]: https://github.com/giantswarm/aws-admission-controller/compare/v2.9.1...v2.10.0
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
