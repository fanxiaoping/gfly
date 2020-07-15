package gfly

import (
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// newServer 实例化grpc.Server
//
// cert cer证书
// opts grpc.ServerOption可变数组
//
// grpc.Server 实例化grpc.Server指针对象
// error    如果没有错误，则为nil，否则为错误对象
//
func newServer(cert tls.Certificate,opts ...grpc.ServerOption)(*grpc.Server,error){
	var err error
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil,err
	}
	opts = append(opts,grpc.Creds(credentials.NewServerTLSFromCert(&cert)))

	return grpc.NewServer(opts...),nil
}




