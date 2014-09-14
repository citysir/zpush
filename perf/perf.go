// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

package perf

import (
	log "code.google.com/p/log4go"
	"net/http"
	"net/http/pprof"
)

// StartPprof start http pprof.
func BindAddr(pprofAddr string) {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/pprof/symbol", pprof.Symbol)
	go func() {
		if err := http.ListenAndServe(pprofAddr, pprofServeMux); err != nil {
			log.Error("http.ListenAndServe(\"%s\", pPprofBindprofServeMux) error(%v)", pprofAddr, err)
			panic(err)
		}
	}()
}
