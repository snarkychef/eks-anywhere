# E2E Test Automation Prompt Plan for New Kubernetes Versions

This document provides a structured prompt plan for automating the creation of e2e tests when adding support for a new Kubernetes version to a provider in EKS Anywhere.

## Overview

When EKS Anywhere adds support for a new Kubernetes version, comprehensive e2e tests need to be created following established patterns. This automation plan breaks down the task into manageable subtasks to avoid context overflow and ensure systematic implementation.

## Task Decomposition Strategy

Use Cline's `new_task` tool to create separate tasks for each major component to manage context effectively:

### Task 1: Update Test Configuration Files
**Scope**: Update configuration files that control test execution, build system, and documentation
**Files**: 
- `test/e2e/QUICK_TESTS.yaml`
- `test/e2e/constants.go`
- `Makefile`
- `scripts/e2e_test_docker.sh`
- `test/e2e/README.md`

### Task 2: Update Core Test Functions (Part 1)
**Scope**: Update existing test functions to use the new Kubernetes version
**Files**: 
- `test/e2e/docker_test.go` (first 500 lines)

### Task 3: Update Core Test Functions (Part 2) 
**Scope**: Continue updating existing test functions
**Files**: 
- `test/e2e/docker_test.go` (lines 500-1000)

### Task 4: Add New Test Functions (Part 1)
**Scope**: Add new test functions for the new Kubernetes version
**Files**: 
- `test/e2e/docker_test.go` (curated packages tests)

### Task 5: Add New Test Functions (Part 2)
**Scope**: Add remaining new test functions
**Files**: 
- `test/e2e/docker_test.go` (auth, registry, simple flow tests)

### Task 6: Update Upgrade Test Functions
**Scope**: Update upgrade test functions to use new version transitions
**Files**: 
- `test/e2e/docker_test.go` (upgrade test functions)

## Detailed Implementation Prompts

### Task 1 Prompt: Update Test Configuration Files

```
I need to update the e2e test configuration files to support Kubernetes version {NEW_VERSION} for the Docker provider.

Based on the reference pattern, please:

1. Update `test/e2e/QUICK_TESTS.yaml`:
   - Change the test pattern from `TestDocker.*{PREVIOUS_VERSION}` to `TestDocker.*{NEW_VERSION}`

2. Update `test/e2e/constants.go`:
   - Add `v1alpha1.Kube{NEW_VERSION}` to the `KubeVersions` slice
   - Maintain the existing order (ascending version numbers)

3. Update `Makefile`:
   - Change the `DOCKER_E2E_TEST` variable from `TestDockerKubernetes{PREVIOUS_VERSION}SimpleFlow` to `TestDockerKubernetes{NEW_VERSION}SimpleFlow`

4. Update `scripts/e2e_test_docker.sh`:
   - Change the `TEST_REGEX` default value from `TestDockerKubernetes{PREVIOUS_VERSION}SimpleFlow` to `TestDockerKubernetes{NEW_VERSION}SimpleFlow`

5. Update `test/e2e/README.md`:
   - Update all occurrences of `TestDockerKubernetes{PREVIOUS_VERSION}SimpleFlow` to `TestDockerKubernetes{NEW_VERSION}SimpleFlow`
   - This includes comments and example commands

Variables to replace:
- {NEW_VERSION}: The new Kubernetes version number (e.g., 134)
- {PREVIOUS_VERSION}: The previous Kubernetes version number (e.g., 133)
```

### Task 2 Prompt: Update Core Test Functions (Part 1)

```
I need to update existing e2e test functions in the first part of `test/e2e/docker_test.go` to use Kubernetes version {NEW_VERSION}.

Please update the following types of functions (approximately first 500 lines):

1. **Label Tests**: Update functions like `TestDockerKubernetesLabels` to use `v1alpha1.Kube{NEW_VERSION}`

2. **Flux Tests**: Update functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}GithubFlux` → `TestDockerKubernetes{NEW_VERSION}GithubFlux`
   - `TestDockerKubernetes{PREVIOUS_VERSION}GitFlux` → `TestDockerKubernetes{NEW_VERSION}GitFlux`
   - Update the `api.WithKubernetesVersion()` calls within these functions

3. **Flux Upgrade Tests**: Update functions like:
   - `TestDockerInstallGitFluxDuringUpgrade`
   - `TestDockerInstallGithubFluxDuringUpgrade`
   - Update both function names and internal version references

Pattern to follow:
- Function names: Change version numbers in function names
- Internal calls: Update `api.WithKubernetesVersion(v1alpha1.Kube{OLD})` to `api.WithKubernetesVersion(v1alpha1.Kube{NEW})`
- Upgrade flows: Update version parameters in upgrade function calls

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 3 Prompt: Update Core Test Functions (Part 2)

```
Continue updating existing e2e test functions in `test/e2e/docker_test.go` (lines 500-1000) to use Kubernetes version {NEW_VERSION}.

Focus on updating:

1. **Workload Cluster Tests**: Update functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}UpgradeWorkloadClusterWithGithubFlux`
   - Update both the function name and internal version references
   - Update upgrade target versions appropriately

2. **Taints Tests**: Update functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}Taints` → `TestDockerKubernetes{NEW_VERSION}Taints`
   - `TestDockerKubernetes{PREVIOUS_VERSION}WorkloadClusterTaints` → `TestDockerKubernetes{NEW_VERSION}WorkloadClusterTaints`

3. **Simple Flow Tests**: Update functions like:
   - Update any remaining simple flow test functions to use the new version

Follow the same pattern as Task 2:
- Update function names with version numbers
- Update `api.WithKubernetesVersion()` calls
- Update upgrade flow version parameters

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 4 Prompt: Add New Test Functions (Part 1)

```
I need to add new e2e test functions for Kubernetes version {NEW_VERSION} in `test/e2e/docker_test.go`.

Please add the following new test functions by copying and modifying existing patterns:

1. **Curated Packages Tests**: Add new functions for:
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesEmissarySimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesHarborSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesAdotSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesPrometheusSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesDisabled`

2. **MetalLB Test**: Add:
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesMetalLB`

Pattern to follow:
- Copy the corresponding {PREVIOUS_VERSION} function
- Update function name to use {NEW_VERSION}
- Update `api.WithKubernetesVersion(v1alpha1.Kube{PREVIOUS_VERSION})` to `api.WithKubernetesVersion(v1alpha1.Kube{NEW_VERSION})`
- Update `packageBundleURI(v1alpha1.Kube{PREVIOUS_VERSION})` to `packageBundleURI(v1alpha1.Kube{NEW_VERSION})`
- Keep all other parameters and function calls identical

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 5 Prompt: Add New Test Functions (Part 2)

```
Continue adding new e2e test functions for Kubernetes version {NEW_VERSION} in `test/e2e/docker_test.go`.

Please add the following new test functions:

1. **Authentication Tests**: Add:
   - `TestDockerKubernetes{NEW_VERSION}AWSIamAuth`
   - `TestDockerKubernetes{NEW_VERSION}OIDC`

2. **Registry Mirror Tests**: Add:
   - `TestDockerKubernetes{NEW_VERSION}RegistryMirrorInsecureSkipVerify`

3. **Simple Flow Test**: Add:
   - `TestDockerKubernetes{NEW_VERSION}SimpleFlow`

4. **Kubelet Configuration Test**: Add:
   - `TestDockerKubernetes{NEW_VERSION}KubeletConfigurationSimpleFlow`

5. **Etcd Scale Tests**: Add:
   - `TestDockerKubernetes{NEW_VERSION}EtcdScaleUp`
   - `TestDockerKubernetes{NEW_VERSION}EtcdScaleDown`

Follow the same pattern as Task 4:
- Copy the corresponding {PREVIOUS_VERSION} function
- Update function name to use {NEW_VERSION}
- Update all `api.WithKubernetesVersion()` calls
- Keep all other parameters identical

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 6 Prompt: Update Upgrade Test Functions

```
I need to update upgrade test functions in `test/e2e/docker_test.go` to support upgrading to Kubernetes version {NEW_VERSION}.

Please update the following types of upgrade tests:

1. **Version-to-Version Upgrade Tests**: Add functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}To{NEW_VERSION}StackedEtcdUpgrade`
   - `TestDockerKubernetes{PREVIOUS_VERSION}To{NEW_VERSION}ExternalEtcdUpgrade`

2. **CLI-based Upgrade from Latest Release Tests**: Add functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}UpgradeFromLatestMinorRelease`
   - `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}GithubFluxEnabledUpgradeFromLatestMinorRelease`

3. **Workload Cluster GitOps Upgrade Tests** (only if {NEW_VERSION} >= 133):
   - Add: `TestDockerKubernetes{NEW_VERSION}UpgradeWorkloadClusterWithGithubFlux`
   - Pattern: Upgrades workload cluster from {PREVIOUS_VERSION}→{NEW_VERSION} using GitOps
   - This test pattern started at K8s 132

4. **API-based Workload Cluster Upgrade Tests** (only if {PREVIOUS_VERSION} >= 131):
   - Add: `TestDockerUpgradeKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}WorkloadClusterScaleupGitHubFluxAPI`
   - Pattern: API-based upgrade with GitHub Flux
   - This test pattern started at K8s 131→132
   - Note: Do NOT add `UpgradeFromLatestMinorReleaseAPI` variant - only the CLI-based version exists

5. **Management Cluster Tests**: Add:
   - `TestDockerKubernetes{NEW_VERSION}WithOIDCManagementClusterUpgradeFromLatestSideEffects`

6. **Etcd Scale with Upgrade Tests**: Add:
   - `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}EtcdScaleUp`
   - `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}EtcdScaleDown`

Pattern for upgrade tests:
- Update function names to reflect new version transitions
- Update initial cluster version: `api.WithKubernetesVersion(v1alpha1.Kube{PREVIOUS_VERSION})`
- Update target upgrade version: `api.WithKubernetesVersion(v1alpha1.Kube{NEW_VERSION})`
- Update upgrade flow parameters: `v1alpha1.Kube{NEW_VERSION}`

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

## Test Pattern Version Requirements

Some test patterns were introduced in later Kubernetes versions. Verify version requirements before adding:

### Patterns Starting from K8s 1.31→1.32:
- `TestDockerKubernetes{PREV}to{NEW}UpgradeFromLatestMinorReleaseAPI` (only API variant, not CLI)
- `TestDockerUpgradeKubernetes{PREV}to{NEW}WorkloadClusterScaleupGitHubFluxAPI`

**For K8s 1.34**: Add `TestDockerUpgradeKubernetes133to134WorkloadClusterScaleupGitHubFluxAPI`

### Patterns Starting from K8s 1.32:
- `TestDockerKubernetes{NEW}UpgradeWorkloadClusterWithGithubFlux` (upgrades from {PREV}→{NEW})

**For K8s 1.34**: Add `TestDockerKubernetes134UpgradeWorkloadClusterWithGithubFlux` (upgrades 133→134)

### Patterns Starting from K8s 1.29:
- `TestDockerKubernetes{NEW}KubeletConfigurationSimpleFlow`

**For K8s 1.34**: Add `TestDockerKubernetes134KubeletConfigurationSimpleFlow`

### Patterns Starting from K8s 1.33:
- `TestDockerKubernetes{NEW}SkipAdmissionForSystemResources`

**For K8s 1.34**: Add `TestDockerKubernetes134SkipAdmissionForSystemResources`

### Important Notes:
- **Do NOT add** `TestDockerKubernetes{PREV}to{NEW}UpgradeFromLatestMinorReleaseAPI` - this API variant doesn't exist. Only the CLI-based variant exists (without "API" suffix).
- Always check existing test patterns before blindly adding new tests
- When in doubt, search for similar patterns in earlier versions to confirm

## Usage Instructions

1. **Preparation**: 
   - Identify the new Kubernetes version number (e.g., 134)
   - Identify the previous version number (e.g., 133)
   - Identify the provider name (Docker)
   - **Check version requirements** above to see which patterns apply

2. **Variable Substitution**:
   Replace the following variables in all prompts:
   - `{NEW_VERSION}`: New Kubernetes version (e.g., 134)
   - `{PREVIOUS_VERSION}`: Previous Kubernetes version (e.g., 133)

3. **Execution**:
   - Use Cline's `new_task` tool to create separate tasks for each section
   - Execute tasks in order (1-6)
   - Verify each task completion before proceeding to the next
   - **Verify version requirements** before adding tests from Task 6

4. **Validation**:
   - After all tasks are complete, run the tests to ensure they compile and execute correctly
   - Check that all version references are updated consistently
   - Verify that upgrade paths are logical (previous → new version)
   - Verify no incorrect test patterns were added

## Recent Test Additions (K8s 1.33+)

### Skip Admission Plugins Test (Added October 2025, commit eb6c48a12)
A new test for skipping admission plugins during critical control plane upgrades:
- `TestDockerKubernetes133SkipAdmissionForSystemResources`

**When adding K8s 1.34**: Add `TestDockerKubernetes134SkipAdmissionForSystemResources` alongside existing 133 test (Tier 2 - full coverage pattern).

## Test Coverage Notes

### Version-Specific Coverage Patterns

Docker maintains **full version coverage** for most test categories, similar to CloudStack and Nutanix. However, note these specific patterns:

1. **Kubelet Configuration Tests**:
   - Start from K8s **1.29 onwards** (no 1.28 test)
   - When adding K8s 1.34: **Add 134 alongside existing 129-133**
   - Pattern: Full coverage from 129+

2. **Skip Admission Plugins Test**:
   - Added in K8s 1.33
   - When adding K8s 1.34: **Add 134 alongside 133**
   - Pattern: Full coverage from 133+

3. **All Other Test Categories**:
   - Maintain full version coverage for all supported versions
   - Follow the standard "add new version" pattern

### Test Coverage Strategy

Docker uses a **full version coverage** approach for nearly all test categories. When adding K8s 1.34:
- **Add new version** (134) alongside existing versions (128-133)
- **Exception**: For KubeletConfiguration, verify pattern starts from 129

Unlike vSphere (oldest/newest only) or Tinkerbell (replacement), Docker maintains comprehensive historical test coverage similar to CloudStack and Nutanix.

## Notes

- This plan is specifically for the Docker provider
- Docker maintains full version coverage similar to CloudStack and Nutanix
- Kubernetes 1.33 support has already been added to the Docker provider (as of the referenced patch)
- Always verify that the new Kubernetes version constant exists in the codebase before starting
- Consider running a subset of tests first to validate the changes before full test suite execution
- When adding support for Kubernetes 1.34, use 1.33 as the {PREVIOUS_VERSION} and 1.32 as the {PREV_PREV_VERSION}
- **Important**: Docker tests are generally additive - old version tests are preserved for backward compatibility
