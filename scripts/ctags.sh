#!/bin/bash

# Check if universal ctags is installed
if ! ctags --version | grep -q "Universal Ctags"; then
    echo "Universal Ctags not found. Installing..."
    # Add the installation commands for your specific OS
    # For Ubuntu, you might do something like this:
    sudo apt-get update
    sudo apt-get install autoconf
    sudo apt-get install pkg-config
    git clone https://github.com/universal-ctags/ctags.git
    cd ctags
    ./autogen.sh
    ./configure
    make
    sudo make install
else
    echo "Universal Ctags is already installed."
fi