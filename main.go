package main

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"studio.sunist.work/platform/alioth-center/core/restoration"
	log "studio.sunist.work/platform/alioth-center/infrastructure/global/utils/logger"
	"studio.sunist.work/platform/alioth-center/infrastructure/initialize"
)

func main() {
	logger := log.DefaultLogger()
	lis, err := net.Listen("tcp",
		fmt.Sprintf("%s:%d", initialize.GlobalConfig().Grpc.ListenIP, initialize.GlobalConfig().Grpc.ListenPort))
	if err != nil {
		panic(err)
	}

	// 初始化rpc和http服务器
	s := grpc.NewServer()
	engine := gin.New()
	engine.Use(gin.Recovery())
	external := engine.Group("/external")

	// 注册rpc和http服务器
	restoration.InitRestorationRpcServer(s)
	restoration.InitRestorationHttpServer(external)

	// 启动rpc和http服务器
	rpcExit := startRpcEngine(s, lis)
	httpExit := startHttpEngine(engine,
		fmt.Sprintf("%s:%d", initialize.GlobalConfig().Http.ListenIP, initialize.GlobalConfig().Http.ListenPort))
	logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Info).WithMessage("server(s) started"))

	// 等待rpc和http服务器退出
	select {
	case err := <-rpcExit:
		logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Panic).
			WithMessage("rpc server exit").WithExtra(err.Error()))
	case err := <-httpExit:
		logger.Log(log.DefaultField().WithCaller(log.Internal).WithLevel(log.Panic).
			WithMessage("rpc server exit").WithExtra(err.Error()))
	}
}

func startHttpEngine(engine *gin.Engine, addr string) (exitChan chan error) {
	exit := make(chan error)
	go func() {
		exit <- engine.Run(addr)
	}()
	return exit
}

func startRpcEngine(engine *grpc.Server, conn net.Listener) (exitChan chan error) {
	exit := make(chan error)
	go func() {
		exit <- engine.Serve(conn)
	}()
	return exit
}
