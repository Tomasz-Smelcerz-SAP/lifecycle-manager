package manifest_controller_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
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
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Manifest readiness check", Ordered, func() {
	customDir := "custom-dir"
	installName := filepath.Join(customDir, "installs")
	deploymentName := "nginx-deployment"
	var imageDigest v1.Hash
	It(
		"setup OCI", func() {
			imageDigest = PushToRemoteOCIRegistry(installName, "../../pkg/test_samples/oci/rendered.yaml")
		},
	)
	BeforeEach(
		func() {
			Expect(os.RemoveAll(filepath.Join(os.TempDir(), customDir))).To(Succeed())
		},
	)
	It("Install OCI specs including an nginx deployment", func() {
		testManifest := NewTestManifest("ready-check")
		manifestName := testManifest.GetName()
		validImageSpec := createOCIImageSpec(installName, server.Listener.Addr().String(), imageDigest, false)
		imageSpecByte, err := json.Marshal(validImageSpec)
		Expect(err).ToNot(HaveOccurred())

		fmt.Println("1111111111111111111111111111111111111111")
		depl, err := listDeployments()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(depl)
		fmt.Println("1111111111111111111111111111111111111111")

		Expect(installManifest(testManifest, imageSpecByte, true)).To(Succeed())

		By("Verifying that the manifest is initially in the Processing state")
		Eventually(expectManifestStateIn(declarative.StateProcessing), standardTimeout, standardInterval).
			WithArguments(manifestName).Should(Succeed())

		By("Verifying manifest status list all resources correctly")
		status, err := getManifestStatus(manifestName)
		Expect(err).ToNot(HaveOccurred())
		Expect(status.Synced).To(HaveLen(2))

		expectedDeployment := asResource(deploymentName, "default", "apps", "v1", "Deployment")
		expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
		Expect(status.Synced).To(ContainElement(expectedDeployment))
		Expect(status.Synced).To(ContainElement(expectedCRD))

		By("Verifying that deployment and Sample CR are deployed and ready")
		deploy := &appsv1.Deployment{}
		Eventually(setDeploymentStatus(deploymentName, deploy), standardTimeout, standardInterval).Should(Succeed())
		sampleCR := emptySampleCR(manifestName)
		Eventually(setCRStatus(sampleCR, declarative.StateReady), standardTimeout, standardInterval).Should(Succeed())

		By("Preparing resources for the CR readiness check")
		testClient, err := declarativeTestClient()
		Expect(err).ToNot(HaveOccurred())
		resources, err := prepareResourceInfosForCustomCheck(testClient, deploy)
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).ToNot(BeEmpty())

		By("Executing the CR readiness check reports the state is Ready")
		customReadyCheck := manifest.NewCustomResourceReadyCheck()
		stateInfo, err := customReadyCheck.Run(ctx, testClient, testManifest, resources)

		Expect(err).NotTo(HaveOccurred())
		Expect(stateInfo.State).To(Equal(declarative.StateReady))

		By("cleaning up the manifest")
		fmt.Println("2222222222222222222222222222222222222222")
		depl, err = listDeployments()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(depl)
		fmt.Println("2222222222222222222222222222222222222222")
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeTrue())

		Eventually(deleteManifestAndVerify(testManifest), standardTimeout, standardInterval).Should(Succeed())

		fmt.Println("3333333333333333333333333333333333333333")
		depl, err = listDeployments()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(depl)
		fmt.Println("3333333333333333333333333333333333333333")

		By("verify target resources got deleted")
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(
			verifyObjectExistsDump(expectedDeployment.ToUnstructured())).
			WithTimeout(standardTimeout).
			WithPolling(standardInterval).
			Should(BeFalse())
	})
})

func listDeployments() ([]string, error) {
	list := appsv1.DeploymentList{}
	err := k8sClient.List(ctx, &list)

	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, dep := range list.Items {
		res = append(res, dep.Name)
	}
	return res, nil
}

var _ = Describe("Manifest warning check", Ordered, func() {
	customDir := "custom-dir"
	installName := filepath.Join(customDir, "installs")
	var imageDigest v1.Hash
	deploymentName := "nginx-deployment-2"

	It(
		"setup OCI", func() {
			imageDigest = PushToRemoteOCIRegistry(installName, "../../pkg/test_samples/oci/rendered.2.yaml")
		},
	)
	BeforeEach(
		func() {
			Expect(os.RemoveAll(filepath.Join(os.TempDir(), customDir))).To(Succeed())
		},
	)
	It("Install OCI specs including an nginx deployment", func() {
		By("Install test Manifest CR")
		testManifest := NewTestManifest("warning-check")
		manifestName := testManifest.GetName()
		validImageSpec := createOCIImageSpec(installName, server.Listener.Addr().String(), imageDigest, false)
		imageSpecByte, err := json.Marshal(validImageSpec)
		Expect(err).ToNot(HaveOccurred())

		fmt.Println("8888888888888888888888888888888888888888")
		depl, err := listDeployments()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(depl)
		fmt.Println("8888888888888888888888888888888888888888")

		Expect(installManifest(testManifest, imageSpecByte, true)).To(Succeed())

		time.Sleep(2 * time.Second)

		fmt.Println("9999999999999999999999999999999999999999")
		depl, err = listDeployments()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(depl)
		fmt.Println("9999999999999999999999999999999999999999")

		By("Ensure that deployment and Sample CR are deployed and ready")
		deploy := &appsv1.Deployment{}
		Eventually(setDeploymentStatus(deploymentName, deploy), standardTimeout, standardInterval).Should(Succeed())
		sampleCR := emptySampleCR(manifestName)
		Eventually(setCRStatus(sampleCR, declarative.StateReady), standardTimeout, standardInterval).Should(Succeed())

		By("Verify the Manifest CR is in the \"Ready\" state")
		Eventually(expectManifestStateIn(declarative.StateReady), standardTimeout, standardInterval).
			WithArguments(manifestName).Should(Succeed())

		By("Verify manifest status list all resources correctly")
		status, err := getManifestStatus(manifestName)
		Expect(err).ToNot(HaveOccurred())
		Expect(status.Synced).To(HaveLen(2))
		expectedDeployment := asResource(deploymentName, "default", "apps", "v1", "Deployment")
		expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
		Expect(status.Synced).To(ContainElement(expectedDeployment))
		Expect(status.Synced).To(ContainElement(expectedCRD))

		By("When the Module CR state is changed to \"Warning\"")
		Eventually(setCRStatus(sampleCR, declarative.StateWarning), standardTimeout, standardInterval).Should(Succeed())

		By("Verify the Manifest CR state also changes to \"Warning\"")
		Eventually(expectManifestStateIn(declarative.StateWarning), standardTimeout, standardInterval).
			WithArguments(manifestName).Should(Succeed())

		By("cleaning up the manifest")
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeTrue())

		Eventually(deleteManifestAndVerify(testManifest), standardTimeout, standardInterval).Should(Succeed())

		By("verify target resources got deleted")
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(
			verifyObjectExists(expectedDeployment.ToUnstructured())).
			WithTimeout(standardTimeout).
			WithPolling(standardInterval).
			Should(BeFalse())
	})
})

func verifyObjectExistsDump(obj *unstructured.Unstructured) func() (bool, error) {
	return func() (bool, error) {
		fmt.Println("1515151515151515151515151515151515151515")
		depl, err := listDeployments()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(depl)
		fmt.Println("1515151515151515151515151515151515151515")

		err = k8sClient.Get(
			ctx, client.ObjectKeyFromObject(obj),
			obj,
		)

		//fmt.Println("========================================")
		//obj.Object["managedFields"] = nil
		//dumpAsJson(" !!!!!!!!! ", obj.Object["metadata"])
		//fmt.Println("========================================")

		if err == nil {
			return true, nil
		} else if util.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}
}

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

func setDeploymentStatus(name string, deploy *appsv1.Deployment) func() error {
	return func() error {
		err := k8sClient.Get(
			ctx, client.ObjectKey{
				Namespace: metav1.NamespaceDefault,
				Name:      name,
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
