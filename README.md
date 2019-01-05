# gopilot - A linux server monitoring and managing tool in go

Hello out there, this project aims an easy-to-use web-based monitoring and administration tool for linux systems.

IT IS NOT READY YET, DONT USE IT ON PRODUCTION AND FEEL FREE TO TEST IT AND GIVE ME FEEDBACK.

### Just, Why ?
Yes, there are many tools out there to manager your linux-systems, but for me it was difficult to setup, all uses
some kind of web-server and magic behind the scenes. So i started this project mainly to manage my home servers and
my servers on the internet from one single Interface. And hey, maybe somebody can use it also :)
There is an angularjs webinterface on my git-repo. This interface is able to select an external node without change of ui/ux,
just switch, manage and switch back

## Rewrite
This is a rewrite of the copilot-project in go, its nearly ready for first use.

## Build
First, set the gopath to local directory
* `export GOPATH=${PWD}:$(go env GOPATH)`

Then we need to install dependencies
* go get

Then build it
* `go build -x -v -o gopilot src/copilotg.go`

## Thats work
 * Node to Server connection over TLS with cherry pick of certificate and challange request-response
 * Websocket-Server for web-client
 * Webserver to serve web-client files
 * Message-BUS with worker ( super easy in golang :D )
 
## And a lot more todo
 * Health-Plugin ( to monitor server-health )
 * Collectd-Support for Health-Plugin
 * LDAP-Client
 * NFT-Plugin ( Firewall )
 * Doku ( yeah thats always the thing )
 * Doku - How to write an plugin
 * Doku - How to use the message-bus







