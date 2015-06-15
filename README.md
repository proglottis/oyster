Oyster
======

GPG password store.

## Installation

Oyster encrypts your passwords with your GPG key. If you do not have GPG key already, [set one up](https://www.gnupg.org/gph/en/manual.html#AEN26) before installed Oyster.

### OSX

You must have a GPG key setup before installing Oyster.

```bash
brew tap proglottis/oyster
brew install oyster
```

### Chrome Extension

* [Install](https://chrome.google.com/webstore/detail/knchgkoimfkgfopjfehdkcchmbmkmfgi)

### Post Install

Make sure your password repository is setup.

```bash
oyster init <your gpg key ID or email>
```

By default the passwords will be GPG encrypted in `~/.oyster/`, this default can be changed in the configuration file `~/.oysterconfig`. All settings in this file are currently optional.

```ini
home = /Users/john/Google Drive/oyster
gpgHome = /Volumes/Johns USB/.gnupg
```
