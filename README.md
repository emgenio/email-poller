# Email-poller
An IMAP email poller written in go to send emails into a Rabbit MQ message queue.

## How it works
When started, email poller sets the IMAP client on an Idle stance (IMAP Protocol feature), waiting for notifications. When the mailbox monitoring receives new incoming messages, it serializes them into raw datas and publishes them to a rabbit MQ message queue.

The Poller is based on this [go package IMAP client](https://github.com/mxk/go-imap).

## How to install
1. First, you need to have two configurated servers (both SMTP and IMAP). I would suggest you [Postfix](http://www.postfix.org/) and [dovecot](http://www.dovecot.org/) which are the most common used servers for this kind of setup.
2. Then get the package: `go get github.com/emgenio/email-poller`
3. Finally run the command: `email-poller`

## Usage
```
email-poller -h
Usage of email-poller:
  -config="./config.yaml": path to the configuration file.
```

## Licence
MIT