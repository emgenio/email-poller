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

func (obj *IWMessage) Display() {
  fmt.Println("=====================================")
  fmt.Println("UID:\n", obj.Uid)
  fmt.Println("SUBJECT:\n", string(obj.Header))
  fmt.Println("BODY:\n", string(obj.Body))
  fmt.Println("=====================================")
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

func (obj *ImapWrapper) FetchAllMessages() ([]IWMessage, bool) {
  set, _ := imap.NewSeqSet("1:*")
  obj.Command, _ = obj.Client.UIDFetch(set, "RFC822.HEADER", "RFC822.TEXT")

  set, _ = imap.NewSeqSet("")
  messages := []IWMessage{}
  for obj.Command.InProgress() {
    // Wait for the next response (no timeout)
    obj.Client.Recv(-1)
    for _, rsp := range obj.Command.Data {
      attrs := rsp.MessageInfo().Attrs
      iwmsg := IWMessage{
        Uid:    imap.AsNumber(attrs["UID"]),
        Header: imap.AsBytes(attrs["RFC822.HEADER"]),
        Body:   imap.AsBytes(attrs["RFC822.TEXT"]),
      }
      messages = append(messages, iwmsg)
      set.AddNum(iwmsg.Uid)
    }
    obj.Command.Data = nil
    obj.Client.Data = nil
  }

  obj.Client.UIDStore(set, "+FLAGS", imap.NewFlagSet(`\Deleted`))
  obj.Client.Expunge(nil)
  return messages, false
}

func (obj *ImapWrapper) ListenIncomingMessages(timeout time.Duration) ([]uint32, bool) {
  obj.Command, obj.error = obj.Client.Idle()
  if obj.ReportError("Error When Idling Client:", true) {
    return nil, false
  }

  obj.error = obj.Client.Recv(timeout)
  if obj.ReportError("Error Recv:", true) {
    return nil, false
  }

  obj.Command, obj.error = imap.Wait(obj.Client.IdleTerm())
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
    obj.Command, obj.error = imap.Wait(obj.Client.Fetch(set, "RFC822.HEADER", "RFC822.TEXT"))
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
    for obj.Command.InProgress() {

    }
    obj.Client.Store(set, "+FLAGS", imap.NewFlagSet(`\Deleted`))
    obj.Client.Expunge(nil)
    return messages, true
  }
  return nil, true
}

func (obj *ImapWrapper) ExpungeMessageFromIds(ids []uint32) {
  set, _ := imap.NewSeqSet("")
  set.AddNum(ids...)
  obj.Client.UIDStore(set, "+FLAGS", imap.NewFlagSet(`\Deleted`))
  obj.Client.Expunge(nil)
}

func (obj *ImapWrapper) RetrieveIWMessages(arguments ...string) ([]IWMessage, bool) {
  var ids_tab []uint32

  ids_tab, _ = obj.SearchForIds(arguments...)

  messages, err := obj.BuildIWMessagesFromIds(ids_tab)
  if err != false {
    return messages, err
  }

  return messages, err
}

// arguments: RECENT / UNSEEN / UIDNEXT ...
func (obj *ImapWrapper) SearchForIds(arguments ...string) ([]uint32, bool) {
  args_tab := toImapFields(arguments...)
  fmt.Println("TAB:", args_tab[0])
  obj.Command, obj.error = imap.Wait(obj.Client.Search(args_tab...))
  if obj.ReportError("Error while searching for ids:", true) {
    return nil, false
  }

  return obj.Command.Data[0].SearchResults(), true
}

func toImapFields(arguments ...string) ([]imap.Field) {
  args_tab := []imap.Field{}
  for _, argument := range arguments {
    args_tab = append(args_tab, argument)
  }
  return args_tab
}

func init() {}

func Create(config string) *ImapWrapper {
  newClient := ImapWrapper{}
  newClient.loadConfigFromYaml(config)
  return &newClient
}