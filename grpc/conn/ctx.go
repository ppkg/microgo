package conn

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func GetCtx(ctx ...interface{}) (context.Context, context.CancelFunc) {
	requestId := uuid.New().String()
	for _, c := range ctx {
		switch v := c.(type) {
		case *gin.Context:
			requestId = v.GetHeader("X-Request-Id")
		case context.Context:
			return metadata.NewOutgoingContext(v, getMD(v)), nil
		default:
			panic("ctx type error")
		}
	}

	cc, cf := context.WithTimeout(context.Background(), time.Second*60)
	cc = metadata.AppendToOutgoingContext(cc, "X-Request-Id", requestId)
	return cc, cf
}

func getMD(c context.Context) metadata.MD {
	if md, ok := metadata.FromIncomingContext(c); ok {
		return md
	} else if md, ok := metadata.FromOutgoingContext(c); ok {
		return md
	}
	return make(metadata.MD)
}
