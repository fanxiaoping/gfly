package gfly

import (
	"log"
	"net/http"
)

var (
	config *apConfig
)

// Run
func Run() {
	if config == nil{
		log.Fatalln("Not configured")
	}
	config.newServer()
}

// SetConfigCert 初始化配置
func SetConfigCert(cer, key []byte,port string){
	config = &apConfig{
		addr:       port,
		cer:        cer,
		key:        key,
	}
	config.initConfig()
}
// SetConfigCertOverride 初始化配置，设置证书测试域
func SetConfigCertOverride(cer, key []byte,port,serverNameOverride string){
	config = &apConfig{
		addr:       port,
		cer:        cer,
		key:        key,
		serverNameOverride:serverNameOverride,
	}
	config.initConfig()
}

// Register 注册grpc服务 && 注册gw服务
func Register(r register){
	if config == nil{
		log.Fatalln("Not configured")
	}
	// 注册grpc服务
	r.RegisterServer(config.grpcServer)
	// 注册gw服务
	r.RegisterHandlerFromEndpoint(config.ctx,config.gwMux,config.addr,config.dopts)
}

// Handle 自定义http路由器处理，支持fromdata
func Handle(pattern string, handler http.Handler){
	if config == nil{
		log.Fatalln("Not configured")
	}
	config.httpMux.Handle(pattern, handler)
}

// HandleFilter http过滤器
func HandleFilter(filter func(w http.ResponseWriter, r *http.Request)bool){
	if config == nil{
		log.Fatalln("Not configured")
	}
	config.gwFilter = filter
}

