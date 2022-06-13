package controllers

import (
	"fmt"
	"strings"

	//. "operator/pkg/executor"
	"github.com/go-logr/logr"
	"github.com/wj24021040/tools/set/hashSet"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type podEventHandle struct {
	client.Client
	L *logr.Logger
}

var _ handler.EventHandler = podEventHandle{}

func (t podEventHandle) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	return
}
func (t podEventHandle) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if evt.Object == nil {
		return
	}
	labDel := evt.Object.GetLabels()
	if labDel == nil {
		return
	}

	for k, _ := range labDel {
		if strings.Contains(k, WLLPre) {
			tmp := strings.Split(k, "/")
			if len(tmp) == 2 {
				namespacedName := types.NamespacedName{Name: tmp[1]}
				q.Add(reconcile.Request{NamespacedName: namespacedName})
			}
		}
	}

}

func (t podEventHandle) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if evt.ObjectOld == nil || evt.ObjectNew == nil {
		return
	}
	fmt.Println("pod update")
	labOld := evt.ObjectOld.GetLabels()
	labNew := evt.ObjectNew.GetLabels()
	if labOld == nil && labNew == nil {
		return
	}
	fmt.Println("pod update0, ", labOld, labNew)
	var whiteNameOld = hashSet.New()
	var whiteNameNew = hashSet.New()
	var ok bool
	if labOld != nil {
		for k, _ := range labOld {
			if strings.Contains(k, WLLPre) {
				tmp := strings.Split(k, "/")
				if len(tmp) == 2 {
					whiteNameOld.Add(tmp[1])
				}
			}
		}
	}

	if labNew != nil {
		for k, _ := range labNew {
			if strings.Contains(k, WLLPre) {
				tmp := strings.Split(k, "/")
				if len(tmp) == 2 {
					whiteNameNew.Add(tmp[1])
				}
			}
		}
	}

	if whiteNameOld.Cap() == 0 && whiteNameNew.Cap() == 0 {
		return
	}

	podOld, ok := evt.ObjectOld.(*corev1.Pod)
	if !ok {
		return
	}
	PodIpOld := podOld.Status.PodIP

	podNew, ok := evt.ObjectNew.(*corev1.Pod)
	if !ok {
		return
	}
	PodIpNew := podNew.Status.PodIP

	fmt.Println("pod update1 ：", PodIpOld, PodIpNew)

	if PodIpNew == PodIpOld { // labels changed
		changeWhistlist := whiteNameOld.Difference(whiteNameNew)
		fmt.Println("changeWhistlist: ", changeWhistlist.String())
		changeLableAction := func(name interface{}) bool {
			namespacedName := types.NamespacedName{Name: name.(string)}
			q.Add(reconcile.Request{NamespacedName: namespacedName})
			return true
		}
		changeWhistlist.Each(changeLableAction)
	} else {
		unionWhistlist := whiteNameOld.Union(whiteNameNew)
		fmt.Println("unionWhistlist: ", unionWhistlist.String())
		changePodAction := func(name interface{}) bool {
			namespacedName := types.NamespacedName{Name: name.(string)}
			q.Add(reconcile.Request{NamespacedName: namespacedName})
			return true
		}
		unionWhistlist.Each(changePodAction)
	}

	return
}

func (t podEventHandle) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}
