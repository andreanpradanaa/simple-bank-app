package gapi

import (
	"context"
	"log"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	GrpcgatewayUserAgent = "grpcgateway-user-agent"
	UserAgent            = "user-agent"
	XForwardedHost       = "x-forwarded-host"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {

	mtdt := Metadata{}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Println("metadata: ", md)
		if userAgent := md.Get("grpcgateway-user-agent"); len(userAgent) > 0 {
			mtdt.UserAgent = userAgent[0]
		}

		if clientIP := md.Get("x-forwarded-host"); len(clientIP) > 0 {
			mtdt.ClientIP = clientIP[0]
		}

		if userAgent := md.Get("user-agent"); len(userAgent) > 0 {
			mtdt.UserAgent = userAgent[0]
		}
	}

	clientIP, ok := peer.FromContext(ctx)
	if ok {
		mtdt.ClientIP = clientIP.Addr.String()
	}

	return &mtdt
}
