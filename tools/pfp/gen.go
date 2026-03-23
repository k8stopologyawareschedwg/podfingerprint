// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/k8stopologyawareschedwg/podfingerprint"
)

func generatePods(seed int64, numNamespaces, numPods int) []podfingerprint.NamespacedName {
	if numNamespaces <= 0 || numPods <= 0 {
		return []podfingerprint.NamespacedName{}
	}
	rng := rand.New(rand.NewSource(seed))
	if numNamespaces == 1 {
		ns := dictNamespaces[rng.Intn(len(dictNamespaces))]
		return generateSingleNamespacePods(rng, ns, numPods)
	}

	pods := []podfingerprint.NamespacedName{}
	for numPods > 0 {
		// we wxpect at least 3 pods per namespace; rand.Intn() returns in the range [0,n)
		upperBound := numPods
		if upperBound > 2 {
			upperBound = 2
		}
		cnt := 1 + rng.Intn(upperBound)
		numPods -= cnt
		if numPods < 0 {
			numPods = 0
		}
		ns := dictNamespaces[rng.Intn(len(dictNamespaces))]
		pods = append(pods, generateSingleNamespacePods(rng, ns, cnt)...)
	}
	sort.Slice(pods, func(i, j int) bool {
		if pods[i].Namespace != pods[j].Namespace {
			return pods[i].Namespace < pods[j].Namespace
		}
		return pods[i].Name < pods[j].Name
	})
	return pods
}

// k8s.io/apimachinery/pkg/util/rand uses this alphabet for GenerateName suffixes:
// lowercase consonants + digits 2-9 (no vowels to avoid real words, no 0/1 to avoid confusion).
const k8sNameSuffixAlphabet = "bcdfghjklmnpqrstvwxz2456789"

func k8sRandSuffix(rng *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = k8sNameSuffixAlphabet[rng.Intn(len(k8sNameSuffixAlphabet))]
	}
	return string(b)
}

func k8sGenerateName(rng *rand.Rand, base string) string {
	return fmt.Sprintf("%s-%s", base, k8sRandSuffix(rng, 5))
}

// Common English words suitable as Kubernetes namespaces.
var dictNamespaces = []string{
	"default", "production", "staging", "development", "monitoring",
	"logging", "storage", "network", "security", "database",
	"frontend", "backend", "middleware", "analytics", "messaging",
	"kube-system", "kube-public", "service-mesh", "cert-manager", "ingress",
}

// Common English words suitable as base names for pods/deployments.
var dictBaseNames = []string{
	"server", "worker", "controller", "scheduler", "proxy",
	"gateway", "collector", "exporter", "importer", "processor",
	"handler", "manager", "watcher", "dispatcher", "aggregator",
	"validator", "transformer", "publisher", "subscriber", "consumer",
	"nginx", "redis", "postgres", "mongodb", "elasticsearch",
	"prometheus", "grafana", "jaeger", "fluentd", "envoy",
}

func generateSingleNamespacePods(rng *rand.Rand, namespace string, numPods int) []podfingerprint.NamespacedName {
	pods := make([]podfingerprint.NamespacedName, numPods)
	for i := range pods {
		base := dictBaseNames[rng.Intn(len(dictBaseNames))]
		pods[i] = podfingerprint.NamespacedName{
			Namespace: namespace,
			Name:      k8sGenerateName(rng, base),
		}
	}
	return pods
}
