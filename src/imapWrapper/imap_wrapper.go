package imapWrapper

import (
  "fmt"
  "time"
  "github.com/mxk/go-imap/imap"
)

type ImapWrapper struct {
  Client *imap.Client
  Command *imap.Command
  Response *imap.Response
  Config *ImapWrapperConfig

  error error
}

func (obj *ImapWrapper) ReportError(msg string, display bool) bool {
   if obj.error != nil {
    obj.error = fmt.Errorf(msg, obj.error)
    if display {
      fmt.Println(obj.error)
    }
    return true
   }
  return false
}

func (obj *ImapWrapper) Connect() bool {
  obj.Client, obj.error = imap.Dial(obj.Config.Hostname)

  if obj.ReportError("Error when trying to connect to IMAP server:", true) {
    return false
  }


  fmt.Println("Connsected to IMAP server:", obj.Client.Data[0].Info)
  return true
}

func (obj *ImapWrapper) Disconnect() {
  obj.Command, obj.error = imap.Wait(obj.Client.Logout(30 * time.Second))
  obj.Client.Close(true)
}

func init() {}

func Create(config string) *ImapWrapper {
  newClient := ImapWrapper{}
  newClient.loadConfigFromYaml(config)
  return &newClient
}