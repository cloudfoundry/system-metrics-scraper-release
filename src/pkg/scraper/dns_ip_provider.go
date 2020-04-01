package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type FileContents struct {
	RecordKeys  []string        `json:"record_keys"`
	RecordInfos [][]interface{} `json:"record_infos"`
}

func NewDNSScrapeTargetProvider(sourceID, dnsFile string, port int) TargetProvider {
	return func() []Target {
		file, err := os.Open(dnsFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var d FileContents
		err = json.NewDecoder(file).Decode(&d)
		if err != nil {
			panic(err)
		}

		var targets []Target

		keyMap := make(map[int]string)
		for idx, keyName := range d.RecordKeys {
			if keyName == "instance_group" || keyName == "deployment" || keyName == "ip" || keyName == "id" {
				keyMap[idx] = keyName
			}
		}

		for _, record := range d.RecordInfos {
			defaultTags := make(map[string]string)
			var ip string
			for idx, keyName := range keyMap {
				recordItem := fmt.Sprintf("%v", record[idx])

				if keyName == "ip" {
					ip = recordItem
				}
				defaultTags[keyName] = recordItem
			}

			targets = append(targets, Target{
				ID:          sourceID,
				MetricURL:   fmt.Sprintf("https://%s:%d/metrics", ip, port),
				DefaultTags: defaultTags,
			})
		}

		return targets
	}
}
