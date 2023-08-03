package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/kardianos/service"
	"go.uber.org/automaxprocs/maxprocs"
	"net"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

const version = "2022.3.2.31"
const serviceName = "Device Tester"
const serviceDescription = "Downloads data from devices using sockets"

var (
	serviceSync      sync.Mutex
	serviceIsRunning = false
)

type program struct{}

func main() {
	maxprocs.Set()
	functionName := getFunctionName()
	fmt.Println(color.Green, "INF", functionName, serviceName, version, "starting...", color.Reset)
	fmt.Println(color.Green, "INF", functionName, "©", time.Now().Year(), "Petr Jahoda", color.Reset)
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		fmt.Println(color.Red, "ERR", functionName, err.Error(), color.Reset)
	}
	err = s.Run()
	if err != nil {
		fmt.Println(color.Red, "ERR", functionName, err.Error(), color.Reset)
	}
}
func (p *program) Start(service.Service) error {
	functionName := getFunctionName()
	fmt.Println(color.Green, "INF", functionName, serviceName, version, "started", color.Reset)
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
	functionName := getFunctionName()
	fmt.Println(color.Green, "INF", functionName, serviceName, version, "stopped", color.Reset)
	return nil
}

func (p *program) run() {
	functionName := getFunctionName()
	fmt.Println(color.Green, "INF", functionName, "started", color.Reset)
	var deviceIpAddress string
	flag.StringVar(&deviceIpAddress, "ip", "192.168.0.1", "port number")
	flag.Parse()
	deviceIpAddress = deviceIpAddress + ":80"
	for {
		dialer := net.Dialer{Timeout: 5 * time.Second}
		conn, err := dialer.Dial("tcp", deviceIpAddress)
		if err != nil {
			fmt.Println(color.Red, "ERR", functionName, deviceIpAddress, err.Error(), color.Reset)
			serviceSync.Lock()
			serviceNowRunning := serviceIsRunning
			serviceSync.Unlock()
			if !serviceNowRunning {
				fmt.Println(color.Green, "INF", functionName, deviceIpAddress, "Communication ended, service is ending", color.Reset)
			}
			break
		}
		fmt.Println(color.Green, "INF", functionName, deviceIpAddress, "Socket connection with device opened", color.Reset)
		_, err = fmt.Fprintf(conn, time.Now().UTC().Format("2006-01-02;15:04:05")+"%\n")
		if err != nil {
			fmt.Println(color.Red, "ERR", functionName, deviceIpAddress, "Error sending date to device", color.Reset)
			conn.Close()
			break
		}
		for conn != nil {
			_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println(color.Green, "INF", functionName, deviceIpAddress, "Communication ended, problem reading data from device", color.Reset)
				conn.Close()
				break
			}
			fmt.Println(message)
			serviceSync.Lock()
			serviceNowRunning := serviceIsRunning
			serviceSync.Unlock()
			if !serviceNowRunning {
				fmt.Println(color.Green, "INF", functionName, deviceIpAddress, "Communication ended, service is ending", color.Reset)
				conn.Close()
				break
			}
		}
	}
}

func getFunctionName() string {
	pc, file, _, _ := runtime.Caller(1)
	wholeFuncName := runtime.FuncForPC(pc).Name()
	funcName := strings.Split(wholeFuncName, ".")[1]
	fileName := path.Base(file)
	return fmt.Sprintf("%s >> %s:", fileName, funcName)
}
