package gfly

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/fanxiaoping/gfly/util"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"net/http"
	"strings"
)
// apConfig Configuration information
type apConfig struct {
	//
	addr string
	//
	cer, key []byte
	//
	serverNameOverride  string
	// Cert is a self signed certificate
	cert tls.Certificate
	// CertPool contains the self signed certificate
	certPool *x509.CertPool
	//
	grpcServer *grpc.Server
	//
	dopts []grpc.DialOption
	//
	gwMux *runtime.ServeMux
	//
	httpMux *http.ServeMux
	//
	ctx context.Context
}

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

func SetConfigCert(cer, key []byte,port string){
	config = &apConfig{
		addr:       port,
		cer:        cer,
		key:        key,
	}
	config.initConfig()
}

func SetConfigCertOverride(cer, key []byte,port,serverNameOverride string){
	config = &apConfig{
		addr:       port,
		cer:        cer,
		key:        key,
		serverNameOverride:serverNameOverride,
	}
	config.initConfig()
}

func Register(r register){
	if config == nil{
		log.Fatalln("Not configured")
	}
	// 注册grpc服务
	r.RegisterServer(config.grpcServer)
	// 注册gw服务区
	r.RegisterHandlerFromEndpoint(config.ctx,config.gwMux,config.addr,config.dopts)
}
// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func Handle(pattern string, handler http.Handler){
	if config == nil{
		log.Fatalln("Not configured")
	}
	config.httpMux.Handle(pattern, handler)
}

func (a *apConfig) newServer(){
	conn, err := net.Listen("tcp", a.addr)
	if err != nil {
		log.Fatalln("TCP Listen err:%v\n", err)
	}
	_, cancel := context.WithCancel(a.ctx)
	defer cancel()

	srv := &http.Server{
		Addr:      a.addr,
		Handler:   a.grpcHandlerFunc(a.grpcServer, a.httpMux),
		TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{a.cert},
				NextProtos:   []string{http2.NextProtoTLS}, // HTTP2 TLS支持
			},
	}

	log.Printf("server listening:%s",a.addr)
	if err = srv.Serve(tls.NewListener(conn, srv.TLSConfig)); err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

//	grpcConfig	grpc 服务配置
func (a *apConfig) grpcConfig() error{
	var err error
	a.cert.Leaf, err = x509.ParseCertificate(a.cert.Certificate[0])
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewServerTLSFromCert(&a.cert)),
		grpc_middleware.WithUnaryServerChain(),
	}
	a.grpcServer = grpc.NewServer(opts...)
	return  nil
}

//	gatewayConfig geteway服务配置
func (a *apConfig)  gatewayConfig(){
	a.dopts = []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(a.certPool, a.serverNameOverride)),
	}
	a.gwMux = runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))

	a.httpMux = http.NewServeMux()
	a.httpMux.Handle("/", a.gwMux)
}

// initConfig
func (a *apConfig) initConfig() {
	var err error
	a.cert, err = tls.X509KeyPair(a.cer, a.key)
	if err != nil {
		log.Fatalln("Failed to parse key pair:", err)
	}
	a.cert.Leaf, err = x509.ParseCertificate(a.cert.Certificate[0])
	if err != nil {
		log.Fatalln("Failed to parse certificate:", err)
	}
	a.certPool = x509.NewCertPool()
	a.certPool.AddCert(a.cert.Leaf)
	a.ctx = context.Background()

	err = a.grpcConfig()
	if err != nil {
		log.Fatalln("Failed to parse grpcConfig:", err)
	}
	a.gatewayConfig()
}

//	grpcHandlerFunc	判断请求是grpc|http
func (a *apConfig) grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
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
			otherHandler.ServeHTTP(w, r)
		}
	})
}

