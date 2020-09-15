// Copyright © 2020 Vulcanize, Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package statediff

import (
	"context"
	sd "github.com/ethereum/go-ethereum/statediff"
)

// APIName is the namespace used for the state diffing service API
const APIName = "statediff"

// APIVersion is the version of the state diffing service API
const APIVersion = "0.0.1"

// PublicStateDiffAPI provides an RPC interface
// that can be used to fetch historical diffs from leveldb directly
type PublicStateDiffAPI struct {
	sds IService
}

// NewPublicStateDiffAPI creates an rpc interface for the underlying statediff service
func NewPublicStateDiffAPI(sds IService) *PublicStateDiffAPI {
	return &PublicStateDiffAPI{
		sds: sds,
	}
}

// StateDiffAt returns a state diff payload at the specific blockheight
func (api *PublicStateDiffAPI) StateDiffAt(ctx context.Context, blockNumber uint64, params sd.Params) (*sd.Payload, error) {
	return api.sds.StateDiffAt(blockNumber, params)
}

// StateTrieAt returns a state trie payload at the specific blockheight
func (api *PublicStateDiffAPI) StateTrieAt(ctx context.Context, blockNumber uint64, params sd.Params) (*sd.Payload, error) {
	return api.sds.StateTrieAt(blockNumber, params)
}
