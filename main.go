package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	goSystemd "github.com/alirezasn3/go-systemd"
)

func main() {
	address := flag.String("address", "", "target address")
	timeout := flag.Int("timeout", 500, "dial timeout in ms")
	delay := flag.Int("delay", 3000, "delay after successfull restart in ms")
	interval := flag.Int("interval", 3000, "how often to check address in ms")

	flag.Parse()

	if *address == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// check for install and uninstall commands
	if slices.Contains(os.Args, "--install") {
		execPath, err := os.Executable()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		execPath += " " + strings.Join(os.Args[1:], " ")
		err = goSystemd.CreateService(&goSystemd.Service{Name: "xray-restart", ExecStart: execPath, Restart: "on-failure", RestartSec: "3s"})
		if err != nil {
			log.Println(err)
			os.Exit(1)
		} else {
			log.Println("xray-restart service created")
			os.Exit(0)
		}
	} else if slices.Contains(os.Args, "--uninstall") {
		err := goSystemd.DeleteService("xray-restart")
		if err != nil {
			log.Println(err)
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
