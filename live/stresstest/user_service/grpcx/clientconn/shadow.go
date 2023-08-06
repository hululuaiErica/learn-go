package clientconn

import (
	"context"
	"google.golang.org/grpc"
)

type ShadowClientConn struct {
	liveCC   grpc.ClientConnInterface
	shadowCC grpc.ClientConnInterface
}

func NewShadowClientConn(liveCC grpc.ClientConnInterface, shadowCC grpc.ClientConnInterface) *ShadowClientConn {
	return &ShadowClientConn{
		liveCC:   liveCC,
		shadowCC: shadowCC,
	}
}

func (s *ShadowClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	if s.isShadow(ctx) {
		return s.shadowCC.Invoke(ctx, method, args, reply, opts...)
	}
	return s.liveCC.Invoke(ctx, method, args, reply, opts...)
}

func (s *ShadowClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if s.isShadow(ctx) {
		return s.shadowCC.NewStream(ctx, desc, method, opts...)
	}
	return s.liveCC.NewStream(ctx, desc, method, opts...)
}

func (s *ShadowClientConn) isShadow(ctx context.Context) bool {
	return ctx.Value("stress-test") == "true"
}
