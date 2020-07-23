package log

import (
"fmt"
	"io"
	"log"
"strings"

"github.com/luno/jettison/errors"
"github.com/luno/jettison/models"
"gopkg.in/yaml.v2"
)

const errorRedBold = "\033[1;31m%s\033[0m"

// newCmdLogger returns a human friendly command line logger.
func newCmdLogger(w io.Writer) LoggerFunc {
	return func(l Log) string {
		return cmdLog(log.New(w, "", 0), l)
	}
}

func cmdLog(logger *log.Logger, l Log) string {
	text := fmt.Sprintf("%s %s %s: %s",
		strings.ToUpper(string(l.Level))[:1],
		l.Timestamp.Format("15:04:05.000"),
		conciseSource(l.Source),
		makeMsg(l),
	)

	if len(l.Hops) > 0 {
		hops, err := yamlHops(l.Hops)
		if err != nil {
			logger.Printf("error printing hops: %v", err)
		} else {
			text += "\n" + hops
		}
	}

	if l.Level == LevelError {
		var lines []string
		for _, line := range strings.Split(text, "\n") {
			lines = append(lines, fmt.Sprintf(errorRedBold, line))
		}
		text = strings.Join(lines, "\n")
	}

	logger.Print(text)

	return text
}

// makeMsg returns the log message with parameters if present.
func makeMsg(l Log) string {
	var res strings.Builder
	res.WriteString(l.Message)
	if len(l.Parameters) == 0 {
		return res.String()
	}
	var pl []string
	for _, p := range l.Parameters {
		pl = append(pl, fmt.Sprintf("%s=%s", p.Key, p.Value))
	}
	res.WriteString("[")
	res.WriteString(strings.Join(pl, ","))
	res.WriteString("]")
	return res.String()
}

// conciseSource returns the source with the leading package
// import path abbreviated to first letters only.
//   github.com/luno/jettison/log/log.go:136 > g/l/j/l/log.go:136
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

// yamlHops returns the stack traces of the hops as indented yaml.
func yamlHops(hops []models.Hop) (string, error) {
	var v interface{}
	if len(hops) == 0 {
		return "", errors.New("missing hops")
	} else if len(hops) == 1 {
		// If single binary (no network hops), just print the stack.
		v = hops[0].StackTrace
	} else {
		// Else if network hops, print binaries with stacks.
		var stacks yaml.MapSlice
		for _, hop := range hops {
			stacks = append(stacks, yaml.MapItem{
				Key:   hop.Binary,
				Value: hop.StackTrace,
			})
		}
		v = stacks
	}

	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}

	res := "  " + strings.TrimSpace(string(b))
	res = strings.Replace(res, "\n", "\n  ", -1)
	return res, nil
}
