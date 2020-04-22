package cfg_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCfg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfg Suite")
}
