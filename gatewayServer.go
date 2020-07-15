package gfly

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strings"
)

// newServeMux 实例化runtime.ServeMux对象
//
// mpts runtime.ServeMuxOption可变数组
//
// runtime.ServeMux 实例化runtime.ServeMux指针对象
//
func newServeMux(mpts ...runtime.ServeMuxOption)*runtime.ServeMux{
	//json序列化字段空值返回
	mpts = append(mpts, runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	//header注入到MD中
	mpts = append(mpts, runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
		kv := []string{}
		for k, _ := range req.Header {
			v := req.Header.Get(k)
			if v == "" {
				continue
			}
			k = strings.ToLower(k)
			if k == "connection" || k == "content-length"{
				continue
			}
			kv = append(kv, k, v)
		}
		return metadata.Pairs(kv...)
	}))
	return runtime.NewServeMux(mpts...)
}