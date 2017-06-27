// Copyright 2017 The go-bazacoin Authors
// This file is part of go-bazacoin.
//
// go-bazacoin is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-bazacoin is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-bazacoin. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"

	"github.com/bazacoin/go-bazacoin/log"
)

// deployBzcstats queries the user for various input on deploying an bzcstats
// monitoring server, after which it executes it.
func (w *wizard) deployBzcstats() {
	// Select the server to interact with
	server := w.selectServer()
	if server == "" {
		return
	}
	client := w.servers[server]

	// Retrieve any active bzcstats configurations from the server
	infos, err := checkBzcstats(client, w.network)
	if err != nil {
		infos = &bzcstatsInfos{
			port:   80,
			host:   client.server,
			secret: "",
		}
	}
	// Figure out which port to listen on
	fmt.Println()
	fmt.Printf("Which port should bzcstats listen on? (default = %d)\n", infos.port)
	infos.port = w.readDefaultInt(infos.port)

	// Figure which virtual-host to deploy bzcstats on
	if infos.host, err = w.ensureVirtualHost(client, infos.port, infos.host); err != nil {
		log.Error("Failed to decide on bzcstats host", "err", err)
		return
	}
	// Port and proxy settings retrieved, figure out the secret and boot bzcstats
	fmt.Println()
	if infos.secret == "" {
		fmt.Printf("What should be the secret password for the API? (must not be empty)\n")
		infos.secret = w.readString()
	} else {
		fmt.Printf("What should be the secret password for the API? (default = %s)\n", infos.secret)
		infos.secret = w.readDefaultString(infos.secret)
	}
	// Try to deploy the bzcstats server on the host
	trusted := make([]string, 0, len(w.servers))
	for _, client := range w.servers {
		if client != nil {
			trusted = append(trusted, client.address)
		}
	}
	if out, err := deployBzcstats(client, w.network, infos.port, infos.secret, infos.host, trusted); err != nil {
		log.Error("Failed to deploy bzcstats container", "err", err)
		if len(out) > 0 {
			fmt.Printf("%s\n", out)
		}
		return
	}
	// All ok, run a network scan to pick any changes up
	w.networkStats(false)
}
