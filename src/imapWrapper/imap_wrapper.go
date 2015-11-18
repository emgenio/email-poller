package imapWrapper

import (
  "fmt"
  "time"
  "github.com/mxk/go-imap/imap"
)

type IWMessage struct {
  Uid uint32
  Header []byte
  Body []byte
}

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

  fmt.Println("Connected to IMAP server:", obj.Client.Data[0].Info)
  return true
}

func (obj *ImapWrapper) Disconnect() {
  obj.Command, obj.error = imap.Wait(obj.Client.Logout(30 * time.Second))
  obj.Client.Close(true)
}

func (obj *ImapWrapper) Login() bool {
  obj.Command, obj.error = imap.Wait(obj.Client.Login(obj.Config.User, obj.Config.Password))
  if obj.ReportError("Error Auth to IMAP server:", true) {
    return false
  }
  return true
}

func (obj *ImapWrapper) SelectBox(mbox string, readonly bool) bool {
  obj.Command, obj.error = imap.Wait(obj.Client.Select(mbox, readonly))
  if obj.ReportError("Error Selecting mBox:", true) {
    return false
  }
  return true
}

func (obj *ImapWrapper) FetchAllMessages() bool {
  set, _ := imap.NewSeqSet("1:*")
  obj.Command, obj.error = imap.Wait(obj.Client.UIDFetch(set, "RFC822.HEADER", "RFC822.TEXT"))

  if obj.ReportError("Error while fetching messages:", true) {
    return false
  }
  return true
}

func (obj *ImapWrapper) ListenIncomingMessages(timeout time.Duration) ([]uint32, bool) {
  _, obj.error = obj.Client.Idle()
  if obj.ReportError("Error When Idling Client:", true) {
    return nil, false
  }

  obj.error = obj.Client.Recv(timeout)
  if obj.ReportError("Error Recv:", true) {
    return nil, false
  }

  _, obj.error = imap.Wait(obj.Client.IdleTerm())
  if obj.ReportError("Error while retrieving response from idle:", true) {
    return nil, false
  }

  ids := []uint32{}
  for _, response := range obj.Client.Data {
    switch response.Label {
    case "EXISTS":
      ids = append(ids, imap.AsNumber(response.Fields[0]))
    }
  }
  return ids, true
}

func (obj *ImapWrapper) BuildIWMessagesFromIds(ids []uint32) ([]IWMessage, bool) {
  if len(ids) > 0 {
    set, _ := imap.NewSeqSet("")
    set.AddNum(ids...)
    obj.Command, obj.error = imap.Wait(obj.Client.Fetch(set, "RFC822"))
    if obj.error != nil {
      return nil, false
    }

    messages := []IWMessage{}
    for _, response := range obj.Command.Data {
      attrs := response.MessageInfo().Attrs
      iwmsg := IWMessage{
        Uid:    imap.AsNumber(attrs["UID"]),
        Header: imap.AsBytes(attrs["RFC822.HEADER"]),
        Body:   imap.AsBytes(attrs["RFC822.TEXT"]),
      }
      messages = append(messages, iwmsg)
    }
    return messages, true
  }
  return nil, true
}

func init() {}

func Create(config string) *ImapWrapper {
  newClient := ImapWrapper{}
  newClient.loadConfigFromYaml(config)
  return &newClient
}