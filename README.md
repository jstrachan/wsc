# wsc

A simplistic tool for sending and receiving websocket messages from a command line.
Mainly useful to test websocket servers.

Forked from [github.com/raphael/wsc](https://github.com/raphael/wsc) to add pretty printing and color to the JSON that comes back

Getting started:
```
$ go get github.com/jstrachan/wsc
$ wsc -o http://websocket.org -H "Sample-Header-1: foo" -H "Sample-Header-2: bar" -u ws://echo.websocket.org
2016/03/08 22:51:51 connecting to ws://echo.websocket.org...
2016/03/08 22:51:52 ready, exit with CTRL+C.
foo 
>> foo
<< foo
^C
exiting
```
