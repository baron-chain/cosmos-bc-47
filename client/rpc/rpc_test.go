package rpc_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	abci "github.com/baron-chain/cometbft-bc/abci/types"
	"github.com/baron-chain/cosmos-bc-47/client/rpc"
	"github.com/baron-chain/cosmos-bc-47/types/grpc"
	"github.com/baron-chain/cosmos-bc-47/x/bank/types"
	clitestutil "github.com/baron-chain/cosmos-bc-47/testutil/cli"
	"github.com/baron-chain/cosmos-bc-47/testutil/network"
	"github.com/baron-chain/cosmos-bc-47/testutil/testdata"
)

const (
	minBlockHeight = 1
	testMessage    = "hello"
)

type IntegrationTestSuite struct {
	suite.Suite
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	cfg, err := network.DefaultConfigWithAppConfig(network.MinimumAppConfig())
	s.NoError(err)

	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestStatusCommand() {
	validator := s.network.Validators[0]
	cmd := rpc.StatusCommand()

	out, err := clitestutil.ExecTestCLICmd(validator.ClientCtx, cmd, []string{})
	s.Require().NoError(err)
	s.Require().Contains(out.String(), fmt.Sprintf("\"moniker\":\"%s\"", validator.Moniker))
}

func (s *IntegrationTestSuite) TestGRPCQuery() {
	var header metadata.MD
	validator := s.network.Validators[0]
	testClient := testdata.NewQueryClient(validator.ClientCtx)

	res, err := testClient.Echo(
		context.Background(),
		&testdata.EchoRequest{Message: testMessage},
		grpc.Header(&header),
	)
	s.NoError(err)

	blockHeight, err := s.getBlockHeightFromHeader(header)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(blockHeight, minBlockHeight)
	s.Equal(testMessage, res.Message)
}

func (s *IntegrationTestSuite) getBlockHeightFromHeader(header metadata.MD) (int, error) {
	heightStr := header.Get(grpctypes.GRPCBlockHeightHeader)
	if len(heightStr) == 0 {
		return 0, fmt.Errorf("no block height in header")
	}
	return strconv.Atoi(heightStr[0])
}

func (s *IntegrationTestSuite) TestQueryABCIHeight() {
	testCases := []struct {
		name      string
		reqHeight int64
		ctxHeight int64
		expHeight int64
	}{
		{
			name:      "specified request height",
			reqHeight: 3,
			ctxHeight: 1,
			expHeight: 3,
		},
		{
			name:      "use context height when request height is zero",
			reqHeight: 0,
			ctxHeight: 3,
			expHeight: 3,
		},
		{
			name:      "use latest height when both heights are zero",
			reqHeight: 0,
			ctxHeight: 0,
			expHeight: 4,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.network.WaitForHeight(tc.expHeight)
			validator := s.network.Validators[0]
			
			clientCtx := validator.ClientCtx.WithHeight(tc.ctxHeight)
			req := abci.RequestQuery{
				Path:   fmt.Sprintf("store/%s/key", banktypes.StoreKey),
				Height: tc.reqHeight,
				Data:   banktypes.CreateAccountBalancesPrefix(validator.Address),
				Prove:  true,
			}

			res, err := clientCtx.QueryABCI(req)
			s.Require().NoError(err)
			s.Require().Equal(tc.expHeight, res.Height)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
