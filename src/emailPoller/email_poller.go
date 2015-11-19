package emailPoller

import (
  "imapWrapper"
  "time"
  "fmt"
)

const (
  TimeOut = time.Minute * 30
)

type emailPoller struct {
  ImapClient *imapWrapper.ImapWrapper
  MessagesToQ chan[]imapWrapper.IWMessage

  error error
}

func (obj *emailPoller) Initialize() {
  obj.ImapClient.Connect()
  obj.ImapClient.Login()
  obj.ImapClient.SelectBox(obj.ImapClient.Config.Mbox, false)
}

func (obj *emailPoller) PushMessagesToQueue() {
  for {
    messages := <- obj.MessagesToQ
    cpt := 0
    for _, message := range messages {
      // push to RABBIT
      message.Display()
      cpt += 1
    }
    fmt.Println("Number of messages that have been push to RabbitMQ:", cpt)
  }
}

func (obj *emailPoller) Start() {
  go obj.PushMessagesToQueue()

  ids := []uint32{}
  for {
    // ids, _ = obj.ImapClient.ListenIncomingMessages(TimeOut)
    // obj.ImapClient.Client.Idle()
    // obj.ImapClient.Client.Recv(TimeOut)
    // imap.Wait(obj.ImapClient.Client.IdleTerm())
    // iwmsg, _ := obj.ImapClient.BuildIWMessagesFromIds(ids)
    // obj.ImapClient.ExpungeMessageFromIds(ids)
    time.Sleep(1 * time.Second)
    fmt.Println("Waiting for Incoming Messages...")
    iwmsg, _ := obj.ImapClient.FetchAllMessages()
    if len(iwmsg) > 0 {
      obj.MessagesToQ <- iwmsg
    }
  }

  if false {
    _ = ids
  }
}

func Stop() {}
func init() {}

func Create(cfg string) (*emailPoller) {
  return &emailPoller{
    ImapClient: imapWrapper.Create(cfg),
    MessagesToQ: make(chan []imapWrapper.IWMessage),
  }
}