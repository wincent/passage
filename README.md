# Passage

## Overview

Passage (short for "Pass Agent") is a Go process that serves as caching proxy for macOS keychain items.

It is intended to fulfil the use case where you have a command-line process that needs to periodically access a passphrase in the keychain, but you don't want to have to click "Allow" every single time (too cumbersome), nor "Always Allow" (too coarse). For more information, see **Security**, below.

## Setup

### Installing

If you have a working Go environment and `$GOPATH` is set, you can do:

```
go get github.com/wincent/passage
```

Additionally, you will need to install the [keybase/go-keychain](https://github.com/keybase/go-keychain) library as a dependency:

```
go get github.com/keybase/go-keychain
```

To build manually:

```
cd $GOPATH/src/wincent/passage
go build
```

## Running automatically as a launch agent

```
sudo cp passage /usr/local/bin
cp contrib/com.wincent.passage.plist ~/Library/LaunchAgents

# NOTE: This won't work if run inside a tmux session:
launchctl load -w -S Aqua ~/Library/LaunchAgents/com.wincent.passage.plist
```

## Usage

Send a JSON object with keys "service" and "account" to the Passage socket at `~/.passage.sock`. On the first request, you may see a keychain permission dialog, but on subsequent requests, you'll get a cached result back immediately.

```
echo '{"service":"test item","account":"test name"}' | nc -U ~/.passage.sock
```

The resulting password is written back over the socket. You could call this a "write of Passage" (ha!). In the event that the item is missing or there is an error, the connection closes without any output, although errors do get printed to the standard output of the Passage process itself. This is for convenient consumption by clients (ie. so that they don't have to implement JSON parsing themselves).

### Invalidating the cache

There are two ways to invalidate the cache. One is to just restart the Passage process. The other is to send it a `SIGUSR1` signal.

You'll need the process PID in order to do that, and you can get it in several ways:

```
# The old-fashioned way:
ps xww | grep passage

# Or to be more specific, assuming you installed and ran from `/usr/local/bin/`:
ps xww | grep /usr/local/bin/passage | grep -v grep

# The first column so you can extract it like this:
ps xww | grep /usr/local/bin/passage | grep -v grep | awk '{print $1}'

# If you launched using launchctl (but note: only works outside tmux):
launchctl list | grep passage

# The PID is the first column in the output of `launchctl list`, so you can
# extract it like this:
launchctl list | grep passage | awk '{print $1}'

# Once you've got the PID, say in a variable $PID, send the signal like this:
kill -USR1 $PID

# A more clever way using launchctl that works anywhere (including inside tmux)
# and doesn't just get the PID but actually sends the signal as well:
launchctl kill SIGUSR1 "gui/$(id -u)/com.wincent.passage"
```

### Configuration

There are no options. This was quickly hacked together to address a specific need.

## Security

The `~/.passage.sock` socket is created with user-only (`0700`) permissions, but can still be accessed by privileged users and processes running on the system. As such, it falls somewhere in the middle of the spectrum of password storage (ordered approximately from most to least secure):

* **Storage in the system keychain, requiring explicit "Allow" intervention for each access:** This is the most secure, but also the most onerous. Key material only leaves the key chain with active approval (interaction with an intrusive GUI dialog).
* **Temporary storage in a process like `ssh-agent`:** This is quite secure not only because of the limited storage duration, but because the key material *never* leaves the agent; rather, other processes delegate work to be done by the agent. For example, when `ssh` tries to connect to a server, it will ask `ssh-agent` to perform the operations for private-key authentication, but there is no straightforward way to extract the actual private key material for the agent. The most a local attacker can do is perform operations as you, and only while they have access to the agent. Note, however, this approach is not viable for situations like those in which you *do* need access to the private key.
* **Temporary storage in the memory of the `passage` process:** Passphrases can only be inserted into `passage` via explicit approval, and only remain there for the duration of the session. You can shorten the duration in various ways, such as:
  1. Terminating the app or logging out.
  2. Setting up a cron job to periodically send a `SIGUSR1` to `passage` to cause it to drop its cache; you can do this with `crontab -e` and a method like the `launchctl kill` one documented above.
  3. Arranging to send `SIGUSR1` whenever the screensaver kicks in or the screen locks (something you should be able to do with a tool like [Hammerspoon](http://www.hammerspoon.org/)).
* **Storage in the system keychain, having previously granted "Always Allow" access:** This is problematic if you're using the `security` command-line tool provided by Apple, because once you've granted "Always Allow", you will never be asked for permission again. Even if you intended to grant access only to `my-special-script.sh`, the permission will actually be associated with the `security` tool, which any process can call. So here the security is contingent on an attacker not being able to run code as you (or trick/coerce you into running it).
* **Storage in plain-text on the filesystem:** The security here is only so good as the security of the filesystem: privileged users will be able to retrieve the password regardless of filesystem permission. An unprivileged, determined attacker with local access will be able to elevate privileges by one means or another and gain access, even in the face of restrictive permissions.

## Troubleshooting

## Limitations

Passage uses the [keybase/go-keychain](https://github.com/keybase/go-keychain) library to access the keychain, which currently only knows how to read "generic" (A.K.A. "application") passwords, not "Internet" passwords.

# Authors

Passage is written and maintained by Greg Hurrell (greg@hurrell.net).

## Development

Mirrors exist at:

- https://github.com/wincent/passage
- https://gitlab.com/wincent/passage
- https://bitbucket.org/ghurrell/passage

Patches are welcome via the usual mechanisms (pull requests, email, posting to the project issue tracker etc).

## Website

The official website for Passage is:

- https://github.com/wincent/passage

Bug reports should be submitted to the issue tracker at:

- https://github.com/wincent/passage/issues

## History

Please see [`CHANGELOG.md`](CHANGELOG.md) in this repository.

## License

Copyright 2016-present Greg Hurrell. All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDERS OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
