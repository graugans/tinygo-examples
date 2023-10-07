#!/bin/sh

# Remove any leftover settings from the TinyGo extension
# Myabe some jq clever hack is more appropriate
rm -f ./.vscode/settings.json

# print the Go version
go version

# Install Bash Autocompletion
sudo apt update
sudo apt install bash-completion

TINY_GO_VERSION="0.30.0"
# Install Tiny Go
TINY_GO_DEB="tinygo_${TINY_GO_VERSION}_amd64.deb"
wget https://github.com/tinygo-org/tinygo/releases/download/v${TINY_GO_VERSION}/${TINY_GO_DEB}
sudo dpkg -i ${TINY_GO_DEB} && rm ${TINY_GO_DEB}

export PATH=$PATH:/usr/local/bin

go get -u github.com/spf13/cobra@latest
go install github.com/spf13/cobra-cli@latest
go install github.com/godoctor/godoctor@latest
go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/sago35/tinygo-autocmpl@latest


echo 'eval "$(tinygo-autocmpl --completion-script-bash)"' >> ~/.bashrc