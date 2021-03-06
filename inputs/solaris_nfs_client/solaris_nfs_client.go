package solaris_nfs_client

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/siebenmann/go-kstat"
	sh "github.com/snltd/sunos_helpers"
	"log"
	"strings"
)

var sampleConfig = `
	## The NFS versions you wish to monitor
	#NfsVersions = [3, 4]
  ## The kstat fields you wish to emit. 'kstat -p -m nfs -i 0 | grep rfs'
	## will list the possibilities
	#Fields = ["read", "write", "remove", "create", "getattr", "setattr"]
`

func (s *SolarisNfsClient) Description() string {
	return "Reports Solaris NFS client statistics"
}

func (s *SolarisNfsClient) SampleConfig() string {
	return sampleConfig
}

type SolarisNfsClient struct {
	Fields      []string
	NfsVersions []string
}

func (s *SolarisNfsClient) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	if err != nil {
		log.Fatal("cannot get kstat token")
	}

	ks := sh.KstatModule(token, "nfs")

	for _, stat := range ks {
		if !strings.HasPrefix(stat.Name, "rfsreqcnt_v") {
			continue
		}

		nfs_ver := stat.Name[len(stat.Name)-1:]

		if !sh.WeWant(nfs_ver, s.NfsVersions) {
			continue
		}

		tags["nfs_version"] = nfs_ver
		stats, _ := stat.AllNamed()

		for _, stat := range stats {
			if !sh.WeWant(stat.Name, s.Fields) {
				continue
			}

			fname := stat.Name
			stat_type := fmt.Sprintf("%s", stat.Type)

			if stat_type == "uint32" || stat_type == "uint64" {
				fields[fname] = stat.UintVal
			} else if stat_type == "int32" || stat_type == "int64" {
				fields[fname] = stat.IntVal
			}

		}
	}

	acc.AddFields("solaris_nfs_client", fields, tags)
	token.Close()
	return nil
}

func init() {
	inputs.Add("solaris_nfs_client", func() telegraf.Input {
		return &SolarisNfsClient{}
	})
}
