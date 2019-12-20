package clusterconnection

import (
	"context"
	"fmt"

	"github.com/tigera/operator/pkg/controller/installation"
	"github.com/tigera/operator/pkg/controller/status"
	"github.com/tigera/operator/pkg/controller/utils"
	"github.com/tigera/operator/pkg/render"

	operatorv1 "github.com/tigera/operator/pkg/apis/operator/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("clusterconnection_controller")

// Add creates a new ManagementClusterConnection Controller and adds it to the Manager. The Manager will set fields on the Controller
// and start it when the Manager is started.
func Add(mgr manager.Manager, p operatorv1.Provider, tsee bool) error {
	if !tsee {
		// No need to start this controller.
		return nil
	}
	return add(mgr, newReconciler(mgr, p))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, p operatorv1.Provider) reconcile.Reconciler {
	return &ReconcileConnection{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Provider: p,
	}
}

// add adds a new controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterconnection-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return fmt.Errorf("failed to create clusterconnection-controller: %v", err)
	}

	// Watch for changes to primary resource ManagementClusterConnection
	err = c.Watch(&source.Kind{Type: &operatorv1.ManagementClusterConnection{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return fmt.Errorf("clusterconnection-controller failed to watch primary resource: %v", err)
	}

	// Watch for changes to the secrets associated with the ManagementClusterConnection.
	if err = utils.AddSecretsWatch(c, render.GuardianSecretName, render.OperatorNamespace()); err != nil {
		return fmt.Errorf("clusterconnection-controller failed to watch Secret resource %s: %v", render.GuardianSecretName, err)
	}

	return nil
}

// blank assignment to verify that ReconcileConnection implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileConnection{}

// ReconcileConnection reconciles a ManagementClusterConnection object
type ReconcileConnection struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	Provider operatorv1.Provider
}

// Reconcile reads that state of the cluster for a ManagementClusterConnection object and makes changes based on the
// state read and what is in the ManagementClusterConnection.Spec. The Controller will requeue the Request to be
// processed again if the returned error is non-nil or Result.Requeue is true, otherwise upon completion it will
// remove the work from the queue.
func (r *ReconcileConnection) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling the management cluster connection")
	ctx := context.Background()

	// Fetch the managementClusterConnection.
	mcc := &operatorv1.ManagementClusterConnection{}
	err := r.Client.Get(ctx, utils.DefaultTSEEInstanceKey, mcc)
	if err != nil {
		if errors.IsNotFound(err) {
			// We do not want to show errors if this CR is not present.
			// However we do want to cleanup Guardian related resources.
			return CleanupGuardian(ctx, r)
		}
		return reconcile.Result{}, err
	}

	addr := mcc.Spec.ManagementClusterAddr
	if err = utils.ValidateClusterAddr(addr); err != nil {
		return reconcile.Result{}, err
	}

	//We should create the Guardian deployment.
	return CreateOrModifyGuardian(ctx, r, mcc)
}

func CleanupGuardian(ctx context.Context, r *ReconcileConnection) (reconcile.Result, error) {
	// Create a dummy, so we can use it to delete resources.
	guardian, err := render.Guardian(
		"",
		[]*v1.Secret{},
		"",
		r.Provider == operatorv1.ProviderNone,
		"",
	)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Instantiate a dummy guardian that helps us remove objects in the opposite order of their creation.
	g := guardian.(*render.GuardianComponent)

	err = r.Client.Delete(ctx, g.Deployment())
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	err = r.Client.Delete(ctx, g.ClusterRoleBinding())
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	err = r.Client.Delete(ctx, g.ClusterRole())
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	err = r.Client.Delete(ctx, g.ConfigMap())
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	err = r.Client.Delete(ctx, g.ServiceAccount())
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	// Delete secret that may have been copied over by the operator.
	sec := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      render.GuardianSecretName,
			Namespace: render.GuardianNamespace,
		},
	}
	err = r.Client.Delete(ctx, sec)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	err = r.Client.Delete(ctx, g.Namespace())
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func CreateOrModifyGuardian(ctx context.Context, r *ReconcileConnection, mcc *operatorv1.ManagementClusterConnection) (reconcile.Result, error) {
	instl, err := installation.GetInstallation(context.Background(), r.Client, r.Provider)
	if err != nil {
		return reconcile.Result{}, err
	}

	clusterName, err := utils.ClusterName(ctx, r.Client)
	if err != nil {
		log.Error(err, "Failed to get the cluster name")
		return reconcile.Result{}, err
	}

	pullSecrets, err := utils.GetNetworkingPullSecrets(instl, r.Client)
	if err != nil {
		log.Error(err, "Error with Pull secrets")
		return reconcile.Result{}, err
	}

	// Copy the secret from the operator namespace to the guardian namespace if it is present.
	err = copyGuardianSecret(ctx, r.Client)
	if err != nil {
		log.Error(err, "Failed to copy the guardian secret to the tigera-guardian namespace")
		return reconcile.Result{}, err
	}

	handler := utils.NewComponentHandler(log, r.Client, r.Scheme, mcc)
	component, err := render.Guardian(
		utils.FormatManagementClusterAddr(mcc.Spec.ManagementClusterAddr),
		pullSecrets,
		clusterName,
		r.Provider == operatorv1.ProviderOpenShift,
		instl.Spec.Registry,
	)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := handler.CreateOrUpdate(ctx, component, &status.StatusManager{}); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func copyGuardianSecret(ctx context.Context, client client.Client) error {
	opSec := &corev1.Secret{}
	if err := client.Get(ctx, types.NamespacedName{Name: render.GuardianSecretName, Namespace: render.OperatorNamespace()}, opSec); err != nil {
		if errors.IsNotFound(err) {
			// Nothing to copy.
			return nil
		}
		return err
	}
	guarSec := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Name: render.GuardianSecretName, Namespace: render.GuardianNamespace}, guarSec)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	guarSec.Data = opSec.Data
	if err != nil {
		// We create the secret if it does not exist yet
		guarSec.Name = render.GuardianSecretName
		guarSec.Namespace = render.GuardianNamespace
		return client.Create(ctx, guarSec)
	} else {
		// Otherwise, we update it.
		return client.Update(ctx, guarSec)
	}
}
