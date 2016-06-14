package libcontainerd

import (
	"io"

	"github.com/docker/containerd/api/grpc/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (b *bridge) CreateContainer(ctx context.Context, in *types.CreateContainerRequest, opts ...grpc.CallOption) (*types.CreateContainerResponse, error) {
	return b.s.CreateContainer(ctx, in)
}

func (b *bridge) UpdateContainer(ctx context.Context, in *types.UpdateContainerRequest, opts ...grpc.CallOption) (*types.UpdateContainerResponse, error) {
	return b.s.UpdateContainer(ctx, in)
}

func (b *bridge) Signal(ctx context.Context, in *types.SignalRequest, opts ...grpc.CallOption) (*types.SignalResponse, error) {
	return b.s.Signal(ctx, in)
}

func (b *bridge) UpdateProcess(ctx context.Context, in *types.UpdateProcessRequest, opts ...grpc.CallOption) (*types.UpdateProcessResponse, error) {
	return b.s.UpdateProcess(ctx, in)
}

func (b *bridge) AddProcess(ctx context.Context, in *types.AddProcessRequest, opts ...grpc.CallOption) (*types.AddProcessResponse, error) {
	return b.s.AddProcess(ctx, in)
}

func (b *bridge) CreateCheckpoint(ctx context.Context, in *types.CreateCheckpointRequest, opts ...grpc.CallOption) (*types.CreateCheckpointResponse, error) {
	return b.s.CreateCheckpoint(ctx, in)
}

func (b *bridge) DeleteCheckpoint(ctx context.Context, in *types.DeleteCheckpointRequest, opts ...grpc.CallOption) (*types.DeleteCheckpointResponse, error) {
	return b.s.DeleteCheckpoint(ctx, in)
}

func (b *bridge) ListCheckpoint(ctx context.Context, in *types.ListCheckpointRequest, opts ...grpc.CallOption) (*types.ListCheckpointResponse, error) {
	return b.s.ListCheckpoint(ctx, in)
}

func (b *bridge) State(ctx context.Context, in *types.StateRequest, opts ...grpc.CallOption) (*types.StateResponse, error) {
	return b.s.State(ctx, in)
}

func (b *bridge) GetServerVersion(ctx context.Context, in *types.GetServerVersionRequest, opts ...grpc.CallOption) (*types.GetServerVersionResponse, error) {
	return b.s.GetServerVersion(ctx, in)
}

func (b *bridge) Events(ctx context.Context, in *types.EventsRequest, opts ...grpc.CallOption) (types.API_EventsClient, error) {
	c := make(chan *types.Event, 1024)
	client := &aPI_EventsClient{
		c: c,
	}
	client.ctx = ctx
	server := &aPI_EventsServer{
		c: c,
	}
	server.ctx = ctx
	go b.s.Events(in, server)
	return client, nil
}

func (b *bridge) Stats(ctx context.Context, in *types.StatsRequest, opts ...grpc.CallOption) (*types.StatsResponse, error) {
	return b.s.Stats(ctx, in)
}

type aPI_EventsServer struct {
	serverStream
	c chan *types.Event
}

func (a *aPI_EventsServer) Send(msg *types.Event) error {
	a.c <- msg
	return nil
}

type serverStream struct {
	stream
}

func (s *serverStream) SendHeader(metadata.MD) error {
	return nil
}

func (s *serverStream) SetTrailer(metadata.MD) {
}

type aPI_EventsClient struct {
	clientStream
	c chan *types.Event
}

func (a *aPI_EventsClient) Recv() (*types.Event, error) {
	e, ok := <-a.c
	if !ok {
		return nil, io.EOF
	}
	return e, nil
}

type stream struct {
	ctx context.Context
}

func (s *stream) Context() context.Context {
	return s.ctx
}
func (s *stream) SendMsg(m interface{}) error {
	return nil
}
func (s *stream) RecvMsg(m interface{}) error {
	return nil
}

// ClientStream defines the interface a client stream has to satify.
type clientStream struct {
	stream
}

func (c *clientStream) Header() (metadata.MD, error) {
	return nil, nil
}
func (c *clientStream) Trailer() metadata.MD {
	return nil
}
func (c *clientStream) CloseSend() error {
	return nil
}
