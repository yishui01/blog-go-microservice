package article_service_v1

import (
	"google.golang.org/grpc"
)

const ServerAddr2 = "127.0.0.1:9000"

// NewClient new grpc client
func NewClient(opts ...grpc.DialOption) (ArticleClient, error) {

	opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())
	cc, err := grpc.Dial(ServerAddr2, opts...)
	if err != nil {
		return nil, err
	}
	return NewArticleClient(cc), nil
}
