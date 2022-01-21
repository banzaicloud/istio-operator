package controllers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/banzaicloud/istio-operator/v2/controllers"
)

var _ = Describe("IsIstioVersionSupported()", func() {
	It("should deny unsupported versions", func() {
		for _, version := range []string{"2.11", "2.11.1", "2.11.1-dev", "1.12", "1.12.1", "1.12.1-dev"} {
			Expect(controllers.IsIstioVersionSupported(version)).To(BeFalse(), "invalid: "+version)
		}
	})
	It("should accept all 1.11 versions", func() {
		Expect(controllers.IsIstioVersionSupported("1.11")).To(BeTrue())
	})
	It("should accept all 1.11 versions with qualifier", func() {
		Expect(controllers.IsIstioVersionSupported("1.11-dev")).To(BeTrue())
	})
	It("should accept micro versions", func() {
		Expect(controllers.IsIstioVersionSupported("1.11.1")).To(BeTrue())
	})
	It("should accept micro versions with qualifier", func() {
		Expect(controllers.IsIstioVersionSupported("1.11.1-dev")).To(BeTrue())
	})
})
