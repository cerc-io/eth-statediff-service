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

package prom

import (
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

// HTTPMiddleware http connection metric reader
func HTTPMiddleware(next http.Handler) http.Handler {
	if !metrics {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCount.Inc()

		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Now().Sub(start)
		httpDuration.Observe(float64(duration.Seconds()))
	})
}

// IPCMiddleware unix-socket connection counter
func IPCMiddleware(server *rpc.Server, client rpc.Conn) {
	if metrics {
		ipcCount.Inc()
	}
	server.ServeCodec(rpc.NewCodec(client), 0)
	if metrics {
		ipcCount.Dec()
	}
}
