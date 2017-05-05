package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ldkingvivi/unix_domain_key_value/access"
	"github.com/ldkingvivi/unix_domain_key_value/unix"
)

var (
	configFile = flag.String("config", "", "Configuration file to use")
)

func main() {

	flag.Parse()

	if *configFile == "" {
		fmt.Fprint(os.Stderr, "Missing configuration file")
		flag.PrintDefaults()
		os.Exit(1)
	}

	content, _ := ioutil.ReadFile(*configFile)
	var conf unix.Config
	err := json.Unmarshal(content, &conf)
	if err != nil {
		fmt.Print("Erros:", err)
	}

	fmt.Println(conf)

	runtime.GOMAXPROCS(conf.Daemon.Cpus)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	s, errS := unix.InitUnixServer(conf)
	if errS != nil {
		log.Fatal(errS)
	}

	go func() {
		<-sigChan
		s.Stop()
		fmt.Println("Exiting ... ")
		os.Exit(0)
	}()

	d, errD := access.InitMap(s.UConnChan, conf.Daemon.Updateinterval, conf.Daemon.Sourceendpoint, conf.Daemon.Apitoken)
	if errD != nil {
		log.Fatal(errD)
	}

	d.Start()
	s.Start()

}
