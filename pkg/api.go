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

	sd "github.com/cerc-io/plugeth-statediff"
)

// APIName is the namespace used for the state diffing service API
const APIName = "statediff"

// APIVersion is the version of the state diffing service API
const APIVersion = "0.0.1"

// PublicStateDiffAPI provides an RPC interface
// that can be used to fetch historical diffs from LevelDB directly
type PublicStateDiffAPI struct {
	sds *Service
}

// NewPublicStateDiffAPI creates an rpc interface for the underlying statediff service
func NewPublicStateDiffAPI(sds *Service) *PublicStateDiffAPI {
	return &PublicStateDiffAPI{
		sds: sds,
	}
}

// StateDiffAt returns a state diff payload at the specific blockheight
func (api *PublicStateDiffAPI) StateDiffAt(ctx context.Context, blockNumber uint64, params sd.Params) (*sd.Payload, error) {
	return api.sds.StateDiffAt(blockNumber, params)
}

// WriteStateDiffAt writes a state diff object directly to DB at the specific blockheight
func (api *PublicStateDiffAPI) WriteStateDiffAt(ctx context.Context, blockNumber uint64, params sd.Params) error {
	return api.sds.WriteStateDiffAt(blockNumber, params)
}

// WriteStateDiffsInRange writes the state diff objects for the provided block range, with the provided params
func (api *PublicStateDiffAPI) WriteStateDiffsInRange(ctx context.Context, start, stop uint64, params sd.Params) error {
	return api.sds.WriteStateDiffsInRange(start, stop, params)
}
