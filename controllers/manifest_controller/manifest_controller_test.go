package manifest_controller_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	declarative "github.com/kyma-project/lifecycle-manager/internal/declarative/v2"
	"github.com/kyma-project/lifecycle-manager/pkg/ocmextensions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var ErrManifestStateMisMatch = errors.New("ManifestState mismatch")

var _ = Describe(
	"Rendering manifest install layer", func() {
		mainOciTempDir := "main-dir"
		installName := filepath.Join(mainOciTempDir, "installs")
		var imageDigest v1.Hash

		It(
			"setup OCI", func() {
				imageDigest = PushToRemoteOCIRegistry(installName, "../../pkg/test_samples/oci/rendered.yaml")
				fmt.Println(imageDigest)
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
			) {
				manifest := NewTestManifest("oci")
				Eventually(givenCondition, standardTimeout, standardInterval).
					WithArguments(manifest).Should(Succeed())
				Eventually(expectManifestState, standardTimeout, standardInterval).
					WithArguments(manifest.GetName()).Should(Succeed())
				Eventually(deleteManifestAndVerify(manifest), standardTimeout, standardInterval).Should(Succeed())
			},
			Entry(
				"When Manifest CR contains a valid install OCI image specification, "+
					"expect state in ready",
				withValidInstallImageSpec(imageDigest, installName, false, false),
				expectManifestStateIn(declarative.StateReady),
			),
			Entry(
				"When Manifest CR contains a valid install OCI image specification and enabled deploy resource, "+
					"expect state in ready",
				withValidInstallImageSpec(imageDigest, installName, true, false),
				expectManifestStateIn(declarative.StateReady),
			),
			Entry(
				"When Manifest CR contains an invalid install OCI image specification, "+
					"expect state in error",
				withInvalidInstallImageSpec(imageDigest, false),
				expectManifestStateIn(declarative.StateError),
			),
		)
	},
)

var _ = Describe(
	"Given manifest with private registry", func() {
		mainOciTempDir := "private-oci"
		installName := filepath.Join(mainOciTempDir, "crs")

		var imageDigest v1.Hash
		It(
			"setup remote oci Registry",
			func() {
				imageDigest = PushToRemoteOCIRegistry(installName, "../../pkg/test_samples/oci/rendered.yaml")
			},
		)
		BeforeEach(
			func() {
				Expect(os.RemoveAll(filepath.Join(os.TempDir(), mainOciTempDir))).To(Succeed())
			},
		)

		It("Manifest should be in Error state with no auth secret found error message", func() {
			manifestWithInstall := NewTestManifest("private-oci-registry")
			Eventually(withValidInstallImageSpec(imageDigest, installName, false, true), standardTimeout, standardInterval).
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
