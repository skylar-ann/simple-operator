package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	simplev1alpha1 "github.worldpay.com/Atlas/simple-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("SimpleReconciler", func() {

	Context("when reconciling a Simple with no starting spec", func() {
		var (
			ctx             = context.Background()
			simpleObjectKey = client.ObjectKey{Name: "simple", Namespace: "basic-test"}

			namespace = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: simpleObjectKey.Namespace}}
			simple    = &simplev1alpha1.Simple{
				ObjectMeta: metav1.ObjectMeta{
					Name:      simpleObjectKey.Name,
					Namespace: simpleObjectKey.Namespace,
				},
			}
		)

		Specify("a Simple resource exists", func() {
			Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
			Expect(k8sClient.Create(ctx, simple)).To(Succeed())
		})

		It("should create a configmap with no data", func() {
			cm := &corev1.ConfigMap{}

			Eventually(func() map[string]string {
				k8sClient.Get(ctx, simpleObjectKey, cm)
				return cm.Data
			}, time.Second*3, time.Millisecond*500).Should(BeEmpty())
		})

		It("should reconcile the configmap when foo is added", func() {
			simple.Spec.Foo = "newdata"

			Eventually(func() error {
				return k8sClient.Update(ctx, simple)
			}, time.Second*3, time.Millisecond*500).Should(Succeed())

			cm := &corev1.ConfigMap{}

			Eventually(func() map[string]string {
				k8sClient.Get(ctx, simpleObjectKey, cm)
				return cm.Data
			}, time.Second*3, time.Millisecond*500).Should(
				HaveKeyWithValue("something.conf", fmt.Sprintf("setting = %s", simple.Spec.Foo)),
			)
		})

		It("should reconcile the configmap when it is tampered", func() {
			cm := &corev1.ConfigMap{}

			Eventually(func() error {
				return k8sClient.Get(ctx, simpleObjectKey, cm)
			}, time.Second*3, time.Millisecond*500).Should(Succeed())

			// Tamper Simple's configmap
			cm.Data["something.conf"] = "setting = tampered"
			Eventually(func() error {
				return k8sClient.Update(ctx, cm)
			}, time.Second*3, time.Millisecond*500).Should(Succeed())

			// The value should be restored automatically
			Eventually(func() map[string]string {
				k8sClient.Get(ctx, simpleObjectKey, cm)
				return cm.Data
			}, time.Second*3, time.Millisecond*500).Should(
				HaveKeyWithValue("something.conf", fmt.Sprintf("setting = %s", simple.Spec.Foo)),
			)
		})

		It("should reconcile the configmap when it is deleted", func() {
			cm := &corev1.ConfigMap{}

			Eventually(func() error {
				return k8sClient.Get(ctx, simpleObjectKey, cm)
			}, time.Second*3, time.Millisecond*500).Should(Succeed())

			// Delete Simple's configmap
			Eventually(func() error {
				return k8sClient.Delete(ctx, cm)
			}, time.Second*3, time.Millisecond*500).Should(Succeed())

			// Simple's configmap should reconcile
			Eventually(func() map[string]string {
				k8sClient.Get(ctx, simpleObjectKey, cm)
				return cm.Data
			}, time.Second*3, time.Millisecond*500).Should(
				HaveKeyWithValue("something.conf", fmt.Sprintf("setting = %s", simple.Spec.Foo)),
			)
		})

		It("should have an empty configmap when simple Foo is empty", func() {
			simple.Spec.Foo = ""

			Eventually(func() error {
				return k8sClient.Update(ctx, simple)
			}, time.Second*3, time.Millisecond*500).Should(Succeed())

			cm := &corev1.ConfigMap{}

			Eventually(func() map[string]string {
				k8sClient.Get(ctx, simpleObjectKey, cm)
				return cm.Data
			}, time.Second*3, time.Millisecond*500).Should(BeEmpty())
		})
	})
})
