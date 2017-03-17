package recurrent

import (
	"testing"
	"time"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

const (
	TIMEOUT  = time.Millisecond * 1
	INTERVAL = time.Millisecond / 100
)

func Test(t *testing.T) {
	sweet.T(func(s *sweet.S) {
		RegisterFailHandler(sweet.GomegaFail)

		s.RunSuite(t, &ChannelsSuite{})
		s.RunSuite(t, &SchedulerSuite{})
	})
}

func eventually(actual interface{}) GomegaAsyncAssertion {
	return Eventually(actual, TIMEOUT, INTERVAL)
}
