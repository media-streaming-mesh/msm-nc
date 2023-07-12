/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	stream_mapper "github.com/media-streaming-mesh/msm-nc/internal/stream-mapper"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mediastreamsv1 "github.com/media-streaming-mesh/msm-nc/api/v1"
	"github.com/sirupsen/logrus"
)

// StreamdataReconciler reconciles a Streamdata object
type StreamdataReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	StreamMapper *stream_mapper.StreamMapper
	Log          *logrus.Logger
	// TODO - Do we need a node mapper or node channel - can watch K8s nodes directly
	// TODO How is config parsed in
}

//+kubebuilder:rbac:groups=mediastreams.media-streaming-mesh.io,resources=streamdata,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mediastreams.media-streaming-mesh.io,resources=streamdata/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mediastreams.media-streaming-mesh.io,resources=streamdata/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Streamdata object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *StreamdataReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	stream := &mediastreamsv1.Streamdata{}
	err := r.Client.Get(ctx, req.NamespacedName, stream)
	r.Log.Infof("Reconcile request %v", req)
	r.Log.Infof("stream object %v", stream)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Object not found, return.
			// Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		// TODO - decide on requeue vs. ignore
		return reconcile.Result{Requeue: true}, err
	}
	// Write dataplane here

	stream.Status.Status = "SUCCESS"
	stream.Status.Reason = "NONE"
	stream.Status.StreamStatus = "PENDING"
	updateErr := r.Status().Update(ctx, stream)
	if updateErr != nil {
		r.Log.Infof("Error Updating status %s", updateErr)
		// Error updating the object - requeue the request.
		return reconcile.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StreamdataReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediastreamsv1.Streamdata{}).
		Complete(r)
}

func (r *StreamdataReconciler) setFailedStatus(ctx context.Context, err error, stream *mediastreamsv1.Streamdata) error {
	stream.Status.Status = "ERROR"
	stream.Status.Reason = err.Error()
	updateErr := r.Status().Update(ctx, stream)
	if updateErr != nil {
		return concatErrors(updateErr, err)
	}
	return err
}

func concatErrors(errs ...error) error {
	errList := []string{}
	for _, v := range errs {
		errList = append(errList, v.Error())
	}
	return fmt.Errorf(strings.Join(errList, ", "))
}
