# file-poller

## usage: 

> `file-poller -conf path/to/toml`

## config:

sample toml config file:
```
SourceDir = "c:\\temp\\incoming"
DestDir = "c:\\temp\\poll"
LogDir = "c:\\temp\\log"
LogLevel = "DEBUG"
```

## service:

download nssm and setup service