/*
Copyright 2021 Banzai Cloud.

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

package util

import (
	"time"

	"github.com/pkg/errors"
)

func WaitForCondition(timeout time.Duration, interval time.Duration, f func() (bool, error)) error {
	start := time.Now()
	timer := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timer:
			return errors.Errorf("Timeout after %s", timeout)
		case <-ticker.C:
			result, err := f()
			if err != nil {
				elapsed := time.Now().Sub(start)
				return errors.Wrapf(err, "got error while checking for condition (elapsed: %s)", elapsed)
			}
			if result {
				return nil
			}
		}
	}
}
