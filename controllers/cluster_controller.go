package controllers

import (
	"context"
	"fmt"
	"time"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/handlers"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	defaultRequeueTime = time.Minute
	// ClusterFinalizerName is the finalizer added to clusters to handle deletion.
	ClusterFinalizerName = "clusters.anywhere.eks.amazonaws.com/finalizer"
	releaseV022          = "v0.22.0"
)

// ClusterReconciler reconciles a Cluster object.
type ClusterReconciler struct {
	client                     client.Client
	providerReconcilerRegistry ProviderClusterReconcilerRegistry
	awsIamAuth                 AWSIamConfigReconciler
	clusterValidator           ClusterValidator
	packagesClient             PackagesClient
	machineHealthCheck         MachineHealthCheckReconciler
	vSpherefailureDomainMover  FailureDomainApplier
}

// PackagesClient handles curated packages operations from within the cluster
// controller.
type PackagesClient interface {
	EnableFullLifecycle(ctx context.Context, log logr.Logger, clusterName, kubeConfig string, chart *v1alpha1.Image, registry *registrymirror.RegistryMirror, options ...curatedpackages.PackageControllerClientOpt) error
	ReconcileDelete(context.Context, logr.Logger, curatedpackages.KubeDeleter, *anywherev1.Cluster) error
	Reconcile(context.Context, logr.Logger, client.Client, *anywherev1.Cluster) error
}

type ProviderClusterReconcilerRegistry interface {
	Get(datacenterKind string) clusters.ProviderClusterReconciler
}

// AWSIamConfigReconciler manages aws-iam-authenticator installation and configuration for an eks-a cluster.
type AWSIamConfigReconciler interface {
	EnsureCASecret(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
	Reconcile(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
	ReconcileDelete(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) error
}

// MachineHealthCheckReconciler manages machine health checks for an eks-a cluster.
type MachineHealthCheckReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) error
}

// ClusterValidator runs cluster level preflight validations before it goes to provider reconciler.
type ClusterValidator interface {
	ValidateManagementClusterName(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error
}

// ClusterReconcilerOption allows to configure the ClusterReconciler.
type ClusterReconcilerOption func(*ClusterReconciler)

// SpecBuilder builds a cluster specification from an EKS Anywhere Cluster object.
type SpecBuilder interface {
	BuildSpec(ctx context.Context, eksaCluster *anywherev1.Cluster) (*c.Spec, error)
}

// FailureDomainSpecBuilder transforms a cluster specification into VSphere-specific failure domains.
type FailureDomainSpecBuilder interface {
	BuildFailureDomainSpec(log logr.Logger, clusterSpec *c.Spec) (*vsphere.FailureDomains, error)
}

// ObjectReconciler applies failure domain objects to a Kubernetes cluster.
type ObjectReconciler interface {
	ReconcileObjects(ctx context.Context, fd *vsphere.FailureDomains) error
}

// FailureDomainApplier orchestrates the end-to-end process of applying failure domains to a cluster.
type FailureDomainApplier interface {
	ApplyFailureDomains(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error
}

// DefaultSpecBuilder is the standard implementation of SpecBuilder that uses a Kubernetes client.
type DefaultSpecBuilder struct {
	client client.Client
}

// BuildSpec is a wrapper method for building and obtaining all neccessary objects from a cluster.
func (b *DefaultSpecBuilder) BuildSpec(ctx context.Context, cluster *anywherev1.Cluster) (*c.Spec, error) {
	return c.BuildSpec(ctx, clientutil.NewKubeClient(b.client), cluster)
}

// DefaultFailureDomainSpecBuilder is the standard implementation of FailureDomainSpecBuilder.
type DefaultFailureDomainSpecBuilder struct{}

// BuildFailureDomainSpec wrapper to the vsphere package's FailureDomainsSpec function.
func (b *DefaultFailureDomainSpecBuilder) BuildFailureDomainSpec(log logr.Logger, clusterSpec *c.Spec) (*vsphere.FailureDomains, error) {
	return vsphere.FailureDomainsSpec(log, clusterSpec)
}

// DefaultObjectReconciler is the standard implementation of ObjectReconciler that applies objects using a client.
type DefaultObjectReconciler struct {
	client client.Client
}

// ReconcileObjects applies failure domain objects to the cluster using serverside reconciliation.
func (r *DefaultObjectReconciler) ReconcileObjects(ctx context.Context, fd *vsphere.FailureDomains) error {
	return serverside.ReconcileObjects(ctx, r.client, fd.Objects())
}

// FailureDomainMover defines config for applying failure domain objects.
type FailureDomainMover struct {
	specBuilder      SpecBuilder
	fdSpecBuilder    FailureDomainSpecBuilder
	objectReconciler ObjectReconciler
}

// NewFailureDomainMover builds FailureDomainMover with default dependencies.
func NewFailureDomainMover(client client.Client) *FailureDomainMover {
	return &FailureDomainMover{
		specBuilder:      &DefaultSpecBuilder{client: client},
		fdSpecBuilder:    &DefaultFailureDomainSpecBuilder{},
		objectReconciler: &DefaultObjectReconciler{client: client},
	}
}

// NewFailureDomainMoverWithDependencies builds FailureDomainMover with specified dependencies.
func NewFailureDomainMoverWithDependencies(
	specBuilder SpecBuilder,
	fdSpecBuilder FailureDomainSpecBuilder,
	objectReconciler ObjectReconciler,
) *FailureDomainMover {
	return &FailureDomainMover{
		specBuilder:      specBuilder,
		fdSpecBuilder:    fdSpecBuilder,
		objectReconciler: objectReconciler,
	}
}

// NewClusterReconciler constructs a new ClusterReconciler.
func NewClusterReconciler(client client.Client, registry ProviderClusterReconcilerRegistry, awsIamAuth AWSIamConfigReconciler, clusterValidator ClusterValidator, pkgs PackagesClient, machineHealthCheck MachineHealthCheckReconciler, failuredomainmover FailureDomainApplier, opts ...ClusterReconcilerOption) *ClusterReconciler {
	c := &ClusterReconciler{
		client:                     client,
		providerReconcilerRegistry: registry,
		awsIamAuth:                 awsIamAuth,
		clusterValidator:           clusterValidator,
		packagesClient:             pkgs,
		machineHealthCheck:         machineHealthCheck,
		vSpherefailureDomainMover:  failuredomainmover,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager, log logr.Logger) error {
	childObjectHandler := handlers.ChildObjectToClusters(log)

	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.Cluster{}).
		Watches(
			&anywherev1.OIDCConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.AWSIamConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.GitOpsConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.FluxConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.VSphereDatacenterConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.VSphereMachineConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.SnowDatacenterConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.SnowMachineConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.TinkerbellDatacenterConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.TinkerbellMachineConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.DockerDatacenterConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.CloudStackDatacenterConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.CloudStackMachineConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.NutanixDatacenterConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&anywherev1.NutanixMachineConfig{},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Complete(r)
}

// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch;update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;delete;update;patch
// +kubebuilder:rbac:groups="",namespace=eksa-system,resources=secrets,verbs=patch;update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;create;delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=list
// +kubebuilder:rbac:groups=addons.cluster.x-k8s.io,resources=clusterresourcesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;gitopsconfigs;snowmachineconfigs;snowdatacenterconfigs;snowippools;vspheredatacenterconfigs;vspheremachineconfigs;dockerdatacenterconfigs;tinkerbellmachineconfigs;tinkerbelltemplateconfigs;tinkerbelldatacenterconfigs;cloudstackdatacenterconfigs;cloudstackmachineconfigs;nutanixdatacenterconfigs;nutanixmachineconfigs;awsiamconfigs;oidcconfigs;awsiamconfigs;fluxconfigs,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;snowmachineconfigs/status;snowippools/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;dockerdatacenterconfigs/status;tinkerbelldatacenterconfigs/status;tinkerbellmachineconfigs/status;tinkerbelltemplateconfigs/status;cloudstackdatacenterconfigs/status;cloudstackmachineconfigs/status;awsiamconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=bundles,verbs=get;list;watch
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;snowmachineconfigs/finalizers;snowippools/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;cloudstackdatacenterconfigs/finalizers;cloudstackmachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers;tinkerbelldatacenterconfigs/finalizers;tinkerbellmachineconfigs/finalizers;tinkerbelltemplateconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=kubeadmconfigtemplates,verbs=create;get;list;patch;update;watch
// +kubebuilder:rbac:groups="cluster.x-k8s.io",resources=machinedeployments,verbs=list;watch;get;patch;update;create;delete
// +kubebuilder:rbac:groups="cluster.x-k8s.io",resources=clusters,verbs=list;watch;get;patch;update;create;delete
// +kubebuilder:rbac:groups="cluster.x-k8s.io",resources=machinehealthchecks,verbs=list;watch;get;patch;create
// +kubebuilder:rbac:groups=clusterctl.cluster.x-k8s.io,resources=providers,verbs=get;list;watch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplanes,verbs=list;get;watch;patch;update;create;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create;get;list;update;watch;delete
// +kubebuilder:rbac:groups=distro.eks.amazonaws.com,resources=releases,verbs=get;list;watch
// +kubebuilder:rbac:groups=etcdcluster.cluster.x-k8s.io,resources=*,verbs=create;get;list;patch;update;watch
// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware,verbs=list;watch
// +kubebuilder:rbac:groups=bmc.tinkerbell.org,resources=machines,verbs=list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awssnowclusters;awssnowmachinetemplates;awssnowippools;vsphereclusters;vspheremachinetemplates;dockerclusters;dockermachinetemplates;tinkerbellclusters;tinkerbellmachinetemplates;cloudstackclusters;cloudstackmachinetemplates;nutanixclusters;nutanixmachinetemplates;vspherefailuredomains;vspheredeploymentzones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=packages.eks.amazonaws.com,resources=packages,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=packages.eks.amazonaws.com,namespace=eksa-system,resources=packagebundlecontrollers,verbs=delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=eksareleases,verbs=get;list;watch
// The eksareleases permissions are being moved to the ClusterRole due to client trying to list this resource from cache.
// When trying to list resources not already in cache, it starts an informer for that type using the scope of the cache.
// So if the manager is cluster-scoped, the new informers created by the cache will be cluster-scoped

// Reconcile reconciles a cluster object.
// nolint:gocyclo
// TODO: Reduce high cycomatic complexity. https://github.com/aws/eks-anywhere-internal/issues/1449
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	// Fetch the Cluster objects
	cluster := &anywherev1.Cluster{}
	log.Info("Reconciling cluster")
	if err := r.client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(cluster, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		err := r.updateStatus(ctx, log, cluster)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Always attempt to patch the object and status after each reconciliation.
		patchOpts := []patch.Option{}

		// We want the observedGeneration to indicate, that the status shown is up-to-date given the desired spec of the same generation.
		// However, if there is an error while updating the status, we may get a partial status update, In this case,
		// a partially updated status is not considered up to date, so we should not update the observedGeneration

		// Patch ObservedGeneration only if the reconciliation completed without error
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchCluster(ctx, patchHelper, cluster, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Only requeue if we are not already re-queueing and the Cluster ready condition is false.
		// We do this to be able to update the status continuously until the cluster becomes ready,
		// since there might be changes in state of the world that don't trigger reconciliation requests

		if reterr == nil && !result.Requeue && result.RequeueAfter <= 0 && conditions.IsFalse(cluster, anywherev1.ReadyCondition) {
			result = ctrl.Result{RequeueAfter: 10 * time.Second}
		}
	}()

	if !cluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, cluster)
	}

	// If the cluster is paused, return without any further processing.
	if cluster.IsReconcilePaused() {
		log.Info("Cluster reconciliation is paused")
		return ctrl.Result{}, nil
	}

	// AddFinalizer	is idempotent
	controllerutil.AddFinalizer(cluster, ClusterFinalizerName)

	if !cluster.IsSelfManaged() && cluster.Spec.BundlesRef == nil && cluster.Spec.EksaVersion == nil {
		if err = r.setDefaultBundlesRefOrEksaVersion(ctx, cluster); err != nil {
			return ctrl.Result{}, nil
		}
	}

	config, err := r.buildClusterConfig(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err = r.ensureClusterOwnerReferences(ctx, cluster, config); err != nil {
		return ctrl.Result{}, err
	}

	aggregatedGeneration := aggregatedGeneration(config)

	// If there is no difference between the aggregated generation and childrenReconciledGeneration,
	// and there is no difference in the reconciled generation and .metadata.generation of the cluster,
	// then return without any further processing.
	if aggregatedGeneration == cluster.Status.ChildrenReconciledGeneration && cluster.Status.ReconciledGeneration == cluster.Generation {
		log.Info("Generation and aggregated generation match reconciled generations for cluster and child objects, skipping reconciliation.")

		// Failure messages are cleared in the reconciler loop after running validations. But sometimes,
		// it seems that Cluster failure messages on the status are is not cleared for some reason
		//  after successfully passing the validation. The theory is that if the inital patch operation
		// is not successful, and the reconciliation is skipped going forward, it may never be cleared.
		//
		// When the controller reaches here, it denotes a completed reconcile. So, we can safely
		// clear any failure messages or reasons that may be left over as there is no further processing
		// for the controller to do.
		if cluster.HasFailure() {
			cluster.ClearFailure()
		}

		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, cluster, aggregatedGeneration)
}

func (r *ClusterReconciler) reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster, aggregatedGeneration int64) (ctrl.Result, error) {
	clusterProviderReconciler := r.providerReconcilerRegistry.Get(cluster.Spec.DatacenterRef.Kind)

	var reconcileResult controller.Result
	var err error

	reconcileResult, err = r.preClusterProviderReconcile(ctx, log, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileResult.Return() {
		return reconcileResult.ToCtrlResult(), nil
	}

	reconcileResult, err = clusterProviderReconciler.Reconcile(ctx, log, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileResult.Return() {
		return reconcileResult.ToCtrlResult(), nil
	}

	reconcileResult, err = r.postClusterProviderReconcile(ctx, log, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileResult.Return() {
		return reconcileResult.ToCtrlResult(), nil
	}

	// At the end of the reconciliation, if there have been no requeues or errors, we update the cluster's status.
	// NOTE: This update must be the last step in the reconciliation process to denote the complete reconciliation.
	// No other mutating changes or reconciliations must happen in this loop after this step, so all such changes must
	// be placed above this line.
	cluster.Status.ReconciledGeneration = cluster.Generation
	cluster.Status.ChildrenReconciledGeneration = aggregatedGeneration

	// TODO(eksa-controller-SME): properly handle packages reconcile error and not triggering machine upgrade when
	// packages reconcile is still in progress.
	// Moving the packages reconcile after the above two generation fields are set, so that packages reconcile error
	// does not cause side effect of rolling out of workload cluster machines during management cluster upgrade.
	reconcileResult, err = r.packagesReconcile(ctx, log, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileResult.Return() {
		return reconcileResult.ToCtrlResult(), nil
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) preClusterProviderReconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	// Run some preflight validations that can't be checked in webhook
	if cluster.HasAWSIamConfig() {
		if result, err := r.awsIamAuth.EnsureCASecret(ctx, log, cluster); err != nil {
			return controller.Result{}, err
		} else if result.Return() {
			return result, nil
		}
	}
	if cluster.IsManaged() {
		if err := r.clusterValidator.ValidateManagementClusterName(ctx, log, cluster); err != nil {
			log.Error(err, "Invalid cluster configuration")
			cluster.SetFailure(anywherev1.ManagementClusterRefInvalidReason, err.Error())
			return controller.Result{}, err
		}

		mgmt, err := getManagementCluster(ctx, cluster, r.client)
		if err != nil {
			return controller.Result{}, err
		}

		if err := validations.ValidateManagementEksaVersion(mgmt, cluster); err != nil {
			return controller.Result{}, err
		}
	}

	if err := validateEksaRelease(ctx, r.client, cluster); err != nil {
		return controller.Result{}, err
	}

	if err := validateExtendedKubernetesVersionSupport(ctx, r.client, cluster); err != nil {
		return controller.Result{}, err
	}

	if cluster.RegistryAuth() {
		rUsername, rPassword, err := config.ReadCredentialsFromSecret(ctx, r.client)
		if err != nil {
			return controller.Result{}, err
		}

		if err := config.SetCredentialsEnv(rUsername, rPassword); err != nil {
			return controller.Result{}, err
		}
	}

	return controller.Result{}, nil
}

func (r *ClusterReconciler) postClusterProviderReconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	if cluster.HasAWSIamConfig() {
		if result, err := r.awsIamAuth.Reconcile(ctx, log, cluster); err != nil {
			return controller.Result{}, err
		} else if result.Return() {
			return result, nil
		}
	}

	if err := r.machineHealthCheck.Reconcile(ctx, log, cluster); err != nil {
		return controller.Result{}, err
	}

	return controller.Result{}, nil
}

func (r *ClusterReconciler) packagesReconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	// Self-managed clusters can support curated packages, but that support
	// comes from the CLI at this time.
	if cluster.IsManaged() && cluster.IsPackagesEnabled() {
		if err := r.packagesClient.Reconcile(ctx, log, r.client, cluster); err != nil {
			return controller.Result{}, err
		}
	}

	return controller.Result{}, nil
}

func (r *ClusterReconciler) updateStatus(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	// When EKS-A cluster is fully deleted, we do not need to update the status. Without this check
	// the subsequent patch operations would fail if the status is updated after it is fully deleted.
	if !cluster.DeletionTimestamp.IsZero() && len(cluster.GetFinalizers()) == 0 {
		log.Info("Cluster is fully deleted, skipping cluster status update")
		return nil
	}

	log.Info("Updating cluster status")

	if err := clusters.UpdateClusterStatusForControlPlane(ctx, r.client, cluster); err != nil {
		return errors.Wrap(err, "updating status for control plane")
	}

	if err := clusters.UpdateClusterStatusForWorkers(ctx, r.client, cluster); err != nil {
		return errors.Wrap(err, "updating status for workers")
	}

	clusters.UpdateClusterStatusForCNI(ctx, cluster)

	if err := clusters.UpdateClusterCertificateStatus(ctx, r.client, log, cluster); err != nil {
		return errors.Wrap(err, "updating cluster certificate status for cluster")
	}

	summarizedConditionTypes := []anywherev1.ConditionType{
		anywherev1.ControlPlaneInitializedCondition,
		anywherev1.ControlPlaneReadyCondition,
		anywherev1.WorkersReadyCondition,
	}

	defaultCNIConfiguredCondition := conditions.Get(cluster, anywherev1.DefaultCNIConfiguredCondition)
	if defaultCNIConfiguredCondition == nil ||
		(defaultCNIConfiguredCondition.Status == "False" &&
			defaultCNIConfiguredCondition.Reason != anywherev1.SkipUpgradesForDefaultCNIConfiguredReason) {
		summarizedConditionTypes = append(summarizedConditionTypes, anywherev1.DefaultCNIConfiguredCondition)
	}

	// Always update the readyCondition by summarizing the state of other conditions.
	conditions.SetSummary(cluster,
		conditions.WithConditions(summarizedConditionTypes...),
	)

	return nil
}

func (r *ClusterReconciler) reconcileDelete(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	if cluster.IsSelfManaged() && !cluster.IsManagedByCLI() {
		return ctrl.Result{}, errors.New("deleting self-managed clusters is not supported")
	}

	if cluster.IsReconcilePaused() && !cluster.CanDeleteWhenPaused() {
		log.Info("Cluster reconciliation is paused, won't process cluster deletion")
		return ctrl.Result{}, nil
	}

	// Creates vspheredeploymentzone and vspherefailuredomain CR on bootstrap cluster prior to delete
	// These CRs are not migrated over during pivot and must be present to cleanly delete vspheremachines
	// This solution isn't ideal, but would require redesign
	if cluster.Spec.DatacenterRef.Kind == "VSphereDatacenterConfig" && cluster.IsSelfManaged() {
		if err := r.applyFailureDomains(ctx, log, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	capiCluster := &clusterv1.Cluster{}
	capiClusterName := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: cluster.Name}
	log.Info("Deleting", "name", cluster.Name)
	err := r.client.Get(ctx, capiClusterName, capiCluster)

	switch {
	case err == nil:
		log.Info("Deleting CAPI cluster", "name", capiCluster.Name)
		if err := r.client.Delete(ctx, capiCluster); err != nil {
			log.Info("Error deleting CAPI cluster", "name", capiCluster.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	case apierrors.IsNotFound(err):
		log.Info("Deleting EKS Anywhere cluster", "name", capiCluster.Name, "cluster.DeletionTimestamp", cluster.DeletionTimestamp, "finalizer", cluster.Finalizers)

		// TODO delete GitOps,Datacenter and MachineConfig objects
		controllerutil.RemoveFinalizer(cluster, ClusterFinalizerName)
	default:
		return ctrl.Result{}, err

	}

	if cluster.HasAWSIamConfig() {
		if err := r.awsIamAuth.ReconcileDelete(ctx, log, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	if cluster.IsManaged() {
		if err := r.packagesClient.ReconcileDelete(ctx, log, r.client, cluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("deleting packages for cluster %q: %w", cluster.Name, err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) buildClusterConfig(ctx context.Context, clus *anywherev1.Cluster) (*c.Config, error) {
	builder := c.NewDefaultConfigClientBuilder()
	config, err := builder.Build(ctx, clientutil.NewKubeClient(r.client), clus)
	if err != nil {
		var notFound apierrors.APIStatus
		if apierrors.IsNotFound(err) && errors.As(err, &notFound) {
			failureMessage := fmt.Sprintf("Dependent cluster objects don't exist: %s", notFound)
			clus.SetFailure(anywherev1.MissingDependentObjectsReason, failureMessage)
		}
		return nil, err
	}

	return config, nil
}

func (r *ClusterReconciler) applyFailureDomains(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	return r.vSpherefailureDomainMover.ApplyFailureDomains(ctx, log, cluster)
}

// ApplyFailureDomains orchestrates the end-to-end process of applying failure domain objects to a cluster.
func (m *FailureDomainMover) ApplyFailureDomains(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	clusterSpec, err := m.specBuilder.BuildSpec(ctx, cluster)
	if err != nil {
		return err
	}

	// Check if VSphereDatacenter has no failure domains
	if len(clusterSpec.VSphereDatacenter.Spec.FailureDomains) == 0 {
		return nil
	}

	log.Info("Creating vspheredeploymentzones and vspherefailuredomains on bootstrap")
	fd, err := m.fdSpecBuilder.BuildFailureDomainSpec(log, clusterSpec)
	if err != nil {
		return err
	}

	return m.objectReconciler.ReconcileObjects(ctx, fd)
}

func (r *ClusterReconciler) ensureClusterOwnerReferences(ctx context.Context, clus *anywherev1.Cluster, config *c.Config) error {
	for _, obj := range config.ChildObjects() {
		numberOfOwnerReferences := len(obj.GetOwnerReferences())
		if err := controllerutil.SetOwnerReference(clus, obj, r.client.Scheme()); err != nil {
			return errors.Wrapf(err, "setting cluster owner reference for %s", obj.GetObjectKind())
		}

		if numberOfOwnerReferences == len(obj.GetOwnerReferences()) {
			// obj already had the owner reference
			continue
		}

		if err := r.client.Update(ctx, obj); err != nil {
			return errors.Wrapf(err, "updating object (%s) with cluster owner reference", obj.GetObjectKind())
		}
	}

	return nil
}

func patchCluster(ctx context.Context, patchHelper *patch.Helper, cluster *anywherev1.Cluster, patchOpts ...patch.Option) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	options := append([]patch.Option{
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			// Add each condition her that the controller should ignored conflicts for.
			anywherev1.ReadyCondition,
			anywherev1.ControlPlaneInitializedCondition,
			anywherev1.ControlPlaneReadyCondition,
			anywherev1.WorkersReadyCondition,
			anywherev1.DefaultCNIConfiguredCondition,
		}},
	}, patchOpts...)

	// Always attempt to patch the object and status after each reconciliation.
	return patchHelper.Patch(ctx, cluster, options...)
}

// aggregatedGeneration computes the combined generation of the resources linked
// by the cluster by summing up the .metadata.generation value for all the child
// objects of this cluster.
func aggregatedGeneration(config *c.Config) int64 {
	var aggregatedGeneration int64
	for _, obj := range config.ChildObjects() {
		aggregatedGeneration += obj.GetGeneration()
	}

	return aggregatedGeneration
}

func getManagementCluster(ctx context.Context, clus *anywherev1.Cluster, client client.Client) (*anywherev1.Cluster, error) {
	mgmtCluster, err := clusters.FetchManagementEksaCluster(ctx, client, clus)
	if apierrors.IsNotFound(err) {
		clus.SetFailure(
			anywherev1.ManagementClusterRefInvalidReason,
			fmt.Sprintf("Management cluster %s does not exist", clus.Spec.ManagementCluster.Name),
		)
	}
	if err != nil {
		return nil, err
	}

	return mgmtCluster, nil
}

func (r *ClusterReconciler) setDefaultBundlesRefOrEksaVersion(ctx context.Context, clus *anywherev1.Cluster) error {
	mgmtCluster, err := getManagementCluster(ctx, clus, r.client)
	if err != nil {
		return err
	}

	if mgmtCluster.Spec.EksaVersion != nil {
		clus.Spec.EksaVersion = mgmtCluster.Spec.EksaVersion
		return nil
	}

	if mgmtCluster.Spec.BundlesRef != nil {
		clus.Spec.BundlesRef = mgmtCluster.Spec.BundlesRef
		return nil
	}

	clus.Status.FailureMessage = ptr.String("Management cluster must have either EksaVersion or BundlesRef")
	return fmt.Errorf("could not set default values")
}

func validateEksaRelease(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) error {
	if cluster.Spec.EksaVersion == nil {
		return nil
	}
	err := validations.ValidateEksaReleaseExistOnManagement(ctx, clientutil.NewKubeClient(client), cluster)
	if apierrors.IsNotFound(err) {
		errMsg := fmt.Sprintf("eksarelease %v could not be found on the management cluster", *cluster.Spec.EksaVersion)
		reason := anywherev1.EksaVersionInvalidReason
		cluster.Status.FailureMessage = ptr.String(errMsg)
		cluster.Status.FailureReason = &reason
		return err
	} else if err != nil {
		return err
	}
	return nil
}

func validateExtendedKubernetesVersionSupport(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) error {
	eksaVersion := cluster.Spec.EksaVersion
	if cluster.Spec.DatacenterRef.Kind == "SnowDatacenterConfig" || eksaVersion == nil {
		return nil
	}
	skip, err := validations.ShouldSkipBundleSignatureValidation((*string)(eksaVersion))
	if err != nil {
		return err
	}

	// Skip the signature validation for those versions prior to 'v0.22.0'
	if skip {
		return nil
	}

	bundle, err := c.BundlesForCluster(ctx, clientutil.NewKubeClient(client), cluster)
	if err != nil {
		reason := anywherev1.BundleNotFoundReason
		cluster.Status.FailureMessage = ptr.String(err.Error())
		cluster.Status.FailureReason = &reason
		return fmt.Errorf("getting bundle for cluster: %w", err)
	}

	// Get the release manifest using Kubernetes client
	releaseManifest, err := getReleaseManifestFromCluster(ctx, *cluster, bundle, clientutil.NewKubeClient(client))
	if err != nil {
		reason := anywherev1.ExtendedK8sVersionSupportNotSupportedReason
		cluster.Status.FailureMessage = ptr.String(fmt.Sprintf("getting release manifest: %v", err))
		cluster.Status.FailureReason = &reason
		return fmt.Errorf("getting release manifest: %w", err)
	}

	if err = validations.ValidateExtendedK8sVersionSupport(ctx, *cluster, bundle, releaseManifest, clientutil.NewKubeClient(client)); err != nil {
		reason := anywherev1.ExtendedK8sVersionSupportNotSupportedReason
		cluster.Status.FailureMessage = ptr.String(err.Error())
		cluster.Status.FailureReason = &reason
		return err

	}
	return nil
}

// getReleaseManifestFromCluster retrieves the EKS Distro release manifest using the Kubernetes client.
// This is used in the controller context where we always use the Kubernetes client to fetch the manifest.
func getReleaseManifestFromCluster(ctx context.Context, clusterSpec anywherev1.Cluster, bundle *v1alpha1.Bundles, k *clientutil.KubeClient) (*eksdv1alpha1.Release, error) {
	versionsBundle, err := c.GetVersionsBundle(clusterSpec.Spec.KubernetesVersion, bundle)
	if err != nil {
		return nil, fmt.Errorf("getting versions bundle for %s kubernetes version: %w", clusterSpec.Spec.KubernetesVersion, err)
	}

	releaseManifest := &eksdv1alpha1.Release{}
	if err := k.Get(ctx, versionsBundle.EksD.Name, constants.EksaSystemNamespace, releaseManifest); err != nil {
		return nil, fmt.Errorf("getting %s eks distro release: %w", versionsBundle.EksD.Name, err)
	}

	return releaseManifest, nil
}
