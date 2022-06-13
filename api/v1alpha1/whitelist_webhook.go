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

package v1alpha1

import (
	"context"
	"fmt"
	"whitelist-operator/pkg/executor"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var (
	whitelistlog = logf.Log.WithName("whitelist-resource")
	cli          client.Client
)

func (r *Whitelist) SetupWebhookWithManager(mgr ctrl.Manager) error {
	cli = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-apps-whitelist-fly-io-v1alpha1-whitelist,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.whitelist.fly.io,resources=whitelists,verbs=create;update,versions=v1alpha1,name=mwhitelist.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Whitelist{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Whitelist) Default() {
	whitelistlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.Annotations == nil {
		whitelistlog.Info("new default")

		r.Spec.Annotations = make(map[string]string)
	}
	executor.Exec(executor.DEF, r.Spec.Provider, r.Spec.Service, r.Spec.Annotations)
	whitelistlog.Info("get wls", "instance", r)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-apps-whitelist-fly-io-v1alpha1-whitelist,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.whitelist.fly.io,resources=whitelists,verbs=create;update,versions=v1alpha1,name=vwhitelist.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Whitelist{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Whitelist) ValidateCreate() error {
	whitelistlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	//param validation
	param := r.Spec.DeepCopy().Annotations
	err := executor.Exec(executor.CHECK, r.Spec.Provider, r.Spec.Service, param)
	if err != nil {
		return err
	}

	//unique
	var wls WhitelistList
	ctx := context.Background()
	err = cli.List(ctx, &wls)
	if err != nil {
		whitelistlog.Error(err, "get wls failed")
		return err
	}

	for _, ins := range wls.Items {
		if ins.Spec.Provider == r.Spec.Provider && ins.Spec.Service == r.Spec.Service && ins.Spec.ServiceId == r.Spec.ServiceId && ins.Spec.Level == r.Spec.Level {
			same := executor.Duplicate(r.Spec.Provider, r.Spec.Service, r.Spec.Annotations, ins.Spec.Annotations)
			if same {
				return fmt.Errorf("the config is same to %s", ins.Name)
			}
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Whitelist) ValidateUpdate(old runtime.Object) error {
	whitelistlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Whitelist) ValidateDelete() error {
	whitelistlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
