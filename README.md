# Email-poller
An IMAP email poller written in go to send emails into a Rabbit MQ message queue.

## How it works
When started, email poller sets the IMAP client on an Idle stance (IMAP Protocol feature), waiting for notifications. When the mailbox monitoring receives new incoming messages, it serializes them into raw datas and publishes them to a rabbit MQ message queue.

The Poller is based on this [go package IMAP client](https://github.com/mxk/go-imap).

## How to install
1. First, you need to have a configurated SMTP and IMAP server. I would suggest you [Postfix](http://www.postfix.org/) and [dovecot](http://www.dovecot.org/) which are the most common used servers for this kind of setup.

2. Then clone the git repository:
  ```
  git clone https://github.com/emgenio/email-poller && cd email-poller
  ```
3. Finally compile and execute:
  ```
  make && ./build/email-poller
  ```

## Licence
MIT