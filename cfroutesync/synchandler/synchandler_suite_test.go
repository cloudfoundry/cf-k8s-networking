package synchandler_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRoutelister(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SyncHandler Suite")
}
