package imapWrapper

import (
  "github.com/mxk/go-imap/imap"
)

type ImapWrapper struct {
  c   *imap.Client
  cmd *imap.Command
  rsp *imap.Response
  cfg *ImapWrapperConfig
}

func init() {}

func Create(config string) *ImapWrapper {
  newClient := ImapWrapper{}
  newClient.loadConfigFromYaml(config)
  return &newClient
}