// The code in this file manages logstorage for the "Managed" cluster type. It sets up the service required to communicate
// with the Elasticsearch in the management cluster
package logstorage

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	operatorv1 "github.com/tigera/operator/pkg/apis/operator/v1"
	"github.com/tigera/operator/pkg/controller/utils"
	"github.com/tigera/operator/pkg/render"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileManaged sets up the ExternalService required for the other components to communicate with the Elasticsearch
// in the management cluster. If the LogStorage CR still exists then the old Elasticsearch service likely exists, so an
// error is returned.
func (r *ReconcileLogStorage) reconcileManaged(ctx context.Context, network *operatorv1.Installation, reqLogger logr.Logger) (reconcile.Result, error) {
	if _, err := GetLogStorage(ctx, r.client); err == nil {
		return reconcile.Result{}, fmt.Errorf("cluster type is Managed but logstorage still exists")
	} else if !errors.IsNotFound(err) {
		r.status.SetDegraded("Failed to get LogStorage CR", err.Error())
		return reconcile.Result{}, err
	}

	hdler := utils.NewComponentHandler(log, r.client, r.scheme, network)
	component := render.ElasticsearchManaged(r.localDNS, r.provider)
	if err := hdler.CreateOrUpdate(ctx, component, r.status); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
