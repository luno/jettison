package trace

import "fmt"

type Merge struct {
	traces   [][]string
	binaries []string
}

func (m *Merge) Add(trace []string, binary string) {
	m.traces = append(m.traces, trace)
	m.binaries = append(m.binaries, binary)
}

func (m *Merge) FullTrace() []string {
	var ret []string
	for i := len(m.traces) - 1; i >= 0; i-- {
		ret = append(ret, m.traces[i]...)
		if i > 0 {
			ret = append(ret,
				fmt.Sprintf("%s -> %s", m.binaries[i-1], m.binaries[i]),
			)
		}
	}
	return ret
}
