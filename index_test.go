package structomancer_test

import (
	"github.com/listenonrepeat/backend/common/log/loggomega"
    // "github.com/brynbellomy/ginkgo-reporter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStructomancer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithCustomReporters(t, "Structomancer Suite", []Reporter{
		&loggomega.TerseReporter{Logger: &loggomega.DefaultLogger{}},
	})
}
