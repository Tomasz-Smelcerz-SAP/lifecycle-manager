package manifest_controller_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	declarative "github.com/kyma-project/lifecycle-manager/internal/declarative/v2"
	"github.com/kyma-project/lifecycle-manager/internal/manifest"
	"github.com/kyma-project/lifecycle-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Manifest readiness check", Ordered, func() {
	customDir := "custom-dir"
	installName := filepath.Join(customDir, "installs")
	It(
		"setup OCI", func() {
			PushToRemoteOCIRegistry(installName)
		},
	)
	BeforeEach(
		func() {
			Expect(os.RemoveAll(filepath.Join(os.TempDir(), customDir))).To(Succeed())
		},
	)
	It("Install OCI specs including an nginx deployment", func() {
		testManifest := NewTestManifest("custom-check-oci")
		manifestName := testManifest.GetName()
		validImageSpec := createOCIImageSpec(installName, server.Listener.Addr().String(), false)
		imageSpecByte, err := json.Marshal(validImageSpec)
		Expect(err).ToNot(HaveOccurred())
		Expect(installManifest(testManifest, imageSpecByte, true)).To(Succeed())

		Eventually(expectManifestStateIn(declarative.StateReady), standardTimeout, standardInterval).
			WithArguments(manifestName).Should(Succeed())

		testClient, err := declarativeTestClient()
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that deployment and Sample CR are deployed and ready")
		deploy := &appsv1.Deployment{}
		Eventually(setDeploymentStatus(deploy), standardTimeout, standardInterval).Should(Succeed())
		sampleCR := emptySampleCR()
		Eventually(setCRStatus(sampleCR, declarative.StateReady), standardTimeout, standardInterval).Should(Succeed())

		By("Verifying manifest status list all resources correctly")
		status, err := getManifestStatus(manifestName)
		Expect(err).ToNot(HaveOccurred())
		Expect(status.Synced).To(HaveLen(2))

		expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
		expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
		Expect(status.Synced).To(ContainElement(expectedDeployment))
		Expect(status.Synced).To(ContainElement(expectedCRD))

		By("Preparing resources for the CR readiness check")
		resources, err := prepareResourceInfosForCustomCheck(testClient, deploy)
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).ToNot(BeEmpty())

		By("Executing the CR readiness check")
		customReadyCheck := manifest.NewCustomResourceReadyCheck()
		stateInfo, err := customReadyCheck.Run(ctx, testClient, testManifest, resources)

		Expect(err).NotTo(HaveOccurred())
		Expect(stateInfo.State).To(Equal(declarative.StateReady))

		By("cleaning up the manifest")
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeTrue())

		Eventually(deleteManifestAndVerify(testManifest), standardTimeout, standardInterval).Should(Succeed())

		By("verify target resources got deleted")
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
	})
})

var _ = Describe("Manifest warning check", Ordered, func() {
	customDir := "custom-dir"
	installName := filepath.Join(customDir, "installs")
	It(
		"setup OCI", func() {
			PushToRemoteOCIRegistry(installName)
		},
	)
	BeforeEach(
		func() {
			Expect(os.RemoveAll(filepath.Join(os.TempDir(), customDir))).To(Succeed())
		},
	)
	It("Install OCI specs including an nginx deployment", func() {
		By("Install test Manifest CR")
		testManifest := NewTestManifest("custom-check-oci")
		manifestName := testManifest.GetName()
		validImageSpec := createOCIImageSpec(installName, server.Listener.Addr().String(), false)
		imageSpecByte, err := json.Marshal(validImageSpec)
		Expect(err).ToNot(HaveOccurred())
		Expect(installManifest(testManifest, imageSpecByte, true)).To(Succeed())

		By("Ensure that deployment and Sample CR are deployed and ready")
		deploy := &appsv1.Deployment{}
		Eventually(setDeploymentStatus(deploy), standardTimeout, standardInterval).Should(Succeed())
		sampleCR := emptySampleCR()
		Eventually(setCRStatus(sampleCR, declarative.StateReady), standardTimeout, standardInterval).Should(Succeed())

		By("Verify the Manifest CR is in the \"Ready\" state")
		Eventually(expectManifestStateIn(declarative.StateReady), standardTimeout, standardInterval).
			WithArguments(manifestName).Should(Succeed())

		/*
			un2 := sampleCR.DeepCopy()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(un2), un2)
			Expect(err).To(BeNil())
			dumpAsJson(">>", un2)
		*/

		By("Verify manifest status list all resources correctly")
		status, err := getManifestStatus(manifestName)
		Expect(err).ToNot(HaveOccurred())
		Expect(status.Synced).To(HaveLen(2))
		expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
		expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
		Expect(status.Synced).To(ContainElement(expectedDeployment))
		Expect(status.Synced).To(ContainElement(expectedCRD))

		By("When the Module CR state is changed to \"Warning\"")
		Eventually(setCRStatus(sampleCR, declarative.StateWarning), standardTimeout, standardInterval).Should(Succeed())

		By("Verify the Manifest CR state also changes to \"Warning\"")
		Eventually(expectManifestStateIn(declarative.StateWarning), standardTimeout, standardInterval).
			WithArguments(manifestName).Should(Succeed())
	})
})

func verifyObjectExists(obj *unstructured.Unstructured) func() (bool, error) {
	return func() (bool, error) {

		err := k8sClient.Get(
			ctx, client.ObjectKeyFromObject(obj),
			obj,
		)

		if err == nil {
			return true, nil
		} else if util.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}
}

func asResource(name, namespace, group, version, kind string) declarative.Resource {
	return declarative.Resource{Name: name, Namespace: namespace,
		GroupVersionKind: metav1.GroupVersionKind{
			Group: group, Version: version, Kind: kind},
	}
}

func emptySampleCR() *unstructured.Unstructured {
	res := &unstructured.Unstructured{}
	res.SetGroupVersionKind(schema.GroupVersionKind{Group: "operator.kyma-project.io", Version: "v1alpha1", Kind: "Sample"})
	res.SetName("sample-crd-from-manifest")
	res.SetNamespace(metav1.NamespaceDefault)
	return res
}

func setCRStatus(cr *unstructured.Unstructured, statusValue declarative.State) func() error {
	return func() error {
		err := k8sClient.Get(
			ctx, client.ObjectKeyFromObject(cr),
			cr,
		)
		if err != nil {
			return err
		}
		unstructured.SetNestedMap(cr.Object, map[string]any{}, "status")
		unstructured.SetNestedField(cr.Object, string(statusValue), "status", "state")
		err = k8sClient.Status().Update(ctx, cr)
		return err
	}
}

func setDeploymentStatus(deploy *appsv1.Deployment) func() error {
	return func() error {
		err := k8sClient.Get(
			ctx, client.ObjectKey{
				Namespace: metav1.NamespaceDefault,
				Name:      "nginx-deployment",
			}, deploy,
		)
		if err != nil {
			return err
		}
		deploy.Status.Replicas = *deploy.Spec.Replicas
		deploy.Status.ReadyReplicas = *deploy.Spec.Replicas
		deploy.Status.AvailableReplicas = *deploy.Spec.Replicas
		deploy.Status.Conditions = append(deploy.Status.Conditions,
			appsv1.DeploymentCondition{
				Type:   appsv1.DeploymentAvailable,
				Status: corev1.ConditionTrue,
			})
		err = k8sClient.Status().Update(ctx, deploy)
		if err != nil {
			return err
		}
		return nil
	}
}

func prepareResourceInfosForCustomCheck(clt declarative.Client, deploy *appsv1.Deployment) ([]*resource.Info, error) {
	deployUnstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deploy)
	if err != nil {
		return nil, err
	}
	deployUnstructured := &unstructured.Unstructured{}
	deployUnstructured.SetUnstructuredContent(deployUnstructuredObj)
	deployUnstructured.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))
	deployInfo, err := clt.ResourceInfo(deployUnstructured, true)
	if err != nil {
		return nil, err
	}
	return []*resource.Info{deployInfo}, nil
}

func declarativeTestClient() (declarative.Client, error) {
	cluster := &declarative.ClusterInfo{
		Config: cfg,
		Client: k8sClient,
	}

	return declarative.NewSingletonClients(cluster)
}

func dumpAsJson(prefix string, obj any) {
	objSer, err := json.MarshalIndent(obj, prefix, "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(objSer))
}
