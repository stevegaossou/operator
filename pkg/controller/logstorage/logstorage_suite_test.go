package logstorage

import (
"testing"

. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"

"github.com/onsi/ginkgo/reporters"
)

func TestInstallation(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../../report/logstorage_suite.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "pkg/controller/logstorage Suite", []Reporter{junitReporter})
}

