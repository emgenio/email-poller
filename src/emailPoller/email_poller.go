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
      _ = message
      cpt += 1
    }
    fmt.Println("Number of messages that have been push to RabbitMQ:", cpt)
  }
}

func (obj *emailPoller) Start() {
  go obj.PushMessagesToQueue()

  ids := []uint32{}
  for {
    ids, _ = obj.ImapClient.ListenIncomingMessages(TimeOut)
    iwmsg, _ := obj.ImapClient.BuildIWMessagesFromIds(ids)
    obj.MessagesToQ <- iwmsg
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