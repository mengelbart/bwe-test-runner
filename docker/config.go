package docker

import (
	"encoding/json"
	"io/ioutil"
)

func copyFile(src, dst string) error {
	bytesRead, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dst, bytesRead, 0644); err != nil {
		return err
	}
	return nil
}

func mustParseConfig(config string, data interface{}) {
	err := json.Unmarshal([]byte(config), data)
	if err != nil {
		panic(err)
	}
}
