/*
Copyright 2022.

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

package controllers

import (
	"context"

	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1alpha1 "whitelist-operator/api/v1alpha1"
	"whitelist-operator/controllers/utils"
	"whitelist-operator/pkg/executor"

	"github.com/go-logr/logr"
	//appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	//"k8s.io/apimachinery/pkg/labels"
	//"sigs.k8s.io/controller-runtime/pkg/builder"
	//ctrLog "sigs.k8s.io/controller-runtime/pkg/log"
	//"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	finalizer = "apps.whitelist.fly.io/finalizer"
	WLService = "whitelist" // whitelist = name
	WLLPre    = "whitelist.fly.io"
)

// WhitelistReconciler reconciles a Whitelist object
type WhitelistReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Logr     logr.Logger
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.whitelist.fly.io,resources=whitelists,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.whitelist.fly.io,resources=whitelists/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.whitelist.fly.io,resources=whitelists/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Whitelist object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *WhitelistReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &appsv1alpha1.Whitelist{}
	r.Logr.Info("get", "req", req.NamespacedName)

	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logr.Info("Whitelist resource not found. Ignore this object must be deleted")
			return ctrl.Result{}, nil
		}
		r.Logr.Error(err, "failed to get Whitelist resourc")
		return ctrl.Result{}, err
	}

	//Finalizer,delete
	markedToDelete := instance.GetDeletionTimestamp()
	if markedToDelete != nil {
		if controllerutil.ContainsFinalizer(instance, finalizer) {
			r.Logr.Info("remvoe finalizer for instance", "whistlist", req.NamespacedName)
			if err = r.cleanWhitelist(instance); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(instance, finalizer)
			err = r.Update(ctx, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	} else if !controllerutil.ContainsFinalizer(instance, finalizer) { // Add finalizer for this CR
		r.Logr.Info("add finalizer for instance")
		controllerutil.AddFinalizer(instance, finalizer)

		err = r.Update(ctx, instance)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if instance.Status.Created == nil {
		instance.Status.Created = make(map[string]string)
	}

	key := WLLPre + "/" + req.NamespacedName.Name
	l := map[string]string{key: ""}
	var pods corev1.PodList
	// pod label: whitelist.fly.io/name =
	err = r.List(ctx, &pods, client.MatchingLabels(l))
	if err != nil {
		r.Logr.Error(err, "whitelist get podlist")
		return ctrl.Result{}, err
	}

	var add, del map[string]string
	if instance.Spec.Level == "Pod" {
		add, del = utils.AnalyseWhitelistByUnderlay(instance, &pods)
	} else {
		add, del = utils.AnalyseWhitelistByOverlay(instance, &pods)
	}

	err = r.createWhitelistForPod(instance, add, del)

	if err == nil && len(add) != 0 || len(del) != 0 {
		err = r.Status().Update(ctx, instance)
		if err != nil {
			r.Logr.Error(err, "update Whitelist's status")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WhitelistReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Whitelist{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, podEventHandle{Client: r.Client, L: &(r.Logr)}).
		//Watches(&source.Kind{Type: &corev1.Node{}}, nodeEventHandle{Client: r.Client, L: &(r.Logr)}).
		Complete(r)
}

func (r *WhitelistReconciler) cleanWhitelist(instance *appsv1alpha1.Whitelist) error {
	//namespacedName := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}
	if len(instance.Status.Created) > 0 {
		dels := []string{}
		for _, pst := range instance.Status.Created {
			dels = append(dels, pst)
		}
		param := instance.Spec.DeepCopy().Annotations
		param[executor.REGIP] = strings.Join(dels, ",")
		param[executor.SERVICEID] = instance.Spec.ServiceId
		err := executor.Exec(executor.DEL, instance.Spec.Provider, instance.Spec.Service, param)
		if err != nil {
			r.Logr.Error(err, "cleanWhitelist")
			return err
		}
	}

	return nil
}

func (r *WhitelistReconciler) createWhitelistForPod(instance *appsv1alpha1.Whitelist, addpod, delpod map[string]string) error {
	if len(addpod) == 0 && len(delpod) == 0 {
		return nil
	}

	param := make(map[string]string)
	param[executor.SERVICEID] = instance.Spec.ServiceId
	if instance.Spec.Annotations != nil {
		for k, v := range instance.Spec.Annotations {
			param[k] = v
		}
	}

	if len(addpod) > 0 {
		adds := utils.Map2IpString(addpod)
		param[executor.REGIP] = adds
		r.Logr.Info(adds)
		err := executor.Exec(executor.ADD, instance.Spec.Provider, instance.Spec.Service, param)
		if err != nil {
			r.Logr.Error(err, "add Whitelist", "ip", adds)
			r.Recorder.Event(instance, corev1.EventTypeWarning, "add Whitelist for "+adds, err.Error())
			return err
		}
		for k, v := range addpod {
			instance.Status.Created[k] = v
		}
	}

	r.Logr.Info("get ip result: ", "add: ", addpod, "  del: ", delpod)
	//del the ip
	if len(delpod) > 0 {
		dels := utils.Map2IpString(delpod)
		param[executor.REGIP] = dels
		err := executor.Exec(executor.DEL, instance.Spec.Provider, instance.Spec.Service, param)
		if err != nil {
			r.Logr.Error(err, "del Whitelist", "ip", dels)
			r.Recorder.Event(instance, corev1.EventTypeWarning,
				"delete the whitelist failed for "+dels, err.Error())
			return err
		}
		for k, _ := range delpod {
			delete(instance.Status.Created, k)
		}

	}
	return nil
}
