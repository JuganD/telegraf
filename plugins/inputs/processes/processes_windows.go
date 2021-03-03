// +build windows

package processes

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Processes struct {
	Log telegraf.Logger
}

func execWmicProcesses() ([]byte, error) {
	bin, err := exec.LookPath("wmic")
	if err != nil {
		return nil, err
	}

	args := "process get HandleCount, ThreadCount /format:csv"
	out, err := exec.Command(bin, strings.Split(args, " ")...).Output()
	if err != nil {
		return nil, err
	}

	return out, err
}

func (e *Processes) Gather(acc telegraf.Accumulator) error {

	out, err := execWmicProcesses()
	if err != nil {
		e.Log.Infof("Can't fetch process information from WMIC: %s", err.Error())
		return nil
	}

	fields := map[string]interface{}{
		"processesCount": int64(0),
		"handlesCount":   int64(0),
		"threadsCount":   int64(0),
	}

	// Split by new line (Windows style)
	for i, status := range strings.Split(string(out), "\r\n") {
		if i == 0 || i == 1 {
			// This is a header, skip it
			continue
		}

		// Trim any leftover new lines that previous split did not cover and then split by comma (,)
		splitedStatus := strings.Split(strings.TrimSuffix(status, "\r"), ",")

		fields["processesCount"] = fields["processesCount"].(int64) + int64(1)

		if len(splitedStatus) > 2 {
			currentHandlesCount, err := strconv.ParseInt(splitedStatus[1], 10, 64)
			if err != nil {
				e.Log.Infof("Can't parse current handles count: %s", err.Error())
				e.Log.Infof("Was trying to parse: %s", splitedStatus[1])
			} else {
				fields["handlesCount"] = fields["handlesCount"].(int64) + currentHandlesCount
			}

			currentThreadsCount, err := strconv.ParseInt(splitedStatus[2], 10, 64)
			if err != nil {
				e.Log.Infof("Can't parse current threads count: %s", err.Error())
				e.Log.Infof("Was trying to parse: %s", splitedStatus[2])
			} else {
				fields["threadsCount"] = fields["threadsCount"].(int64) + currentThreadsCount
			}

		}

	}

	acc.AddFields("processes", fields, nil)
	return nil
}

func init() {
	inputs.Add("processes", func() telegraf.Input {
		return &Processes{}
	})
}
