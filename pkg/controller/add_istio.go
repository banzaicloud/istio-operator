package controller

import (
	"github.com/banzaicloud/istio-operator/pkg/controller/istio"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, istio.Add)
}
