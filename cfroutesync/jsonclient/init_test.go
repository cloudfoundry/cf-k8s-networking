package jsonclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestJsonclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jsonclient Suite")
}
