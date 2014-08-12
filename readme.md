# TROJE

Troje is a honeypot that creates a real environment within a physical of virtual machine using lxc containers. These containers will be created on the first connection with the desired service. For example ssh service. At the first connection the 'attacker' will get its own attack vector, where Troje will pass all traffic between the service and their own attack vector. All traffic within the lxc container will be monitored, also the changes to the drives are being recorded.

WARNING: this is a proof of concept and hasn't been tested accordingly. 

This version is a proof of concept. With the proof of concept I want to test the following:

- an individual lxc container can be used per remote address
- the lxc container is safe enough and can be constrained for attacks to operate safely
- all actions can be monitored (traffic, filesystem, ssh) 
- the honeypot is realistic

## Quick start

### Install (ubuntu):

```
apt-get install golang lxc aufs-tools

go get http://github.com/lxc/go-lxc
go get http://code.google.com/p/gopacket
```

### Create base container:

```
lxc-create -t download -n troje_base -- --dist ubuntu --release trusty --arch amd64
```

### Start Troje:

```
GOPATH=`pwd` go run ./bin/main.go -b troje_base
```

Now Troje is up and running and you can connect using SSH to Troje. When you connect, the troje_base container will be cloned and all current and following connections from the remote address will be directed to the cloned container.

## Contributing

Contributions are welcome.

## Example

```
root@packer-vmware-iso:/vagrant# go run ./main.go -b u1
2014/08/12 11:40:55 Troje started.
2014/08/12 11:40:57 Received new connection from 172.16.84.1:53483
2014/08/12 11:40:57 Creating new container for ip 172.16.84.1:53483
2014/08/12 11:40:57 Cloning new container 3IeAsSTV
2014/08/12 11:40:57 Starting new container 3IeAsSTV
2014/08/12 11:40:57 Waiting for container to settle 3IeAsSTV
2014/08/12 11:40:57 Waiting for ip to settle 3IeAsSTV (getting IP address on the interface of the container failed)
2014/08/12 11:40:58 Waiting for ip to settle 3IeAsSTV (getting IP address on the interface of the container failed)
2014/08/12 11:40:59 Waiting for ip to settle 3IeAsSTV (getting IP address on the interface of the container failed)
2014/08/12 11:41:00 Waiting for ip to settle 3IeAsSTV (getting IP address on the interface of the container failed)
2014/08/12 11:41:01 Container 3IeAsSTV got ip 10.0.3.84
2014/08/12 11:41:01 Forwarding connections

<<< TRAFFIC BETWEEN CONTAINER AND REMOTE HOST >>>

2014/08/12 11:41:01 ^C2014/08/12 11:41:03 interrupt
2014/08/12 11:41:03 Troje stopping. Cleaning up.
2014/08/12 11:41:03 Destroying container 3IeAsSTV
2014/08/12 11:41:04 2014/08/12 11:41:04 Connection closed.
2014/08/12 11:41:04 Waiting for container to shutdown 3IeAsSTV
2014/08/12 11:41:04 Troje stopped.
```

## Todo: 

- create hot spare containers
- serialize data
- reporting of traffic (pcap)
- reporting of delta (/var/lib/lxc/gvgVHiMV/delta0/)
- how to monitor the commands? using ssh and created username / password?
- custom ssl certicate? for intercepting https traffic?
- how to compare the differences? Use overlayfs?
- when to clean up and create report?
- listen on multiple ip adresses, with different containers
- configure custom forwarding
- use pipes / multithreading / locks 
- code improvements
- start container prevent goroutine to return
- should we use ephemeral storage?
- Go-ify source

## Creators 

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>

## Copyright and license

Code and documentation copyright 2011-2014 Remco Verhoef. Code released under [the MIT license](LICENSE). 

