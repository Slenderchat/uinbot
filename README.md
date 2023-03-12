# UINBOT
## UIN submodule
### Description
This module scans mail and find mails with UIN's. It pulls VEDOMSTVO, all UIN's and their corresponding SUM's, writes it to PostgreSQL database, then reads database to find related data about objects and their owners, and sends all data in readable form to TG channel specified by `tgchatid` via bot specified by `tgtoken`.
### Build
Source files located in `uin` subdirectory of this repo. To build it you should follow this steps:
- Clone this repo:
```
git clone https://github.com/Slenderchat/uinbot.git
```
- Change directory to `uin` subdirectory of this repo:
```
cd uinbot/uin
```
- Run go build:
```
go build -ldflags "-s -w" uin.go
```
### Configuration
Configuration is done in `uinbot.json` file, which should be right next to the binary.
Currently following configuration options are supported:
- `tgtoken` - Telegram Bot API token
- `tgchatid` - Telegram chat id (where to send messages)
- `pgpassword` - PostgreSQL password

Current defaults which may not be overriden is:
- By default local UNIX socket located at `/run/postgresql` is used to connect to PostgreSQL
- Default user and database is `uinbot`
- It uses lock file for queuing parallel execution located at `/tmp/uinbot.lock`
### Usage
- Place the binary in `/etc/dovecot` folder
- Create file `uin.json` in the same directory as the binary with this content replacing example values with your own values
```
{
        "tgtoken": "TOKEN",
        "tgchatid": 123456789,
        "pgpassword": "mYsEcUrEpAsSwOrD9001"
}
```
- Add a rule to your sieve rules to trigger this binary. For example:
```
require ["fileinto", "vnd.dovecot.execute", "imap4flags"];
if allof (address :contains "From" "rr-info@notariat.ru", header :contains "Subject" "УИН по ") {
        addflag "\\Seen";
        fileinto "УИН";
        execute :pipe "uin";
        stop;
}
```
