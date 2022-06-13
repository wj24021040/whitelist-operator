package utils

import (
	"strings"
	appsv1alpha1 "whitelist-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func AnalyseWhitelistByUnderlay(instance *appsv1alpha1.Whitelist, plist *corev1.PodList) (add, del map[string]string) {
	del = instance.Status.DeepCopy().Created
	add = map[string]string{}

	for _, p := range plist.Items {
		p := p
		if p.Status.PodIP == "" {
			continue
		}
		key := "pod/" + p.GetNamespace() + "/" + p.GetName()
		if _, ok := del[key]; !ok {
			add[key] = p.Status.PodIP
			continue
		}

		delete(del, key)
	}
	return
}

//for node maybe Duplicate
func AnalyseWhitelistByOverlay(instance *appsv1alpha1.Whitelist, plist *corev1.PodList) (add, del map[string]string) {
	del = instance.Status.DeepCopy().Created
	add = map[string]string{}
	for _, p := range plist.Items {
		p := p
		if p.Status.HostIP == "" {
			continue
		}
		key := "Node/" + p.Spec.NodeName
		add[key] = p.Status.HostIP //duplicate removal
	}

	for k, _ := range add {
		if _, ok := del[k]; ok {
			delete(add, k)
			delete(del, k)
		}
	}

	return
}

//map[pod]ip----> ip1,ip2,ip3
func Map2IpString(src map[string]string) string {
	ips := []string{}
	for _, v := range src {
		ips = append(ips, v)
	}
	return strings.Join(ips, ",")
}
