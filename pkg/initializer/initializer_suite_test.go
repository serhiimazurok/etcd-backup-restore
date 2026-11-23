// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package initializer

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInitializer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Initializer Suite")
}
