// VulcanizeDB
// Copyright © 2020 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package rpc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"

	"github.com/cerc-io/eth-statediff-service/pkg/prom"
)

// StartHTTPEndpoint starts the HTTP RPC endpoint, configured with cors/vhosts/modules.
func StartHTTPEndpoint(endpoint string, apis []rpc.API, modules []string, cors []string, vhosts []string, timeouts rpc.HTTPTimeouts) (*rpc.Server, error) {

	srv := rpc.NewServer()
	err := node.RegisterApis(apis, modules, srv)
	if err != nil {
		utils.Fatalf("Could not register HTTP API: %w", err)
	}
	handler := node.NewHTTPHandlerStack(srv, cors, vhosts, nil)

	// start http server
	_, addr, err := node.StartHTTPEndpoint(endpoint, rpc.DefaultHTTPTimeouts, prom.HTTPMiddleware(handler))
	if err != nil {
		utils.Fatalf("Could not start RPC api: %v", err)
	}
	extapiURL := fmt.Sprintf("http://%v/", addr)
	log.Infof("HTTP endpoint opened %s", extapiURL)

	return srv, err
}
