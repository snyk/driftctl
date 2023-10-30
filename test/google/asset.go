package google

import (
	"context"
	"net"

	asset "cloud.google.com/go/asset/apiv1"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FakeAssetServer struct {
	SearchAllResourcesResults []*assetpb.ResourceSearchResult
	ListAssetsResults         []*assetpb.Asset
	err                       error
	assetpb.UnimplementedAssetServiceServer
}

func (s *FakeAssetServer) SearchAllResources(context.Context, *assetpb.SearchAllResourcesRequest) (*assetpb.SearchAllResourcesResponse, error) {
	return &assetpb.SearchAllResourcesResponse{Results: s.SearchAllResourcesResults}, s.err
}

func (s *FakeAssetServer) ListAssets(context.Context, *assetpb.ListAssetsRequest) (*assetpb.ListAssetsResponse, error) {
	return &assetpb.ListAssetsResponse{Assets: s.ListAssetsResults}, s.err
}

func NewFakeAssertServerWithList(listResults []*assetpb.Asset, err error) (*asset.Client, error) {
	return newAssetClient(&FakeAssetServer{ListAssetsResults: listResults, err: err})
}

func NewFakeAssetServer(searchResults []*assetpb.ResourceSearchResult, err error) (*asset.Client, error) {
	return newAssetClient(&FakeAssetServer{SearchAllResourcesResults: searchResults, err: err})
}

func newAssetClient(fakeServer *FakeAssetServer) (*asset.Client, error) {
	ctx := context.Background()
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
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}
