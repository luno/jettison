package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/luno/jettison/models"
)

// NewCmdLogger returns a stdout human friendly command line logger.
func NewCmdLogger(w io.Writer, stripTime bool) *CmdLogger {
	return &CmdLogger{
		logger:    log.New(w, "", 0),
		stripTime: stripTime,
	}
}

type CmdLogger struct {
	logger    *log.Logger
	stripTime bool
}

func (c *CmdLogger) Log(_ context.Context, l Entry) string {
	timestamp := l.Timestamp.Format("15:04:05.000")
	if c.stripTime {
		timestamp = "00:00:00.000"
	}

	errs := l.ErrorObjects
	if l.ErrorObject != nil {
		errs = append(errs, *l.ErrorObject)
	}

	var sb strings.Builder
	if len(errs) == 0 {
		_, _ = fmt.Fprintf(&sb, "%s %s %s: %s",
			strings.ToUpper(string(l.Level))[:1],
			timestamp,
			conciseSource(l.Source),
			makeMsg(l),
		)
	} else {
		_, _ = fmt.Fprintf(&sb, "%s %s %s: error(s) %s\n",
			strings.ToUpper(string(l.Level))[:1],
			timestamp,
			conciseSource(l.Source),
			parameterString(l.Parameters),
		)
		for _, err := range errs {
			writeError(&sb, err)
		}
	}
	c.logger.Print(sb.String())
	return sb.String()
}

// makeMsg returns the log message with parameters if present.
func makeMsg(l Entry) string {
	return fmt.Sprint(l.Message, parameterString(l.Parameters))
}

func parameterString(params []models.KeyValue) string {
	if len(params) == 0 {
		return ""
	}
	pl := make([]string, 0, len(params))
	for _, p := range params {
		pl = append(pl, fmt.Sprintf("%s=%s", p.Key, p.Value))
	}
	return fmt.Sprintf("[%s]", strings.Join(pl, ","))
}

// conciseSource returns the source with the leading package
// import path abbreviated to first letters only.
//
//	github.com/luno/jettison/log/log.go:136 > g/l/j/l/log.go:136
func conciseSource(source string) string {
	split := strings.Split(source, "/")
	var res []string
	for i, s := range split {
		if i < len(split)-2 {
			res = append(res, string([]rune(s)[0]))
		} else {
			res = append(res, s)
		}
	}

	return strings.Join(res, "/")
}

func writeError(w io.Writer, err ErrorObject) {
	ps := parameterString(err.Parameters)
	_, _ = fmt.Fprintf(w, "  %s%s", err.Message, ps)
	if len(err.StackTrace.Content()) == 0 {
		_, _ = fmt.Fprint(w, "(error without stack trace)")
	}
	_, _ = fmt.Fprintln(w)
	for _, line := range err.StackTrace.Content() {
		_, _ = fmt.Fprintf(w, "  - %s\n", line)
	}
}
