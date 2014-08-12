package main

import (
    "io"
    "log"
    "sync"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "crypto/rand"
    "time"
    "net"
    "github.com/lxc/go-lxc"
    _ "code.google.com/p/gopacket"
    _ "code.google.com/p/gopacket/pcap"
)

// An uninteresting service.
type Troje struct {
}

var containers map[string]*lxc.Container = make(map[string]*lxc.Container)

func randString(n int) string {
    const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
    var bytes = make([]byte, n)
    rand.Read(bytes)
    for i, b := range bytes {
        bytes[i] = alphanum[b % byte(len(alphanum))]
    }
    return string(bytes)
}

func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
   buf := make([]byte, 32*1024)
   for {
       nr, er := src.Read(buf)

       log.Printf("%s", buf[:nr])

       if nr > 0 {
           nw, ew := dst.Write(buf[0:nr])

           if nw > 0 {
               written += int64(nw)
           }

           if ew != nil {
               err = ew
               break
           }

           if nr != nw {
               err = io.ErrShortWrite
               break
           }
       }

       if er == io.EOF {
           break
       }

       if er != nil {
           err = er
           break
       }
   }

   return written, err
}

func forward(localConn net.Conn) {
    defer localConn.Close()

    log.Printf("Received new connection from %s\n", localConn.RemoteAddr())

    c2 := containers[localConn.RemoteAddr().Network()]
    if c2 == nil {
        log.Printf("Creating new container for ip %s\n", localConn.RemoteAddr())
        c2 = GetContainer()
        containers[localConn.RemoteAddr().Network()] = c2
    }

    // Get interface name
    var interfaceName string

    for i := 0; i < len(c2.ConfigItem("lxc.network")); i++ {
        interfaceType := c2.RunningConfigItem(fmt.Sprintf("lxc.network.%d.type", i))

        if interfaceType == nil {
                continue
        }

        if interfaceType[0] == "veth" {
                interfaceName = c2.RunningConfigItem(fmt.Sprintf("lxc.network.%d.veth.pair", i))[0]
        } else {
                interfaceName = c2.RunningConfigItem(fmt.Sprintf("lxc.network.%d.link", i))[0]
        }
    }

    _ = interfaceName

    // Get ipaddress
    ip, err := c2.IPAddress("eth0"); 

    if err != nil {
            log.Printf(err.Error())
            return
    }

    dest := fmt.Sprintf("%s:22", ip[0])

    /*
    // Start packet capture for monitoring (outgoing) container traffic
    if handle, err := pcap.OpenLive(interfaceName, 1600, true, 0); err != nil {
      panic(err)
    } else if err := handle.SetBPFFilter("tcp and port 80"); err != nil {  // optional
      panic(err)
    } else {
      go func(handle *pcap.Handle) {
          packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
          for packet := range packetSource.Packets() {
              log.Printf("PACKET: %s", packet)
          }
      }(handle)
    }
    */

    // Start forwarding service
    log.Println("Forwarding connections")

    var conn net.Conn

    for {
        conn, err = net.Dial("tcp", dest)
        defer conn.Close()

        if err != nil {
            log.Println("Waiting for ssh connection to settle.")

            time.Sleep(time.Millisecond * time.Duration(250))
            continue
        }

        break
    }

    var wg sync.WaitGroup

    wg.Add(1)
    go func() {
        defer wg.Done()

        _, err := Copy(conn, localConn)
        if err != nil {
            log.Fatalf("io.Copy failed: %v", err)
        }

    }()

    wg.Add(1)
    go func() {
        defer wg.Done()

        _, err := Copy(localConn, conn)
        if err != nil {
            log.Fatalf("io.Copy failed: %v", err)
        }

    }()
    
    wg.Wait()

    log.Printf("Connection closed.\n")

    // send email with delta of:
    // /var/lib/lxc/gvgVHiMV/delta0/
    // and tcpdump
}


func GetContainer() *lxc.Container {
    c, err := lxc.NewContainer("u1")
    if err != nil {
        log.Fatalf(err.Error())
    }

    defer lxc.PutContainer(c)

    c2_name := randString(8)

    log.Printf("Cloning new container %s\n", c2_name)
    err = c.CloneUsing(c2_name, lxc.Aufs, lxc.CloneSnapshot) // #c.CloneUsing("u6_2")
    if err != nil {
        log.Fatalf(err.Error())
    }

    c2, err := lxc.NewContainer(c2_name)
    if err != nil {
        log.Fatalf(err.Error())
    }

    defer lxc.PutContainer(c2)

    // TODO: StartContainer prevents goroutine to return properly
    log.Printf("Starting new container %s\n", c2_name)
    if err := c2.Start(); err != nil {
        log.Fatalf(err.Error())
    } 

    log.Printf("Waiting for container to settle %s\n", c2_name)
    if !c2.Wait(lxc.RUNNING, 30) {
        log.Fatalf("Container still not running %s\n", c2_name)
    }

    var dest string

    for {
        ip, err := c2.IPAddress("eth0"); 

        if err != nil {
            log.Printf("Waiting for ip to settle %s (%s)\n", c2_name, err.Error())
            time.Sleep(time.Millisecond * time.Duration(1000))
            continue
	}

        dest = ip[0]
        break
    }

    log.Printf("Container %s got ip %s\n", c2_name, dest)

    // increase container reference
    lxc.GetContainer(c2)
    return (c2)
}

var base = flag.String("b", "", "The name of the lxc base container")
func main() {
    flag.Parse()

    if *base == "" {
	log.Fatal("No base container defined.")
    }

    log.Printf("Troje started.")

    // TODO: queue instances for warm start

    // TODO: forward multiple ports
    // var waitGroup *sync.WaitGroup
    var	ch2 chan bool


    go func() {
        listener, err := net.Listen("tcp", ":8022")
        if err != nil {
            log.Fatalf("net.Listen failed: %v", err)
        }

        defer listener.Close()

        for {
            conn, err := listener.Accept()
            if err != nil {
                log.Fatalf("listen.Accept failed: %v", err)
            }
            
            select {
            case <-ch2:
                    log.Println("stopping listening on", listener.Addr())
                    listener.Close()
                    return
            default:
            }


            // serve
            go forward(conn)
        }
    }()

    // TODO: housekeeping, cleaunup containers when inactive for x period

    ch := make(chan os.Signal)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    log.Println(<-ch)

    log.Printf("Troje stopping. Cleaning up.")

    for _, container := range containers {
        log.Printf("Destroying container %s", container.Name())
        if err := container.Shutdown(30); err != nil {
            log.Fatalf(err.Error())
        }

        log.Printf("Waiting for container to shutdown %s\n", container.Name())
        if !container.Wait(lxc.STOPPED, 30) {
            log.Fatalf("Container still not running %s\n", container.Name())
        }

        if err := container.Destroy(); err != nil {
            log.Fatalf(err.Error())
        }

        lxc.PutContainer(container)
    }

    log.Printf("Troje stopped.")
}


