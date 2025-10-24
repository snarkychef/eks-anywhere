# E2E Test Automation Prompt Plan for New Kubernetes Versions - vSphere Provider

This document provides a structured approach for AI coding agents (like Cline) to automate the creation of e2e tests when EKS Anywhere adds support for a new Kubernetes version for the vSphere provider.

## Overview

When adding support for a new Kubernetes version (e.g., 1.34), e2e tests need to be created for the vSphere provider following the established test coverage strategy. Due to the large number of vSphere tests, a reduction strategy has been implemented to keep only essential tests while maintaining adequate coverage.

## Prerequisites

- Target Kubernetes version (e.g., `1.34`)
- Previous Kubernetes version (e.g., `1.33`)
- Oldest supported Kubernetes version (e.g., `1.28`)
- Target provider: `vsphere`

## Version Variables

Throughout this document, the following version variables are used:
- `{NEW_VERSION}`: The new Kubernetes version being added (e.g., 134 for K8s 1.34)
- `{PREV_VERSION}`: The previous Kubernetes version (e.g., 133 for K8s 1.33)
- `{OLDEST_VERSION}`: The oldest supported Kubernetes version (e.g., 128 for K8s 1.28)

**Important**: When using this prompt plan, replace all variable references with actual version numbers for your specific use case.

## Test Coverage Strategy

vSphere tests follow a **tiered coverage approach** to balance comprehensive testing with resource constraints:

### Tier 1: Oldest & Newest Versions Only
These tests are **ONLY** maintained for the **oldest supported version** (currently 1.28) and the **newest version**. When adding K8s 1.34:
- **Replace the newest version** ({PREV_VERSION} → {NEW_VERSION}) 
- **Keep the oldest version** ({OLDEST_VERSION}) unchanged

**Test Categories in Tier 1:**
- **Autoimport**: Only {OLDEST_VERSION} and {NEW_VERSION}
- **Labels Upgrade Flow**: Only {OLDEST_VERSION} and {NEW_VERSION} (both Ubuntu and Bottlerocket)
- **Taints Upgrade Flow**: Only {OLDEST_VERSION} and {NEW_VERSION} (both Ubuntu and Bottlerocket)  
- **Multicluster Workload Cluster**: Only {OLDEST_VERSION} and {NEW_VERSION}
- **Bottlerocket Kubernetes Settings**: Only {OLDEST_VERSION} and {NEW_VERSION}
- **Stacked Etcd Ubuntu**: Only {OLDEST_VERSION} and {NEW_VERSION}
- **Clone Mode** (Full/Linked Clone): Only {OLDEST_VERSION} and {NEW_VERSION} (both Ubuntu and Bottlerocket)
- **NTP Tests**: Only {OLDEST_VERSION} and {NEW_VERSION} (both Ubuntu and Bottlerocket)
- **Etcd Encryption**: Only {OLDEST_VERSION} and {NEW_VERSION} (both Ubuntu and Bottlerocket)
- **Etcd Scale Up/Down**: Only {OLDEST_VERSION} and {NEW_VERSION} (both Ubuntu and Bottlerocket)
- **Kubelet Configuration**: Only 129 and {NEW_VERSION} (both Ubuntu and Bottlerocket) - Note: Uses 129, not {OLDEST_VERSION}
- **In-Place Upgrade Tests**: Heavily reduced - only specific combinations
- **Workload Cluster Taints Flow**: Only {OLDEST_VERSION}

### Tier 2: Full Version Coverage
These tests are maintained for **ALL supported Kubernetes versions**. When adding K8s 1.34:
- **Add new version** ({NEW_VERSION}) alongside existing versions ({OLDEST_VERSION} through {PREV_VERSION})

**Test Categories in Tier 2:**
- **Simple Flow Tests**: All Ubuntu variants (2004, 2204, 2404), Bottlerocket, RedHat9
- **ThreeReplicasFiveWorkers**: All OS variants
- **DifferentNamespace**: All OS variants  
- **Curated Packages**: All package types (base, Emissary, Harbor, ADOT, Cluster Autoscaler, Prometheus)
- **Curated Packages with Proxy**: All versions
- **Workload Cluster Curated Packages**: All package types
- **Flux Tests** (GitHub Flux, Git Flux): All versions
- **OIDC Tests**: All versions
- **Proxy Configuration**: All versions (both Ubuntu and Bottlerocket)
- **Registry Mirror Tests**: All variants (InsecureSkipVerify, AndCert, Authenticated, OciNamespaces)
- **Authenticated Registry Mirror with Curated Packages**: All versions
- **Basic Upgrade Tests**: Most version-to-version upgrades
- **Ubuntu OS Upgrades** (2004→2204, 2004→2404): All versions
- **Stacked Etcd Upgrades**: All versions (Ubuntu, Bottlerocket, RedHat9)
- **Multiple Fields Upgrade**: All versions
- **Node Scaling Upgrades** (CP and Worker): All versions
- **Upgrade from Latest Minor Release**: All versions
- **Airgapped Tests** (Registry Mirror, Proxy): All versions
- **AWS IAM Auth**: Currently {OLDEST_VERSION} through {PREV_VERSION}, continue full coverage pattern

### Tier 3: Latest Version Only  
Some tests are **ONLY** run for the **newest supported version**. When adding K8s 1.34:
- **Replace existing newest** ({PREV_VERSION} → {NEW_VERSION})

**Test Categories in Tier 3:**
- **API Server Extra Args**: Only {NEW_VERSION}
- **Cilium Policy Enforcement Mode**: Only {OLDEST_VERSION} (oldest - special case)
- **Validate Domain Four Levels**: Only {NEW_VERSION}
- **Upgrade Management Components**: Only {OLDEST_VERSION} (oldest)
- **Download Artifacts**: Only {OLDEST_VERSION} (oldest)

## Detailed Task Decomposition Strategy

Use Cline's `new_task` tool to break down the work into these granular subtasks:

### Task 1A: Quick Test Build Configuration
**Scope**: Update quick test buildspec with new template environment variables
**File**: `cmd/integration_test/build/buildspecs/quick-test-eks-a-cli.yml`
**Estimated Lines**: ~10-15 additions

**Specific Actions**:
1. Add template environment variables for new Kubernetes version
2. Follow pattern: `T_VSPHERE_TEMPLATE_{OS}_{VERSION}`

**Template Variables to Add for K8s 1.34**:
```yaml
T_VSPHERE_TEMPLATE_UBUNTU_1_34: "/SDDC-Datacenter/vm/Templates/ubuntu-kube-v1-34"
T_VSPHERE_TEMPLATE_UBUNTU_2204_1_34: "/SDDC-Datacenter/vm/Templates/ubuntu-2204-kube-v1-34"
T_VSPHERE_TEMPLATE_UBUNTU_2404_1_34: "/SDDC-Datacenter/vm/Templates/ubuntu-2404-kube-v1-34"
T_VSPHERE_TEMPLATE_BR_1_34: "/SDDC-Datacenter/vm/Templates/bottlerocket-kube-v1-34"
T_VSPHERE_TEMPLATE_REDHAT_9_1_34: "/SDDC-Datacenter/vm/Templates/redhat-9-kube-v1-34"
```

**Note**: Only RedHat 9 templates are supported for Kubernetes 1.32+. RedHat 8 templates are not included.

### Task 1B: vSphere-Specific Build Configuration
**Scope**: Update vSphere-specific buildspec with new template environment variables
**File**: `cmd/integration_test/build/buildspecs/vsphere-test-eks-a-cli.yml`
**Estimated Lines**: ~10-15 additions

**Specific Actions**:
1. Add new version template variables following same pattern as Task 1A

### Task 2: Quick Tests Configuration Update
**Scope**: Update quick test patterns for new version upgrades
**File**: `test/e2e/QUICK_TESTS.yaml`
**Estimated Lines**: ~10-15 modifications

**Specific Test Patterns to Update** ({PREV_VERSION} → {NEW_VERSION}):
```yaml
- ^TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}RedHatUpgrade$
- TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}StackedEtcdRedHatUpgrade
- ^TestVSphereKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}Upgrade$
- TestVSphereKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}StackedEtcdUpgrade
- TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2204Upgrade
- TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2204StackedEtcdUpgrade
- TestVSphereKubernetes{NEW_VERSION}Ubuntu2004To2204Upgrade
- TestVSphereKubernetes{PREV_VERSION}BottlerocketTo{NEW_VERSION}Upgrade
- TestVSphereKubernetes{PREV_VERSION}BottlerocketTo{NEW_VERSION}StackedEtcdUpgrade
```

## Test Function Implementation

### Tier 1 Tests: Replace Newest Version ({PREV_VERSION} → {NEW_VERSION})

For these tests, **ONLY update/replace the newest version**. Do not add intermediate versions.

#### Task 3A: API Server Extra Args (Tier 3: Latest Only)
**Replace {PREV_VERSION} with {NEW_VERSION}**:
```go
// Replace this:
func TestVSphereKubernetes{PREV_VERSION}BottlerocketAPIServerExtraArgsSimpleFlow(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}BottlerocketAPIServerExtraArgsUpgradeFlow(t *testing.T)

// With:
func TestVSphereKubernetes{NEW_VERSION}BottlerocketAPIServerExtraArgsSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketAPIServerExtraArgsUpgradeFlow(t *testing.T)
```

#### Task 3B: Autoimport (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
// Replace this:
func TestVSphereKubernetes{PREV_VERSION}BottlerocketAutoimport(t *testing.T)

// With:
func TestVSphereKubernetes{NEW_VERSION}BottlerocketAutoimport(t *testing.T)

// Keep unchanged:
func TestVSphereKubernetes{OLDEST_VERSION}BottlerocketAutoimport(t *testing.T)
```

#### Task 3C: Labels Upgrade Flow (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
// Replace these:
func TestVSphereKubernetes{PREV_VERSION}UbuntuLabelsUpgradeFlow(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}BottlerocketLabelsUpgradeFlow(t *testing.T)

// With:
func TestVSphereKubernetes{NEW_VERSION}UbuntuLabelsUpgradeFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketLabelsUpgradeFlow(t *testing.T)

// Keep unchanged:
func TestVSphereKubernetes{OLDEST_VERSION}UbuntuLabelsUpgradeFlow(t *testing.T)
func TestVSphereKubernetes{OLDEST_VERSION}BottlerocketLabelsUpgradeFlow(t *testing.T)
```

#### Task 3D: Taints Upgrade Flow (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
// Replace these:
func TestVSphereKubernetes{PREV_VERSION}UbuntuTaintsUpgradeFlow(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}BottlerocketTaintsUpgradeFlow(t *testing.T)

// With:
func TestVSphereKubernetes{NEW_VERSION}UbuntuTaintsUpgradeFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketTaintsUpgradeFlow(t *testing.T)

// Keep unchanged:
func TestVSphereKubernetes{OLDEST_VERSION}BottlerocketTaintsUpgradeFlow(t *testing.T)
```

#### Task 3E: Multicluster & Special Tests (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
// Multicluster
func TestVSphereKubernetes{NEW_VERSION}MulticlusterWorkloadCluster(t *testing.T) // Replace {PREV_VERSION}

// Bottlerocket Settings (Tier 1)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketWithBottlerocketKubernetesSettings(t *testing.T) // Replace {PREV_VERSION}

// Stacked Etcd (Tier 1)
func TestVSphereKubernetes{NEW_VERSION}StackedEtcdUbuntu(t *testing.T) // Replace {PREV_VERSION}

// Keep {OLDEST_VERSION} versions unchanged
```

#### Task 3F: Clone Mode Tests (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
// Ubuntu Clone Mode
func TestVSphereKubernetes{NEW_VERSION}FullClone(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}LinkedClone(t *testing.T) // Replace {PREV_VERSION}

// Bottlerocket Clone Mode
func TestVSphereKubernetes{NEW_VERSION}BottlerocketFullClone(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}BottlerocketLinkedClone(t *testing.T) // Replace {PREV_VERSION}

// Keep {OLDEST_VERSION} versions unchanged
```

#### Task 3G: NTP Tests (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
func TestVSphereKubernetes{NEW_VERSION}BottleRocketWithNTP(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}UbuntuWithNTP(t *testing.T) // Replace {PREV_VERSION}

// Keep {OLDEST_VERSION} versions unchanged
```

#### Task 3H: Kubelet Configuration (Tier 1: Special Pattern)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep 129** (note: uses 129, not {OLDEST_VERSION}):
```go
func TestVSphereKubernetes{NEW_VERSION}UbuntuKubeletConfiguration(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}BottlerocketKubeletConfiguration(t *testing.T) // Replace {PREV_VERSION}

// Keep unchanged (special case - uses 129, not {OLDEST_VERSION}):
func TestVSphereKubernetes129UbuntuKubeletConfiguration(t *testing.T)
func TestVSphereKubernetes129BottlerocketKubeletConfiguration(t *testing.T)
```

#### Task 3I: Etcd Encryption (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
func TestVSphereKubernetesUbuntu{NEW_VERSION}EtcdEncryption(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetesBottlerocket{NEW_VERSION}EtcdEncryption(t *testing.T) // Replace {PREV_VERSION}

// Keep {OLDEST_VERSION} versions unchanged
```

#### Task 3J: Etcd Scaling Tests (Tier 1: Oldest & Newest)
**Replace {PREV_VERSION} with {NEW_VERSION}, keep {OLDEST_VERSION}**:
```go
// Bottlerocket Etcd Scale
func TestVSphereKubernetes{NEW_VERSION}BottlerocketEtcdScaleUp(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}BottlerocketEtcdScaleDown(t *testing.T) // Replace {PREV_VERSION}

// Ubuntu Etcd Scale  
func TestVSphereKubernetes{NEW_VERSION}UbuntuEtcdScaleUp(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}UbuntuEtcdScaleDown(t *testing.T) // Replace {PREV_VERSION}

// Etcd Scale with Upgrade
func TestVSphereKubernetes{PREV_VERSION}to{NEW_VERSION}UbuntuEtcdScaleUp(t *testing.T) // Replace previous transition
func TestVSphereKubernetes{PREV_VERSION}to{NEW_VERSION}UbuntuEtcdScaleDown(t *testing.T) // Replace previous transition

// Keep {OLDEST_VERSION} versions unchanged
```

#### Task 3K: In-Place Upgrade Tests (Tier 1: Selective)
**Replace {PREV_VERSION} with {NEW_VERSION} for specific tests only**:
```go
// Replace these:
func TestVSphereKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}InPlaceUpgrade_1CP_1Worker(t *testing.T) // Replace previous transition
func TestVSphereKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}InPlaceUpgradeWorkerOnly(t *testing.T) // Replace previous transition
func TestVSphereKubernetes{NEW_VERSION}UbuntuInPlaceCPScaleUp1To3(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}UbuntuInPlaceCPScaleDown3To1(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}UbuntuInPlaceWorkerScaleUp1To2(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereKubernetes{NEW_VERSION}UbuntuInPlaceWorkerScaleDown2To1(t *testing.T) // Replace {PREV_VERSION}
func TestVSphereInPlaceUpgradeMulticlusterWorkloadClusterK8sUpgrade{PREV_VERSION}To{NEW_VERSION}(t *testing.T) // Replace previous transition

// Keep these ({OLDEST_VERSION} tests):
func TestVSphereKubernetes{OLDEST_VERSION}UbuntuTo129InPlaceUpgradeCPOnly(t *testing.T)
func TestVSphereKubernetes{OLDEST_VERSION}UbuntuTo129InPlaceUpgrade_3CP_3Worker(t *testing.T) // If it exists
func TestVSphereKubernetes{OLDEST_VERSION}UbuntuTo{NEW_VERSION}InPlaceUpgrade(t *testing.T) // Update target to {NEW_VERSION}
func TestVSphereInPlaceUpgradeMulticlusterWorkloadClusterK8sUpgrade{OLDEST_VERSION}To129(t *testing.T)
```

#### Task 3L: Special Version Tests (Tier 3)
**Replace with new version**:
```go
// Management components - only oldest version
func TestVSphereKubernetes{OLDEST_VERSION}UpgradeManagementComponents(t *testing.T) // Keep unchanged

// Download artifacts - only oldest version  
func TestVSphereDownloadArtifacts(t *testing.T) // Keep unchanged

// Domain validation - only newest version
func TestVSphereKubernetes{NEW_VERSION}ValidateDomainFourLevelsSimpleFlow(t *testing.T) // Replace {PREV_VERSION}

// Management cluster side effects - only newest version
func TestVSphereKubernetes{NEW_VERSION}WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T) // Replace {PREV_VERSION}

// Cilium policy - only oldest version
func TestVSphereKubernetes{OLDEST_VERSION}CiliumAlwaysPolicyEnforcementModeSimpleFlow(t *testing.T) // Keep unchanged
```

### Tier 2 Tests: Add New Version (Full Coverage)

For these tests, **ADD the new version** alongside existing versions (do not remove old versions).

#### Task 4A: Simple Flow Tests - All Ubuntu Variants
**Add {NEW_VERSION} alongside {OLDEST_VERSION} through {PREV_VERSION}**:
```go
func TestVSphereKubernetes{NEW_VERSION}Ubuntu2004SimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}Ubuntu2204SimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}Ubuntu2404SimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}ThreeReplicasFiveWorkersSimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}DifferentNamespaceSimpleFlow(t *testing.T) // Add new
```

#### Task 4B: Simple Flow Tests - Bottlerocket Variants
**Add {NEW_VERSION} alongside {OLDEST_VERSION} through {PREV_VERSION}**:
```go
func TestVSphereKubernetes{NEW_VERSION}BottleRocketSimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}BottleRocketDifferentNamespaceSimpleFlow(t *testing.T) // Add new
```

#### Task 4C: Simple Flow Tests - RedHat9 Variant
**Add {NEW_VERSION} alongside {OLDEST_VERSION} through {PREV_VERSION}**:
```go
func TestVSphereKubernetes{NEW_VERSION}RedHat9SimpleFlow(t *testing.T) // Add new
```

#### Task 4D: AWS IAM Auth Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
// Ubuntu AWS IAM Auth - add if following full coverage pattern
func TestVSphereKubernetes{NEW_VERSION}AWSIamAuth(t *testing.T) // Add new (check if pattern continues)

// Bottlerocket AWS IAM Auth - add new
func TestVSphereKubernetes{NEW_VERSION}BottleRocketAWSIamAuth(t *testing.T) // Add new

// Upgrade
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}AWSIamAuthUpgrade(t *testing.T) // Add new
```

#### Task 4E: Curated Packages - Core Tests
**Add {NEW_VERSION} alongside {OLDEST_VERSION} through {PREV_VERSION}**:
```go
func TestVSphereKubernetes{NEW_VERSION}CuratedPackagesSimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}BottleRocketCuratedPackagesSimpleFlow(t *testing.T) // Add new
func TestVSphereKubernetes{NEW_VERSION}CuratedPackagesWithProxyConfigFlow(t *testing.T) // Add new
```

#### Task 4F: Curated Packages - All Package Types
**Add {NEW_VERSION} for each package type**:
```go
// Emissary
func TestVSphereKubernetes{NEW_VERSION}CuratedPackagesEmissarySimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T)

// Harbor
func TestVSphereKubernetes{NEW_VERSION}CuratedPackagesHarborSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T)

// ADOT
func TestVSphereKubernetes{NEW_VERSION}CuratedPackagesAdotUpdateFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T)

// Cluster Autoscaler
func TestVSphereKubernetes{NEW_VERSION}UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketWorkloadClusterCuratedPackagesClusterAutoscalerUpgradeFlow(t *testing.T)

// Prometheus
func TestVSphereKubernetes{NEW_VERSION}UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T)
```

#### Task 4G: Workload Cluster Curated Packages
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{NEW_VERSION}UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T)
```

#### Task 4H: Flux Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
// GitHub Flux
func TestVSphereKubernetes{NEW_VERSION}GithubFlux(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketGithubFlux(t *testing.T)

// Git Flux
func TestVSphereKubernetes{NEW_VERSION}GitFlux(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottleRocketGitFlux(t *testing.T)

// Git Flux Upgrade
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}GitFluxUpgrade(t *testing.T)
```

#### Task 4I: OIDC Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{NEW_VERSION}OIDC(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}OIDCUpgrade(t *testing.T)
```

#### Task 4J: Proxy Configuration Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{NEW_VERSION}UbuntuProxyConfigFlow(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketProxyConfigFlow(t *testing.T)
```

#### Task 4K: Registry Mirror Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{NEW_VERSION}UbuntuRegistryMirrorInsecureSkipVerify(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuRegistryMirrorAndCert(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketRegistryMirrorAndCert(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuAuthenticatedRegistryMirror(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketAuthenticatedRegistryMirror(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketRegistryMirrorOciNamespaces(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T)
```

#### Task 4L: Upgrade Tests - Version-to-Version
**Add new upgrade paths**:
```go
// Ubuntu Upgrades
func TestVSphereKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}Upgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2204Upgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2204StackedEtcdUpgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2404Upgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2404StackedEtcdUpgrade(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}Ubuntu2004To2204Upgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}StackedEtcdUpgrade(t *testing.T) // If it exists

// Bottlerocket Upgrades
func TestVSphereKubernetes{PREV_VERSION}BottlerocketTo{NEW_VERSION}Upgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}BottlerocketTo{NEW_VERSION}StackedEtcdUpgrade(t *testing.T)

// RedHat9 Upgrades
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}RedHat9Upgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}StackedEtcdRedHat9Upgrade(t *testing.T)
```

#### Task 4M: Multiple Fields Upgrade Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}MultipleFieldsUpgrade(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}BottlerocketTo{NEW_VERSION}MultipleFieldsUpgrade(t *testing.T)
```

#### Task 4N: Node Scaling Upgrade Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{NEW_VERSION}UbuntuControlPlaneNodeUpgrade(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuWorkerNodeUpgrade(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketControlPlaneNodeUpgrade(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}BottlerocketWorkerNodeUpgrade(t *testing.T)
```

#### Task 4O: Upgrade from Latest Minor Release Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}UbuntuUpgradeFromLatestMinorRelease(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}UbuntuInPlaceUpgradeFromLatestMinorRelease(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat9UpgradeFromLatestMinorRelease(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuUpgradeAndRemoveWorkerNodeGroupsAPI(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}to{NEW_VERSION}UpgradeFromLatestMinorReleaseBottleRocketAPI(t *testing.T)
func TestVSphereKubernetes{PREV_VERSION}Redhat9UpgradeFromLatestMinorRelease(t *testing.T) // Keep if exists
```

#### Task 4P: Airgapped Tests
**Add {NEW_VERSION} alongside existing versions**:
```go
func TestVSphereKubernetes{NEW_VERSION}UbuntuAirgappedRegistryMirror(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UbuntuAirgappedProxy(t *testing.T)
```

#### Task 4Q: Workload API Tests
**Replace {PREV_VERSION} with {NEW_VERSION}**:
```go
func TestVSphereKubernetes{NEW_VERSION}MulticlusterWorkloadClusterAPI(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UpgradeLabelsTaintsUbuntuAPI(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UpgradeWorkerNodeGroupsUbuntuAPI(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}MulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}CiliumUbuntuAPI(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UpgradeLabelsTaintsBottleRocketGitHubFluxAPI(t *testing.T)
func TestVSphereKubernetes{NEW_VERSION}UpgradeWorkerNodeGroupsUbuntuGitHubFluxAPI(t *testing.T)
func TestVSphereUpgradeKubernetes{NEW_VERSION}CiliumUbuntuGitHubFluxAPI(t *testing.T)
```

### Task 5: Framework Helper Functions
**File**: `test/framework/vsphere.go`
**Estimated Lines**: ~60-80 additions

**Add these helper functions for K8s {NEW_VERSION}**:
```go
// OS Variant Helpers
func WithUbuntu{NEW_VERSION}() VSphereOpt
func WithBottleRocket{NEW_VERSION}() VSphereOpt
func WithRedHat9{NEW_VERSION}VSphere() VSphereOpt
func (v *VSphere) WithUbuntu{NEW_VERSION}() api.ClusterConfigFiller
func (v *VSphere) WithBottleRocket{NEW_VERSION}() api.ClusterConfigFiller

// Template Functions
func (v *VSphere) Ubuntu{NEW_VERSION}Template() api.VSphereFiller
func (v *VSphere) Ubuntu2204Kubernetes{NEW_VERSION}Template() api.VSphereFiller
func (v *VSphere) Ubuntu2404Kubernetes{NEW_VERSION}Template() api.VSphereFiller
func (v *VSphere) Bottlerocket{NEW_VERSION}Template() api.VSphereFiller
func (v *VSphere) Redhat9{NEW_VERSION}Template() api.VSphereFiller

// Template Functions for Machine Config
func (v *VSphere) Ubuntu{NEW_VERSION}TemplateForMachineConfig(name string) api.VSphereFiller
```

**Note**: RedHat 8 functions (e.g., `WithRedHat{NEW_VERSION}VSphere()`, `Redhat{NEW_VERSION}Template()`) are not needed as RedHat 8 is not supported for Kubernetes 1.32+.

### Task 6: Provider-Specific Helper Functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~80-100 additions/modifications

**Replace {PREV_VERSION} with {NEW_VERSION}** (Tier 1 pattern):
```go
// Labels helpers
func ubuntu{NEW_VERSION}ProviderWithLabels(t *testing.T) *framework.Vsphere // Replace {PREV_VERSION}
func bottlerocket{NEW_VERSION}ProviderWithLabels(t *testing.T) *framework.Vsphere // Replace {PREV_VERSION}

// Taints helpers
func ubuntu{NEW_VERSION}ProviderWithTaints(t *testing.T) *framework.Vsphere // Replace {PREV_VERSION}
func bottlerocket{NEW_VERSION}ProviderWithTaints(t *testing.T) *framework.Vsphere // Replace {PREV_VERSION}

// Keep {OLDEST_VERSION} versions unchanged
```

**⚠️ CRITICAL: Clean Up Orphaned Helper Functions**

When replacing {PREV_VERSION} helpers with {NEW_VERSION} helpers, **you must also remove orphaned intermediate version helpers** (versions between {OLDEST_VERSION} and {PREV_VERSION}) that are no longer called by any tests.

For example, when adding K8s 1.34 (where OLDEST=128, PREV=133, NEW=134):
- **Keep**: ubuntu128ProviderWithLabels, ubuntu134ProviderWithLabels (oldest + newest)
- **Remove**: ubuntu129ProviderWithLabels, ubuntu130ProviderWithLabels, ubuntu131ProviderWithLabels, ubuntu132ProviderWithLabels, ubuntu133ProviderWithLabels
- **Keep**: bottlerocket128ProviderWithLabels, bottlerocket134ProviderWithLabels (oldest + newest)
- **Remove**: bottlerocket129-133ProviderWithLabels (5 functions)
- **Same pattern for Taints helpers**: Keep 128 & 134, remove 129-133

**Total orphaned functions to remove**: 20
- 5 ubuntu ProviderWithLabels (129-133)
- 5 bottlerocket ProviderWithLabels (129-133)
- 5 ubuntu ProviderWithTaints (129-133)
- 5 bottlerocket ProviderWithTaints (129-133)

**Verification Steps**:
1. Search for unused helper functions: `(ubuntu|bottlerocket)(129|130|131|132|133)ProviderWith(Labels|Taints)`
2. Verify they are defined but never called in test functions
3. Remove all orphaned intermediate version helpers
4. Ensure only {OLDEST_VERSION} and {NEW_VERSION} helper functions remain

**Why This Happens**: vSphere uses Tier 1 coverage for Labels/Taints tests, keeping only oldest+newest. When test reductions occurred, the intermediate helper functions were not cleaned up, leaving orphaned code.

## Implementation Guidelines

### Naming Conventions
- Test functions: `TestVSphereKubernetes{Version}{OS}{Feature}{TestType}` (Note: VSphere with capital S)
- Helper functions: `With{OS}{Version}()`, `{OS}{Version}Template()`
- Provider helper functions: `{os}{version}ProviderWithLabels()`, `{os}{version}ProviderWithTaints()`
- Constants: Follow existing patterns in the codebase

### Code Patterns

#### Tier 1 Test Pattern (Replace Newest):
```go
// Before (for K8s {PREV_VERSION}):
func TestVSphereKubernetes{PREV_VERSION}BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube{PREV_VERSION})),
	)
	runAutoImportFlow(test, provider)
}

// After (for K8s {NEW_VERSION}):
func TestVSphereKubernetes{NEW_VERSION}BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube{NEW_VERSION})),
	)
	runAutoImportFlow(test, provider)
}
```

#### Tier 2 Test Pattern (Add New Version):
```go
// Add alongside existing {OLDEST_VERSION} through {PREV_VERSION}:
func TestVSphereKubernetes{NEW_VERSION}Ubuntu2004SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu{NEW_VERSION}()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube{NEW_VERSION})),
	)
	runSimpleFlow(test)
}
```

### Critical Rules

1. **Tier 1 Tests**: Replace newest version ({PREV_VERSION} → {NEW_VERSION}), keep oldest ({OLDEST_VERSION})
2. **Tier 2 Tests**: Add new version ({NEW_VERSION}) alongside all existing versions
3. **Tier 3 Tests**: Replace newest with new version or keep oldest only
4. **Version References**: Always update `api.WithKubernetesVersion(v1alpha1.Kube{VERSION})`
5. **Template References**: Always update template function calls (e.g., `provider.Ubuntu{NEW_VERSION}Template()`)
6. **Provider Functions**: Update framework helper functions with new version
7. **RedHat 8**: Never add RedHat 8 tests for Kubernetes 1.32+
8. **Ubuntu Variants**: Support all Ubuntu variants (2004, 2204, 2404) for full coverage tests

### Quick Reference: Test Coverage Tiers

| Test Category | Coverage Pattern | K8s 1.34 Action |
|--------------|------------------|-----------------|
| API Server Extra Args | Latest only | Replace {PREV_VERSION} → {NEW_VERSION} |
| Autoimport | Oldest + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep {OLDEST_VERSION} |
| Labels/Taints Upgrade | Oldest + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep {OLDEST_VERSION} |
| Multicluster | Oldest + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep {OLDEST_VERSION} |
| Clone Mode | Oldest + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep {OLDEST_VERSION} |
| NTP | Oldest + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep {OLDEST_VERSION} |
| Etcd Encryption | Oldest + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep {OLDEST_VERSION} |
| Etcd Scaling | Oldest + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep {OLDEST_VERSION} |
| Kubelet Config | 129 + Newest | Replace {PREV_VERSION} → {NEW_VERSION}, keep 129 |
| In-Place Upgrade | Selective | Replace specific {PREV_VERSION} → {NEW_VERSION} |
| Simple Flow | All versions | Add {NEW_VERSION} alongside {OLDEST_VERSION}-{PREV_VERSION} |
| Curated Packages | All versions | Add {NEW_VERSION} alongside {OLDEST_VERSION}-{PREV_VERSION} |
| Flux | All versions | Add {NEW_VERSION} alongside {OLDEST_VERSION}-{PREV_VERSION} |
| OIDC | All versions | Add {NEW_VERSION} alongside {OLDEST_VERSION}-{PREV_VERSION} |
| Proxy Config | All versions | Add {NEW_VERSION} alongside {OLDEST_VERSION}-{PREV_VERSION} |
| Registry Mirror | All versions | Add {NEW_VERSION} alongside {OLDEST_VERSION}-{PREV_VERSION} |
| Upgrade Tests | All versions | Add {NEW_VERSION} upgrade paths |
| Airgapped | All versions | Add {NEW_VERSION} alongside {OLDEST_VERSION}-{PREV_VERSION} |

### Quality Checks
1. Ensure all test functions compile
2. Verify naming consistency across all functions
3. Check that coverage tier is correctly applied for each test category
4. Validate that upgrade paths are correctly defined
5. Ensure framework helper functions are properly integrated
6. Verify only {OLDEST_VERSION} and {NEW_VERSION} exist for Tier 1 tests (no intermediate versions)
7. Verify all versions {OLDEST_VERSION}-{NEW_VERSION} exist for Tier 2 tests

## Execution Strategy

### Phase 1: Infrastructure Setup
Execute Tasks 1-2 to set up the basic infrastructure and configuration

### Phase 2: Tier 1 Tests (Replace Pattern)
Execute Task 3 subtasks to replace newest version tests with K8s {NEW_VERSION}

### Phase 3: Tier 2 Tests (Add Pattern)  
Execute Task 4 subtasks to add K8s {NEW_VERSION} alongside existing versions

### Phase 4: Framework Functions
Execute Task 5 to add framework helper functions

### Phase 5: Provider Helpers
Execute Task 6 to update provider-specific helper functions

### Phase 6: Validation
Run compilation checks and validate coverage tiers are correctly applied

## Context Management

To avoid context overflow:
1. Use `new_task` tool between major task categories
2. Focus on one coverage tier at a time
3. Group related test categories together
4. Preserve context about patterns and conventions between tasks

## Recent Test Additions (K8s 1.33+)

### Ubuntu 24.04 Support (Added October 2025, commit 4ca40fbd6)
Ubuntu 24.04 tests were added with **full version coverage** (Tier 2) for K8s {OLDEST_VERSION} through {PREV_VERSION}:

**Simple Flow Tests** (All versions {OLDEST_VERSION} through {PREV_VERSION}):
- `TestVSphereKubernetes{VERSION}Ubuntu2404SimpleFlow`

**Upgrade Tests** (All version transitions):
- `TestVSphereKubernetes{VERSION}To{VERSION+1}Ubuntu2404Upgrade`
- `TestVSphereKubernetes{VERSION}To{VERSION+1}Ubuntu2404StackedEtcdUpgrade`

**When adding new Kubernetes version**: 
- Add `TestVSphereKubernetes{NEW_VERSION}Ubuntu2404SimpleFlow` (Tier 2 - add pattern)
- Add `TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2404Upgrade` (Tier 2 - add pattern)
- Add `TestVSphereKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2404StackedEtcdUpgrade` (Tier 2 - add pattern)

**Framework Updates Needed**:
- Ensure `Ubuntu2404Kubernetes{NEW_VERSION}Template()` is added
- Template environment variable: `T_VSPHERE_TEMPLATE_UBUNTU_2404_1_{NEW_VERSION}`

## Important Notes

- **Test Reduction**: vSphere implemented significant test reduction (removed ~2000 lines) to manage resource constraints
- **Two Reduction Commits**: Reductions happened in July 2025 (ff258c67f) and September 2025 (b311feb95)
- **Coverage Tiers**: Follow the documented tiers strictly - don't add intermediate versions for Tier 1 tests
- **RedHat 8**: Not supported for Kubernetes 1.32 onwards - only RedHat 9
- **Ubuntu 24.04**: Supported from K8s {OLDEST_VERSION} onwards - include in full coverage tests (Tier 2)
- **Version Progression**: When oldest version is eventually dropped, the next oldest becomes the new baseline for Tier 1 tests

## Example Usage for Adding K8s 1.34

### Tier 1 Example (Replace):
```
Find: TestVSphereKubernetes{PREV_VERSION}BottlerocketAutoimport
Replace with: TestVSphereKubernetes{NEW_VERSION}BottlerocketAutoimport
Update: v1alpha1.Kube{PREV_VERSION} → v1alpha1.Kube{NEW_VERSION}
Keep: TestVSphereKubernetes{OLDEST_VERSION}BottlerocketAutoimport (unchanged)
```

### Tier 2 Example (Add):
```
Add new: TestVSphereKubernetes{NEW_VERSION}Ubuntu2004SimpleFlow
Keep existing: All {OLDEST_VERSION}-{PREV_VERSION} versions unchanged
Pattern: Copy {PREV_VERSION} function, update version to {NEW_VERSION}
```

## Validation Checklist

After implementing all tasks:
- [ ] All Tier 1 tests have exactly 2 versions ({OLDEST_VERSION} and {NEW_VERSION})
- [ ] All Tier 2 tests have all versions ({OLDEST_VERSION} through {NEW_VERSION})
- [ ] All Tier 3 tests have correct version ({OLDEST_VERSION} or {NEW_VERSION} as specified)
- [ ] No intermediate versions exist for Tier 1 tests
- [ ] Framework helper functions added for version {NEW_VERSION}
- [ ] Provider helper functions updated for version {NEW_VERSION}
- [ ] Quick test configuration updated
- [ ] Build configuration files updated
- [ ] All version references updated (constants, function names, etc.)
- [ ] No RedHat 8 references for K8s 1.32+
- [ ] Ubuntu 2404 variant included where applicable

This systematic approach ensures appropriate test coverage while respecting resource constraints through the tiered testing strategy.
