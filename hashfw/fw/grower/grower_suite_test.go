package grower

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGrowerSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Grower Suite")
}
