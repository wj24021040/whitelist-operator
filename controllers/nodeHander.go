package controllers

import (
	"fmt"
	//. "operator/pkg/executor"
	"github.com/go-logr/logr"
	//corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	//"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type nodeEventHandle struct {
	client.Client
	L *logr.Logger
}

var _ handler.EventHandler = nodeEventHandle{}

func (t nodeEventHandle) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	return
}
func (t nodeEventHandle) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {

}

func (t nodeEventHandle) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {

	return
}

func (t nodeEventHandle) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	fmt.Println("pod Generic: ", evt.Object)
}
