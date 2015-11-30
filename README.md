# Email-poller
An IMAP email poller written in go to send emails into a Rabbit MQ message queue.

## How is works
When started, email poller sets the IMAP client on an Idle stance (IMAP Protocol feature), waiting for notifications. When the mailbox monitoring receives new incoming messages, it serializes them into raw datas and publishes them to a rabbit MQ message queue.

The Poller is based on this [go package IMAP client](https://github.com/mxk/go-imap).

# The Worker
The worker has the role of the Consumer. It instanciates a Rabbit MQ queue specified in the worker yaml file, consume all message that have been push into this queue and forward them to their destination via [Mandrill](https://www.mandrill.com/) (an email delivery api from Mailchimp).

# Running the stack
First, you need to have a configurated SMTP and IMAP server. I would sugger you [Postfix](http://www.postfix.org/) and [dovecot](http://www.dovecot.org/) which are the most common used servers for this kind of setup.

Clone the git repository:
```
git clone https://github.com/emgenio/email-poller && cd email-poller
```
Then you just need to compile both binaries and you should be ready to go:
```
make
./build/email-poller &
./build/worker
```

## Licence
MIT