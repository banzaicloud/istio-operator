/*
Copyright 2022 Cisco Systems, Inc. and/or its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/banzaicloud/istio-operator/v2/controllers"
)

var _ = Describe("IsIstioVersionSupported()", func() {
	It("should deny unsupported versions", func() {
		for _, version := range []string{"2.11", "2.11.1", "2.11.1-dev", "1.11", "1.11.1", "1.11.1-dev"} {
			Expect(controllers.IsIstioVersionSupported(version)).To(BeFalse(), "invalid: "+version)
		}
	})
	It("should accept all 1.15 versions", func() {
		Expect(controllers.IsIstioVersionSupported("1.15")).To(BeTrue())
	})
	It("should accept all 1.15 versions with qualifier", func() {
		Expect(controllers.IsIstioVersionSupported("1.15-dev")).To(BeTrue())
	})
	It("should accept micro versions", func() {
		Expect(controllers.IsIstioVersionSupported("1.15.2")).To(BeTrue())
	})
	It("should accept micro versions with qualifier", func() {
		Expect(controllers.IsIstioVersionSupported("1.15.2-dev")).To(BeTrue())
	})
})
