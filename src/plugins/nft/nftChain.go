/*
Copyright (C) 2019 by Martin Langlotz aka stackshadow

This file is part of gopilot, an rewrite of the copilot-project in go

gopilot is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, version 3 of this License

gopilot is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gopilot.  If not, see <http://www.gnu.org/licenses/>.
*/

package nft

import (
	"fmt"
	"os/exec"
)

type nftChain struct {
	table  *nftTable
	name   string
	Hook   string             `json:"hook"`
	Policy nftPolicy          `json:"policy"`
	Rules  map[string]*nftRule `json:"rules"`
}

func (table *nftTable) chainNew(chainName string, hookName string, policy nftPolicy) nftChain {

	// create a new chain
	var chain nftChain
	chain.table = table
	chain.name = chainName
	chain.Hook = hookName
	chain.Policy = policy
	chain.Rules = make(map[string]*nftRule)

	// add it to table
	table.Chains[chainName] = &chain

	return chain
}

func (chain *nftChain) Apply() error {

	// the command
	cmd := exec.Command(
		"sudo",
		"nft",
		"add",
		"chain", chain.table.Family.String(), chain.table.name, chain.name,
		"{ type filter hook "+chain.Hook+" priority 0 ; policy "+chain.Policy.String()+" ; }")
	logging.Info("nftChain.Add", fmt.Sprintf("%s", cmd.Args))

	// run it
	cmdErr := cmd.Run()
	if cmdErr != nil {
		logging.Error("nftChain.Add", fmt.Sprintf("%s", cmdErr))
		return cmdErr
	}

	return nil
}
