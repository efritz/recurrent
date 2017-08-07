package recurrent

import (
	"testing"
	"time"

	"github.com/aphistic/sweet"
	"github.com/aphistic/sweet-junit"
	. "github.com/onsi/gomega"
)

const (
	TIMEOUT  = time.Millisecond * 1
	INTERVAL = time.Millisecond / 100
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.RegisterPlugin(junit.NewPlugin())

		s.AddSuite(&ChannelsSuite{})
		s.AddSuite(&SchedulerSuite{})
	})
}

//
// Overwrite gomega's default timeout and interval

func eventually(actual interface{}) GomegaAsyncAssertion {
	return Eventually(actual, TIMEOUT, INTERVAL)
}
