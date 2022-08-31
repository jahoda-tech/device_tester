package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/kardianos/service"
	"net"
	"strconv"
	"sync"
	"time"
)

const version = "2022.3.2.31"
const serviceName = "Device Tester"
const serviceDescription = "Downloads data from devices using sockets"

var (
	serviceIsRunning = false
	serviceSync      sync.Mutex
)

type program struct{}

func main() {
	fmt.Println(color.Ize(color.Green, "INF [MAIN] "+serviceName+" ["+version+"] starting..."))
	fmt.Println(color.Ize(color.Green, "INF [MAIN] © "+strconv.Itoa(time.Now().Year())+" Petr Jahoda"))
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		fmt.Println(color.Ize(color.Red, "ERR [MAIN] Cannot start: "+err.Error()))
	}
	err = s.Run()
	if err != nil {
		fmt.Println(color.Ize(color.Red, "ERR [MAIN] Cannot start: "+err.Error()))
	}
}
func (p *program) Start(service.Service) error {
	fmt.Println(color.Ize(color.Green, "INF [MAIN] "+serviceName+" ["+version+"] started"))
	go p.run()
	serviceSync.Lock()
	serviceIsRunning = true
	serviceSync.Unlock()
	return nil
}

func (p *program) Stop(service.Service) error {
	serviceSync.Lock()
	serviceIsRunning = false
	serviceSync.Unlock()
	fmt.Println(color.Ize(color.Green, "INF [MAIN] "+serviceName+" ["+version+"] stopped"))
	return nil
}

func (p *program) run() {
	var deviceIpAddress string
	flag.StringVar(&deviceIpAddress, "ip", "192.168.0.1", "port number")
	flag.Parse()
	deviceIpAddress = deviceIpAddress + ":80"
	for {
		dialer := net.Dialer{Timeout: 5 * time.Second}
		conn, err := dialer.Dial("tcp", deviceIpAddress)
		if err != nil {
			fmt.Println(color.Ize(color.Red, "ERR ["+deviceIpAddress+"] Problem opening socket connection: "+err.Error()))
			serviceSync.Lock()
			serviceNowRunning := serviceIsRunning
			serviceSync.Unlock()
			if !serviceNowRunning {
				fmt.Println(color.Ize(color.Green, "INF ["+deviceIpAddress+"] Communication ended, service is ending"))
			}
			break
		}
		fmt.Println(color.Ize(color.Green, "INF ["+deviceIpAddress+"] Socket connection with device opened"))
		_, err = fmt.Fprintf(conn, time.Now().UTC().Format("2006-01-02;15:04:05")+"%\n")
		if err != nil {
			fmt.Println(color.Ize(color.Red, "ERR ["+deviceIpAddress+"] Error sending date to device"))
			conn.Close()
			break
		}
		for conn != nil {
			_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(color.Ize(color.Green, "INF ["+deviceIpAddress+"] Communication ended, problem reading data from device"))
				conn.Close()
				break
			}
			fmt.Println(message)
			serviceSync.Lock()
			serviceNowRunning := serviceIsRunning
			serviceSync.Unlock()
			if !serviceNowRunning {
				fmt.Println(color.Ize(color.Green, "INF ["+deviceIpAddress+"] Communication ended, service is ending"))
				conn.Close()
				break
			}
		}
	}
}
