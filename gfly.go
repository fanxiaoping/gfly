package gfly

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/fanxiaoping/gfly/util"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc"
)

type Fly struct {
	// 地址(127.0.0.1:8080 || :8080)
	addr string
	// Cert is a self signed certificate
	cert tls.Certificate
	//
	ctx context.Context
	//
	mpts []runtime.ServeMuxOption
	//
	opts []grpc.ServerOption
	//
	dopts []grpc.DialOption
	//
	httpMux *http.ServeMux
	//
	register register
}

func NewFlyCert(cer, key []byte, addr string) (fly *Fly, err error) {
	return NewFlyCertOverride(cer, key, addr, "")
}

func NewFlyCertOverride(cer, key []byte, addr, serverNameOverride string) (fly *Fly, err error) {
	fly = &Fly{
		addr:    addr,
		httpMux: http.NewServeMux(),
		ctx:     context.Background(),
	}
	fly.cert, err = tls.X509KeyPair(cer, key)
	if err != nil {
		return nil, err
	}

	if serverNameOverride == "" {
		fly.dopts = append(fly.dopts, grpc.WithTransportCredentials(credentials.NewServerTLSFromCert(&fly.cert)))
	} else {
		fly.cert.Leaf, err = x509.ParseCertificate(fly.cert.Certificate[0])
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		certPool.AddCert(fly.cert.Leaf)
		fly.dopts = append(fly.dopts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(certPool, serverNameOverride)))
	}
	return fly, nil
}

func (_self *Fly) WithServerOption(option ...grpc.ServerOption) {
	_self.opts = append(_self.opts, option...)
}

func (_self *Fly) WithServeMuxOption(option ...runtime.ServeMuxOption) {
	_self.mpts = append(_self.mpts, option...)
}

func (_self *Fly) WithDialOption(option ...grpc.DialOption) {
	_self.dopts = append(_self.dopts, option...)
}

// Register 注册grpc服务 && 注册gw服务
func (_self *Fly) Register(r register) {
	_self.register = r
	//// 注册grpc服务
	//r.RegisterServer(config.grpcServer)
	//for _, item := range option {
	//	config.dopts = append(config.dopts, item)
	//}
	//// 注册gw服务
	//r.RegisterHandlerFromEndpoint(config.ctx, config.gwMux, config.addr, config.dopts)
}

// Handle 自定义http路由器处理，支持fromdata
func (_self *Fly) Handle(pattern string, handler http.Handler) {
	_self.httpMux.Handle(pattern, handler)
}

// Run
func (_self *Fly) Run() error {
	gatewayServer := newServeMux(_self.mpts...)
	_self.httpMux.Handle("/", gatewayServer)

	grpcServer, err := newServer(_self.cert, _self.opts...)
	if err != nil {
		return err
	}
	if _self.register != nil {
		_self.register.RegisterServer(grpcServer)
		_self.register.RegisterHandlerFromEndpoint(_self.ctx, gatewayServer, _self.addr, _self.dopts)
	}

	conn, err := net.Listen("tcp", _self.addr)
	if err != nil {
		return err
	}
	_, cancel := context.WithCancel(_self.ctx)
	defer cancel()

	srv := &http.Server{
		Addr:    _self.addr,
		Handler: _self.grpcHandlerFunc(grpcServer, _self.httpMux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{_self.cert},
			NextProtos:   []string{http2.NextProtoTLS}, // HTTP2 TLS支持
		},
	}

	log.Printf("server listening:%s", _self.addr)
	return srv.Serve(tls.NewListener(conn, srv.TLSConfig))
}

//	grpcHandlerFunc	判断请求是grpc|http
func (a *Fly) grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			//http
			if origin := r.Header.Get("Origin"); origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
					util.PreflightHandler(w, r)
					return
				}
			}
			// 自定义过滤器
			//if config.gwFilter != nil {
			//	res := config.gwFilter(w, r)
			//	if res == false {
			//		return
			//	}
			//}
			otherHandler.ServeHTTP(w, r)
		}
	})
}

//var (
//	config *apConfig
//)

// SetConfigCert 初始化配置
//func SetConfigCert(cer, key []byte, port string, interceptors ...grpc.UnaryServerInterceptor) {
//	config = &apConfig{
//		addr:         port,
//		cer:          cer,
//		key:          key,
//		interceptors: interceptors,
//	}
//	config.initConfig()
//}
//
//// SetConfigCertOverride 初始化配置，设置证书测试域
//func SetConfigCertOverride(cer, key []byte, port, serverNameOverride string, interceptors ...grpc.UnaryServerInterceptor) {
//	config = &apConfig{
//		addr:               port,
//		cer:                cer,
//		key:                key,
//		serverNameOverride: serverNameOverride,
//		interceptors:       interceptors,
//	}
//	config.initConfig()
//}
//
//
//
//// HandleFilter http过滤器
//func HandleFilter(filter func(w http.ResponseWriter, r *http.Request) bool) {
//	if config == nil {
//		log.Fatalln("Not configured")
//	}
//	config.gwFilter = filter
//}
