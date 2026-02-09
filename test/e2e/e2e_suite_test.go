package e2e

import (
	"fmt"
	"testing"

	"github.com/cnap-tech/cloudflare-kubernetes-gateway/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
	log "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/gateway-api/conformance"
	conformancev1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	fmt.Fprintf(GinkgoWriter, "Starting cloudflare-kubernetes-gateway suite\n") //nolint:errcheck
	RunSpecs(t, "e2e suite")

	fmt.Fprintf(GinkgoWriter, "Starting gateway-api conformance suite\n") //nolint:errcheck
	version, err := utils.GetProjectVersion()
	if err != nil {
		t.Fatalf("failed to get project version: %v", err)
	}

	log.SetLogger(GinkgoLogr)
	opts := conformance.DefaultOptions(t)
	opts.CleanupBaseResources = false
	opts.ConformanceProfiles = sets.New(
		suite.GatewayHTTPConformanceProfileName,
	)
	opts.EnableAllSupportedFeatures = true
	opts.Debug = true
	opts.Implementation = conformancev1.Implementation{
		Contact:      []string{"https://github.com/cnap-tech/cloudflare-kubernetes-gateway/issues/new/choose"},
		Organization: "cnap-tech",
		Project:      "cloudflare-kubernetes-gateway",
		URL:          "https://github.com/cnap-tech/cloudflare-kubernetes-gateway",
		Version:      version,
	}
	opts.ReportOutputPath = "standard-" + version + "-default-report.yaml"
	conformance.RunConformanceWithOptions(t, opts)
}
