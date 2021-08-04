package google

import (
	"context"
	"net"

	asset "cloud.google.com/go/asset/apiv1"
	"google.golang.org/api/option"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
	"google.golang.org/grpc"
)

type FakeAssetServer struct {
	SearchAllResourcesResults []*assetpb.ResourceSearchResult
	err                       error
	assetpb.UnimplementedAssetServiceServer
}

func (s *FakeAssetServer) SearchAllResources(context.Context, *assetpb.SearchAllResourcesRequest) (*assetpb.SearchAllResourcesResponse, error) {
	return &assetpb.SearchAllResourcesResponse{Results: s.SearchAllResourcesResults}, s.err
}

func NewFakeAssetServer(results []*assetpb.ResourceSearchResult, err error) (*asset.Client, error) {
	ctx := context.Background()
	fakeServer := &FakeAssetServer{SearchAllResourcesResults: results, err: err}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	gsrv := grpc.NewServer()
	assetpb.RegisterAssetServiceServer(gsrv, fakeServer)
	fakeServerAddr := l.Addr().String()
	go func() {
		if err := gsrv.Serve(l); err != nil {
			panic(err)
		}
	}()
	// Create a client.
	client, err := asset.NewClient(ctx,
		option.WithEndpoint(fakeServerAddr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}
