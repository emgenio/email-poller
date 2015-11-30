##
## Simple Makefile for Email-Poller
## 
## Made by axel catusse
##


# vars
GOCMD = go
GOGET = go get
GOBUILD = $(GOCMD) build
GOCLEAN	= $(GOCMD) clean
GOINSTALL = $(GOCMD) install

# dirs
BUILD_DIR = ./build
CONFIG_DIR = ./config
EMAIL_POLLER_DIR= ./email-poller

# srcs
CONFIG_SRC = $(CONFIG_DIR)/config.go

EMAIL_POLLER_SRC = $(EMAIL_POLLER_DIR)/email-poller.go

# bins
EMAIL_POLLER_BIN = email-poller

all: get-package
	$(GOBUILD) -o $(BUILD_DIR)/$(EMAIL_POLLER_BIN) $(EMAIL_POLLER_SRC)

clean:
	rm -rf $(GOBUILD)/*

re: clean all

get-package:
	$(GOGET) github.com/emgenio/email-poller/imap
	$(GOGET) github.com/streadway/amqp

