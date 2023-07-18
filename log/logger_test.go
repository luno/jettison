package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

type testLogger struct {
	logs []log.Log
}

func (tl *testLogger) Log(l log.Log) string {
	tl.logs = append(tl.logs, l)

	return ""
}

func TestAddLoggers(t *testing.T) {
	tl := new(testLogger)
	log.SetLogger(tl)

	log.Info(nil, "message", j.KV("some", "param"))
	log.Error(nil, errors.New("errMsg"))

	assert.Equal(t, "message,info,some,param,", toStr(tl.logs[0]))
	assert.Equal(t, "errMsg,error,", toStr(tl.logs[1]))
}

func toStr(l log.Log) string {
	str := l.Message + ","
	str += string(l.Level) + ","
	if len(l.Parameters) == 0 {
		return str
	}
	for _, kv := range l.Parameters {
		str += kv.Key + "," + kv.Value + ","
	}
	return str
}
