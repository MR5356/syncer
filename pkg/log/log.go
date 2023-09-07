package log

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"runtime"
)

func init() {
	logrus.SetReportCaller(true)

	logrus.SetFormatter(&nested.Formatter{
		HideKeys:        false,
		FieldsOrder:     []string{"level"},
		TimestampFormat: "2006-01-02 15:04:05",
		TrimMessages:    true,
		CallerFirst:     false,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			return ""
			//return fmt.Sprintf(" %s:%d", frame.Function, frame.Line)
		},
	})
}
