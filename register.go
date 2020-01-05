package gfly

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// register 注册服务
type register interface {
	// 注册grpc服务
	RegisterServer(*grpc.Server)
	// 注册gw服务
	RegisterHandlerFromEndpoint(context.Context,*runtime.ServeMux,string,[]grpc.DialOption)
}
