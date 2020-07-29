FROM fedora:29
MAINTAINER Oliver Kahrmann <oliver.kahrmann@gmail.com>

##
# Base Setup
##
# Update, Install utilities, mingw64 + windows libraries
RUN dnf -y update; \
    dnf -y install bash file gcc make pkg-config tar curl unzip git rsync \
    mingw64-gcc mingw64-pkg-config mingw64-crt mingw64-headers mingw64-binutils \
	libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel; \
    dnf clean all

RUN useradd -ms /bin/bash builduser

##
# Golang
##

# Install go + tools
ENV GOPATH="/home/builduser/go"
ENV PATH="$PATH:/usr/local/go/bin:$GOPATH/bin"

RUN curl -sS 'https://storage.googleapis.com/golang/go1.14.linux-amd64.tar.gz' | tar xz -C /usr/local/ && \
    mkdir -p "$GOPATH/bin"
RUN go get -u github.com/goware/modvendor

RUN chown -R builduser:builduser /home/builduser/go

##
# Use a non-root user
##
USER builduser
