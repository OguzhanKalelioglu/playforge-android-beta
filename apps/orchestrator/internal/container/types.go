package container

import (
	"encoding/json"
	"strings"
)

type ContainerInfo struct {
	Name     string `json:"Name"`
	Service  string `json:"Service"`
	State    string `json:"State"`
	Status   string `json:"Status"`
	Health   string `json:"Health"`
	ExitCode int    `json:"ExitCode"`
	ID       string `json:"ID"`
}

func parseContainerList(output string) ([]ContainerInfo, error) {
	output = strings.TrimSpace(output)
	if output == "" || output == "null" {
		return nil, nil
	}

	dec := json.NewDecoder(strings.NewReader(output))
	var out []ContainerInfo
	for {
		var c ContainerInfo
		if err := dec.Decode(&c); err != nil {
			if err.Error() == "EOF" {
				break
			}
			var arr []ContainerInfo
			if err := json.Unmarshal([]byte(output), &arr); err == nil {
				return arr, nil
			}
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
