package imapClient

import (
  // Wrapping this package
  "github.com/mxk/go-imap/imap"

  "fmt"
  "bytes"
  "net/mail"
  "time"
  "log"
  "os"
  "encoding/json"
)

type GoImapClient struct {
  Client  *imap.Client
  Cmd     *imap.Command
  Rsp     *imap.Response

  // addr
  Hostname  string
  Port      uint32

  // creds
  User      string
  Password  string
}

// Methods
func (obj *GoImapClient)      Connect() (error) {
  client, err := imap.Dial(obj.GetFormatedAddr())
  obj.Client = client
  return err
}

func (obj *GoImapClient)      Login() (*imap.Command, error) {
 if obj.Client.State() == imap.Login {
    cmd, err := imap.Wait(obj.Client.Login(obj.User, obj.Password))
    return cmd, err
  }
  return nil, nil
}

func (obj *GoImapClient)      Logout(timeout time.Duration) (*imap.Command, error) {
  cmd, err := imap.Wait(obj.Client.Logout(timeout))
  return cmd, err
}

func (obj *GoImapClient)      SelectMailBox(mbox string, readonly bool) (*imap.Command, error) {
  cmd, err := imap.Wait(obj.Client.Select(mbox, readonly))
  return cmd, err
}

func (obj *GoImapClient)      GetFormatedAddr() string {
  return fmt.Sprintf("%s:%d", obj.Hostname, obj.Port)
}

func (obj *GoImapClient)      SupportIdleCap() bool {
  return obj.Client.Caps["IDLE"]
}

func (obj GoImapClient)      WaitForNotifications() (cmd *imap.Command) {

  // Setting Client in Idle state
  cmd, err := obj.Client.Idle()
  if err != nil {
    fmt.Println("WaitForNotifications:Recv:Idle", err)
    return cmd
  }

  // Waiting for notifications... 30 sec timeout not to disconnect the Client, RFC says
  // Client gets disconnected passed 29 minutes if no notifs
  err = obj.Client.Recv(3 * time.Second)
  if err != imap.ErrTimeout && err != nil {
    fmt.Println("WaitForNotifications:Recv:", err)
  }

  // Notifications received or timed out
  // Terminating Client Idle stance. Make the call synchronous
  cmd, err = imap.Wait(obj.Client.IdleTerm())
  if err != nil {
    fmt.Println("WaitForNotifications:IdleTerm:", err)
  }
  return cmd
}

// Retrieve message ids after notificatioon
func (obj *GoImapClient)      RetrieveMessageIds() ([]uint32) {
  ids := []uint32{}
  for _, response := range obj.Client.Data {
    switch response.Label {
    case "EXISTS":
      ids = append(ids, imap.AsNumber(response.Fields[0]))
    }
  }
  obj.Client.Data = nil
  return ids
}

func (obj *GoImapClient)      ExpungeMessages(messages []GoImapMessage) (error) {
  set, _ := imap.NewSeqSet("")

  for _, message := range messages {
    set.AddNum(message.UID)
  }
  _, err := imap.Wait(obj.Client.UIDStore(set, "+FLAGS", imap.NewFlagSet(`\Deleted`)))
  if err != nil {
    return err
  }
  _, err = imap.Wait(obj.Client.Expunge(nil))
  if err != nil {
    return err
  }
  return err
}

// Retrieve messages from ids
func (obj *GoImapClient)      RetrieveMessagesFromIds(ids []uint32) ([]GoImapMessage) {
  messages := []GoImapMessage{}

  if len(ids) > 0 {
    set, _ := imap.NewSeqSet("")
    set.AddNum(ids...)

    cmd, err := imap.Wait(obj.Client.Fetch(set, "UID", "RFC822.HEADER", "RFC822.TEXT"))
    if err != nil {
      fmt.Println("RetrieveMessagesFromIds:Fetch:", err)
      return messages
    }

    for _, msg := range cmd.Data {
      message, err := NewMessage(msg.MessageInfo().Attrs)
      if err != nil {
        fmt.Println("RetrieveMessagesFromIds:NewMessage:", err)
        return messages
      }
      messages = append(messages, *message)
    }
    return messages
  }

  return messages
}

func init() {
  imap.DefaultLogger  = log.New(os.Stdout, "", 0)
  imap.DefaultLogMask = imap.LogConn | imap.LogRaw
}

// Constructor
func NewClient(host string, port uint32, login string, pw string) (*GoImapClient) {
  return &GoImapClient{
    Hostname: host,
    Port:     port,
    User:     login,
    Password: pw,
  }
}

// GoImapMessage Object
type GoImapMessage struct {
  UID     uint32

  // header
  Subject string
  To      string
  From    string
  Date    string

  // content
  Body    string
}

// Constructors
func NewMessage(attrs imap.FieldMap) (*GoImapMessage, error) {
  m, err := mail.ReadMessage(bytes.NewReader(imap.AsBytes(attrs["RFC822.HEADER"])))
  if err != nil {
    return nil, err
  }
  newMessage := GoImapMessage{
    UID:      imap.AsNumber(attrs["UID"]),
    Body:     string(imap.AsBytes(attrs["RFC822.TEXT"])),
    Subject:  m.Header.Get("Subject"),
    To:       m.Header.Get("To"),
    From:     m.Header.Get("From"),
    Date:     m.Header.Get("Date"),
  }
  return &newMessage, err
}

func NewMessageFromBytes(bytes []byte) (*GoImapMessage) {
  newMessage := &GoImapMessage{}
  return newMessage.DecodeFromBytes(bytes)
}

func (obj *GoImapMessage) Dump() {
  fmt.Println("-----------Dump message-----------")
  fmt.Println("UID: ",      obj.UID)
  fmt.Println("Subject: ",  obj.Subject)
  fmt.Println("To: ",       obj.To)
  fmt.Println("From: ",     obj.From)
  fmt.Println("Date: ",     obj.Date)
  fmt.Println("Body: ",     obj.Body)
  fmt.Println("----------------------------------")
}

func (obj *GoImapMessage) EncodeAsBytes() ([]byte) {
  out, err := json.Marshal(*obj)
  if err != nil {
    fmt.Println("EncodeAsBytes:", err)
  }
  return out
}

func (obj *GoImapMessage) DecodeFromBytes(message []byte) (*GoImapMessage) {
  msg := GoImapMessage{}
  err := json.Unmarshal(message, &msg)
  if err != nil {
    fmt.Println("DecodeFromBytes:", err)
  }
  return &msg
}