// Copyright 2018 fydrah
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	log "github.com/sirupsen/logrus"
)

// appCheck struct
// used by check function
type Check struct {
	FailedCondition bool
	Message         string
	DefaultAction   func()
}

// Check checks if the check failed or not
// Return true if check pass
// Return false if check fails
func (c *Check) Check() bool {
	if c.FailedCondition {
		if c.DefaultAction != nil {
			c.DefaultAction()
			log.Info(c.Message)
		} else {
			log.Error(c.Message)
		}
		return false
	}
	return true
}
