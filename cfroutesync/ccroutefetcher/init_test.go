package ccroutefetcher_test

import (
	"testing"

	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCcroutefetcher(t *testing.T) {
	RegisterFailHandler(Fail)
	log.SetOutput(GinkgoWriter)
	log.SetFormatter(&log.JSONFormatter{})
	RunSpecs(t, "Ccroutefetcher Suite")
}
