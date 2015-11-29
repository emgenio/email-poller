##
## Simple Makefile for Email-Poller
## 
## Made by axel catusse
##


# vars
GOCMD							=			go
GOGET							=			go get
GOBUILD						=			$(GOCMD) build
GOCLEAN						=			$(GOCMD) clean
GOINSTALL					=			$(GOCMD) install

# dirs
BUILD_DIR					=			./build
CONFIG_DIR				=			./config
EMAIL_POLLER_DIR	=			./email-poller
WORKER_DIR				=			./worker

# srcs
CONFIG_SRC				=			$(CONFIG_DIR)/config.go

EMAIL_POLLER_SRC	=			$(EMAIL_POLLER_DIR)/email-poller.go

WORKER_SRC				=			$(WORKER_DIR)/worker.go

# bins
EMAIL_POLLER_BIN	=			email-poller
WORKER_BIN				=			worker

all: get-package
	$(GOBUILD) -o $(BUILD_DIR)/$(EMAIL_POLLER_BIN) $(EMAIL_POLLER_SRC)
	$(GOBUILD) -o $(BUILD_DIR)/$(WORKER_BIN) $(WORKER_SRC)

clean:
	rm -rf $(GOBUILD)/*

re: clean all

get-package:
	$(GOGET) github.com/emgenio/email-poller/imap