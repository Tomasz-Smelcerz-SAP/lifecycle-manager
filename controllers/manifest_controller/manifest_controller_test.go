package manifest_controller_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	declarative "github.com/kyma-project/lifecycle-manager/internal/declarative/v2"
	"github.com/kyma-project/lifecycle-manager/pkg/ocmextensions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
)

var ErrManifestStateMisMatch = errors.New("ManifestState mismatch")

/*
var _ = Describe(
	"When Manifest CR contains", func() {
		mainOciTempDir := "main-dir"
		installName := filepath.Join(mainOciTempDir, "installs")

		It(
			"setup OCI", func() {
				PushToRemoteOCIRegistry(installName)
			},
		)
		BeforeEach(
			func() {
				Expect(os.RemoveAll(filepath.Join(os.TempDir(), mainOciTempDir))).To(Succeed())
			},
		)

		It("a valid install OCI image specification, expect state is ready", func() {
			manifest := NewTestManifest("foo")

			expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
			expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
			sampleCR := emptySampleCR(manifest.GetName())

			Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())

			//given
			Eventually(withValidInstallImageSpec(installName, false, false), standardTimeout, standardInterval).
				WithArguments(manifest).
				Should(Succeed())

			Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
			Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())

			//when
			deploy := &appsv1.Deployment{}
			Eventually(setDeploymentStatus("nginx-deployment", deploy), standardTimeout, standardInterval).Should(Succeed())

			//then
			Eventually(expectManifestStateIn(declarative.StateReady), standardTimeout, standardInterval).
				WithArguments(manifest.GetName()).Should(Succeed())

			//cleanup
			Eventually(deleteManifestAndVerify(manifest), standardTimeout, standardInterval).Should(Succeed())

			Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
		})

		It("a valid install OCI image specification and enabled deploy resource, expect state is ready", func() {
			manifest := NewTestManifest("bar")

			expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
			expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
			sampleCR := emptySampleCR(manifest.GetName())

			Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())

			//given
			Eventually(withValidInstallImageSpec(installName, true, false), standardTimeout, standardInterval).
				WithArguments(manifest).
				Should(Succeed())

			Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeTrue())
			Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
			Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())

			//when
			deploy := &appsv1.Deployment{}
			Eventually(setDeploymentStatus("nginx-deployment", deploy), standardTimeout, standardInterval).Should(Succeed())
			Eventually(setCRStatus(sampleCR, declarative.StateReady), standardTimeout, standardInterval).Should(Succeed())

			manifestObj, err := getManifest(manifest.Name)
			Expect(err).ShouldNot(HaveOccurred())

			dumpAsJson(" ----> ", manifestObj)

			crObj, err := getCR(sampleCR.GetName())
			Expect(err).ShouldNot(HaveOccurred())
			dumpAsJson(" ====> ", crObj)

			deploymentObj, err := getDeployment("nginx-deployment")
			Expect(err).ShouldNot(HaveOccurred())
			dumpAsJson(" ~~~~> ", deploymentObj)

			//then
			Eventually(expectManifestStateIn(declarative.StateReady), standardTimeout, standardInterval).
				WithArguments(manifest.GetName()).Should(Succeed())

			//cleanup
			Eventually(deleteManifestAndVerify(manifest), standardTimeout, standardInterval).Should(Succeed())

			Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
			Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
		})
	},
)
*/

var _ = Describe(
	"Rendering manifest install layer", func() {
		mainOciTempDir := "main-dir"
		installName := filepath.Join(mainOciTempDir, "installs")

		It(
			"setup OCI", func() {
				PushToRemoteOCIRegistry(installName)
			},
		)
		BeforeEach(
			func() {
				Expect(os.RemoveAll(filepath.Join(os.TempDir(), mainOciTempDir))).To(Succeed())
			},
		)
		DescribeTable(
			"Test OCI specs",
			func(
				givenCondition func(manifest *v1beta2.Manifest) error,
				expectManifestState func(manifestName string) error,
				objectsShouldExist func(manifestName string, shouldExist bool),
			) {
				manifest := NewTestManifest("oci")

				objectsShouldExist(manifest.GetName(), false)
				fmt.Println("~~~~~~~~~~~~~~~~~~~~ 1")
				fmt.Println(listDeployments())

				Eventually(givenCondition, standardTimeout, standardInterval).
					WithArguments(manifest).Should(Succeed())

				objectsShouldExist(manifest.GetName(), true)
				fmt.Println("~~~~~~~~~~~~~~~~~~~~ 2")
				fmt.Println(listDeployments())

				Eventually(expectManifestState, standardTimeout, standardInterval).
					WithArguments(manifest.GetName()).Should(Succeed())
				Eventually(deleteManifestAndVerify(manifest), standardTimeout, standardInterval).Should(Succeed())

				objectsShouldExist(manifest.GetName(), false)
				fmt.Println("~~~~~~~~~~~~~~~~~~~~ 3")
				fmt.Println(listDeployments())
			},
			Entry(
				"When Manifest CR contains a valid install OCI image specification, "+
					"expect state in ready",
				withValidImageSpecAndWithoutCR(installName),
				expectManifestStateIn(declarative.StateReady),
				deploymentAndCRDShouldExist,
			),
			Entry(
				"When Manifest CR contains a valid install OCI image specification and enabled deploy resource, "+
					"expect state in ready",
				withValidImageSpecAndWithCR(installName),
				expectManifestStateIn(declarative.StateReady),
				deploymentAndCRShouldExist,
			),
			Entry(
				"When Manifest CR contains an invalid install OCI image specification, "+
					"expect state in error",
				withInvalidImageSpec(),
				expectManifestStateIn(declarative.StateError),
				noneShouldExist,
			),
		)
	},
)

func deploymentAndCRDShouldExist(manifestName string, shouldExist bool) {
	expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
	expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
	sampleCR := emptySampleCR(manifestName)

	if shouldExist {
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
	} else {
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
	}
}

func deploymentAndCRShouldExist(manifestName string, shouldExist bool) {
	expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
	expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
	sampleCR := emptySampleCR(manifestName)

	if shouldExist {
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeTrue())
	} else {
		Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
	}
}

func noneShouldExist(manifestName string, shouldExist bool) {
	expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
	expectedCRD := asResource("samples.operator.kyma-project.io", "", "apiextensions.k8s.io", "v1", "CustomResourceDefinition")
	sampleCR := emptySampleCR(manifestName)

	Eventually(verifyObjectExists(sampleCR), standardTimeout, standardInterval).Should(BeFalse())
	Eventually(verifyObjectExists(expectedCRD.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
	Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout, standardInterval).Should(BeFalse())
}

func withValidImageSpecAndWithoutCR(installName string) func(manifest *v1beta2.Manifest) error {
	return func(manifest *v1beta2.Manifest) error {
		createManifestFn := withValidInstallImageSpec(installName, false, false)
		err := createManifestFn(manifest)
		if err != nil {
			return err
		}
		deploy := &appsv1.Deployment{}
		Eventually(setDeploymentStatus("nginx-deployment", deploy), standardTimeout, standardInterval).Should(Succeed())
		return nil
	}
}

func withValidImageSpecAndWithCR(installName string) func(manifest *v1beta2.Manifest) error {

	return func(manifest *v1beta2.Manifest) error {
		createManifestFn := withValidInstallImageSpec(installName, true, false)
		err := createManifestFn(manifest)
		if err != nil {
			return err
		}
		expectedDeployment := asResource("nginx-deployment", "default", "apps", "v1", "Deployment")
		Eventually(verifyObjectExists(expectedDeployment.ToUnstructured()), standardTimeout/10, standardInterval).Should(BeTrue())
		deploy := &appsv1.Deployment{}
		Eventually(setDeploymentStatus("nginx-deployment", deploy), standardTimeout, standardInterval).Should(Succeed())
		sampleCR := emptySampleCR(manifest.GetName())
		Eventually(setCRStatus(sampleCR, declarative.StateReady), standardTimeout, standardInterval).Should(Succeed())
		return nil
	}
}

func withInvalidImageSpec() func(manifest *v1beta2.Manifest) error {
	return withInvalidInstallImageSpec(false)
}

var _ = Describe(
	"Given manifest with private registry", func() {
		mainOciTempDir := "private-oci"
		installName := filepath.Join(mainOciTempDir, "crs")

		It(
			"setup remote oci Registry",
			func() {
				PushToRemoteOCIRegistry(installName)
			},
		)
		BeforeEach(
			func() {
				Expect(os.RemoveAll(filepath.Join(os.TempDir(), mainOciTempDir))).To(Succeed())
			},
		)

		It("Manifest should be in Error state with no auth secret found error message", func() {
			manifestWithInstall := NewTestManifest("private-oci-registry")
			Eventually(withValidInstallImageSpec(installName, false, true), standardTimeout, standardInterval).
				WithArguments(manifestWithInstall).Should(Succeed())
			Eventually(func() string {
				status, err := getManifestStatus(manifestWithInstall.GetName())
				if err != nil {
					return err.Error()
				}

				if status.State != declarative.StateError {
					return "manifest not in error state"
				}
				if strings.Contains(status.LastOperation.Operation, ocmextensions.ErrNoAuthSecretFound.Error()) {
					return ocmextensions.ErrNoAuthSecretFound.Error()
				}
				return status.LastOperation.Operation
			}, standardTimeout, standardInterval).
				Should(Equal(ocmextensions.ErrNoAuthSecretFound.Error()))
		})
	},
)
