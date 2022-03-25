package consul

import (
	"google.golang.org/grpc/resolver"
)

type resolverBuilder struct{}

const scheme = "consul"

func NewBuilder() resolver.Builder {
	return &resolverBuilder{}
}

func (*resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	// glog.Info(target.URL.Scheme, target.URL.Host, target.URL.Path)
	r, err := newConsulResolver(cc, target.URL.Scheme, target.URL.Host, target.URL.Path)
	if err != nil {
		return nil, err
	}
	r.start()
	return r, nil
}

func (*resolverBuilder) Scheme() string {
	return scheme
}
