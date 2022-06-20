package main

import (
	"context"
	"distributed/log"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stdlog "log"
)

func main() {
	log.Run("./distributed.log")
	host, port := "localhost", "4000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{
		ServiceName: registry.LogService,
		ServiceURL: serviceAddress,
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: serviceAddress + "/services",
	}
	ctx, err := service.Start(context.Background(), host, port, r, log.RegisterHandlers)

	if err != nil {
		stdlog.Fatalln(err)
	}
	<- ctx.Done()  //等待ctx停止信号，由service.go文件中的cancel函数发出

	fmt.Println("Shutting down log service.")
}