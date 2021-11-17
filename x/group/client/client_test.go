package client_test

import (
	"testing"

	"github.com/fetchai/fetchd/types/testutil/network"
	"github.com/fetchai/fetchd/x/group/client/testsuite"
	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	suite.Run(t, testsuite.NewIntegrationTestSuite(cfg))
}
