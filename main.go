package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	goSystemd "github.com/alirezasn3/go-systemd"
)

func main() {
	address := flag.String("address", "", "target address")
	timeout := flag.Int("timeout", 500, "dial timeout in ms")
	delay := flag.Int("delay", 3000, "delay after successfull restart in ms")
	interval := flag.Int("interval", 3000, "how often to check address in ms")
	install := flag.Bool("install", false, "install systemd service")
	uninstall := flag.Bool("uninstall", false, "uninstall systemd service")

	flag.Parse()

	if *address == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// check for install and uninstall commands
	if *install {
		execPath, e := os.Executable()
		if e != nil {
			log.Println(e)
			os.Exit(1)
		}

		execPath += fmt.Sprintf(" --address %s --timeout %d --delay %d --interval %d", *address, *timeout, *delay, *interval)

		e = goSystemd.CreateService(&goSystemd.Service{Name: "xray-restart", ExecStart: execPath, Restart: "on-failure", RestartSec: "5s"})
		if e != nil {
			log.Println(e)
			os.Exit(1)
		} else {
			log.Println("xray-restart service created")
			os.Exit(0)
		}
	} else if *uninstall {
		e := goSystemd.DeleteService("xray-restart")
		if e != nil {
			log.Println(e)
			os.Exit(1)
		} else {
			log.Println("xray-restart service deleted")
			os.Exit(0)
		}
	}

	log.Println("started")

	for {
		c, e := net.DialTimeout("tcp", *address, time.Millisecond*time.Duration(*timeout))
		if e != nil {
			if exec.Command("systemctl", "restart", "xray").Run() != nil {
				panic(e)
			} else {
				log.Println("xray restarted")
				time.Sleep(time.Millisecond * time.Duration(*delay))
			}
		} else {
			c.Close()
		}
		time.Sleep(time.Millisecond * time.Duration(*interval))
	}
}
