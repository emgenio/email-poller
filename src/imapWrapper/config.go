package imapWrapper

import (
  "log"
  "path/filepath"
  "io/ioutil"
  "gopkg.in/yaml.v2"
)

type ImapWrapperConfig struct {
  Hostname string
  User string
  Password string
  Mbox string
}

func (obj *ImapWrapper) loadConfigFromYaml(configPath string) {
  config := ImapWrapperConfig{}
  filename, err := filepath.Abs(configPath)
  yamlFile, err := ioutil.ReadFile(filename)

  if err != nil {
    log.Fatalf("error: %v", err)
  }
  err = yaml.Unmarshal(yamlFile, &config)
  if err != nil {
    log.Fatalf("error: %v", err)
  }
 obj.Config = &config
}
