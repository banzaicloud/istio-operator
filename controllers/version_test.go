package controllers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/banzaicloud/istio-operator/v2/controllers"
)

var _ = Describe("IsIstioVersionSupported()", func() {
	It("should deny unsupported versions", func() {
		Expect(controllers.IsIstioVersionSupported("2.1")).To(BeFalse())
		Expect(controllers.IsIstioVersionSupported("1.13")).To(BeFalse())
		Expect(controllers.IsIstioVersionSupported("1.13.1")).To(BeFalse())
	})
	It("should accept minor versions", func() {
		Expect(controllers.IsIstioVersionSupported("1.11")).To(BeTrue())
	})
	It("should accept micro versions", func() {
		Expect(controllers.IsIstioVersionSupported("1.11.1")).To(BeTrue())
	})
})
