package imapWrapper

import (
  "log"
  "path/filepath"
  "io/ioutil"
  "gopkg.in/yaml.v2"
)

type ImapWrapperConfig struct {
  Host string
  User string
  Password string
}

func (obj *ImapWrapper) loadConfigFromYaml(configPath string) {
  config := ImapWrapperConfig{}
  filename, _ := filepath.Abs(configPath)
  yamlFile, err := ioutil.ReadFile(filename)

  if err != nil {
    log.Fatalf("error: %v", err)
  }
  err = yaml.Unmarshal(yamlFile, &config)
  if err != nil {
    log.Fatalf("error: %v", err)
  }
 obj.cfg = &config
}
