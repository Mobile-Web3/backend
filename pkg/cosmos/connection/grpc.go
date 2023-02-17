package connection

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	tendermint "github.com/tendermint/tendermint/rpc/client"
	googlerpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	protoCodec = encoding.GetCodec(proto.Name)

	ErrStreamingNotSupported = errors.New("streaming rpc not supported")
)

type GrpcClient struct {
	rpcClient         tendermint.ABCIClient
	interfaceRegistry types.InterfaceRegistry
	codec             codec.Codec
}

func NewGrpcClient(rpcClient tendermint.ABCIClient, interfaceRegistry types.InterfaceRegistry, codec codec.Codec) *GrpcClient {
	return &GrpcClient{
		rpcClient:         rpcClient,
		interfaceRegistry: interfaceRegistry,
		codec:             codec,
	}
}

func (c *GrpcClient) NewStream(context.Context, *googlerpc.StreamDesc, string, ...googlerpc.CallOption) (googlerpc.ClientStream, error) {
	return nil, ErrStreamingNotSupported
}

func sdkErrorToGRPCError(resp abci.ResponseQuery) error {
	switch resp.Code {
	case sdkerrors.ErrInvalidRequest.ABCICode():
		return status.Error(codes.InvalidArgument, resp.Log)
	case sdkerrors.ErrUnauthorized.ABCICode():
		return status.Error(codes.Unauthenticated, resp.Log)
	case sdkerrors.ErrKeyNotFound.ABCICode():
		return status.Error(codes.NotFound, resp.Log)
	default:
		return status.Error(codes.Unknown, resp.Log)
	}
}

func isQueryStoreWithProof(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}

	paths := strings.SplitN(path[1:], "/", 3)

	switch {
	case len(paths) != 3:
		return false
	case paths[0] != "store":
		return false
	case rootmulti.RequireProof("/" + paths[2]):
		return true
	}

	return false
}

func (c *GrpcClient) queryABCI(ctx context.Context, req abci.RequestQuery) (abci.ResponseQuery, error) {
	opts := tendermint.ABCIQueryOptions{
		Height: req.Height,
		Prove:  req.Prove,
	}

	result, err := c.rpcClient.ABCIQueryWithOptions(ctx, req.Path, req.Data, opts)
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	if !result.Response.IsOK() {
		err = fmt.Errorf("rpc error response; %s", sdkErrorToGRPCError(result.Response))
		return abci.ResponseQuery{}, err
	}

	if !opts.Prove || !isQueryStoreWithProof(req.Path) {
		return result.Response, nil
	}

	return result.Response, nil
}

func getHeightFromMetadata(md metadata.MD) (int64, error) {
	height := md.Get(grpctypes.GRPCBlockHeightHeader)
	if len(height) == 1 {
		return strconv.ParseInt(height[0], 10, 64)
	}
	return 0, nil
}

func getProveFromMetadata(md metadata.MD) (bool, error) {
	prove := md.Get("x-cosmos-query-prove")
	if len(prove) == 1 {
		return strconv.ParseBool(prove[0])
	}
	return false, nil
}

func (c *GrpcClient) runGRPCQuery(ctx context.Context, method string, req interface{}, md metadata.MD) (abci.ResponseQuery, metadata.MD, error) {
	reqBz, err := protoCodec.Marshal(req)
	if err != nil {
		err = fmt.Errorf("run grpc query; %s", err.Error())
		return abci.ResponseQuery{}, nil, err
	}

	// parse height header
	if heights := md.Get(grpctypes.GRPCBlockHeightHeader); len(heights) > 0 {
		height, err := strconv.ParseInt(heights[0], 10, 64)
		if err != nil {
			return abci.ResponseQuery{}, nil, err
		}
		if height < 0 {
			err = sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"client.Context.Invoke: height (%d) from %q must be >= 0", height, grpctypes.GRPCBlockHeightHeader)
			return abci.ResponseQuery{}, nil, err
		}
	}

	height, err := getHeightFromMetadata(md)
	if err != nil {
		return abci.ResponseQuery{}, nil, err
	}

	prove, err := getProveFromMetadata(md)
	if err != nil {
		return abci.ResponseQuery{}, nil, err
	}

	abciReq := abci.RequestQuery{
		Path:   method,
		Data:   reqBz,
		Height: height,
		Prove:  prove,
	}

	abciRes, err := c.queryABCI(ctx, abciReq)
	if err != nil {
		return abci.ResponseQuery{}, nil, err
	}

	// Create header metadata. For now the headers contain:
	// - block height
	// We then parse all the call options, if the call option is a
	// HeaderCallOption, then we manually set the value of that header to the
	// metadata.
	md = metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(abciRes.Height, 10))
	return abciRes, md, nil
}

func (c *GrpcClient) Invoke(ctx context.Context, method string, req, reply interface{}, opts ...googlerpc.CallOption) (err error) {
	inMd, _ := metadata.FromOutgoingContext(ctx)
	abciRes, outMd, err := c.runGRPCQuery(ctx, method, req, inMd)
	if err != nil {
		return err
	}

	if err = protoCodec.Unmarshal(abciRes.Value, reply); err != nil {
		err = fmt.Errorf("unmarshaling grpc reply; %s", err.Error())
		return err
	}

	for _, callOpt := range opts {
		header, ok := callOpt.(googlerpc.HeaderCallOption)
		if !ok {
			continue
		}

		*header.HeaderAddr = outMd
	}

	if c.interfaceRegistry != nil {
		if err = types.UnpackInterfaces(reply, c.codec); err != nil {
			err = fmt.Errorf("unpacking reply interfaces with codec; %s", err.Error())
			return err
		}

		return nil
	}

	return nil
}
