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

type nftRule struct {
	chain     *nftChain
	index     int
	Enabled   bool       `json:"enabled"`
	Policy    nftPolicy  `json:"policy"`
	Statement [][]string `json:"statements"`
}

func (chain *nftChain) ruleNew(policy nftPolicy) *nftRule {

	// new rule
	var rule nftRule
	rule.chain = chain
	rule.index = len(chain.Rules)
	rule.Enabled = false
	rule.Policy = policy

	// add rule to chain
	chain.Rules = append(chain.Rules, &rule)

	return &rule
}

func (rule *nftRule) statementAdd(statement []string) {

	// append statement to array
	rule.Statement = append(rule.Statement, statement)

}

func (rule *nftRule) Apply() error {

	// fmt.Printf("%+v\n", rule)

	args := []string{
		"sudo",
		"nft",
		"add",
		"rule",
		rule.chain.table.Family.String(),
		rule.chain.table.name,
		rule.chain.name,
	}

	for _, statement := range rule.Statement {
		args = append(args, statement...)
	}
	args = append(args, rule.Policy.String())

	// if not enable, return
	if rule.Enabled == false {
		logging.Info("nftRule.Add", fmt.Sprintf("Skip rule %s", args))
		return nil
	}

	// the command
	cmd := exec.Command("sudo")
	cmd.Args = args
	logging.Info("nftRule.Add", fmt.Sprintf("%s", cmd.Args))

	// run it
	cmdErr := cmd.Run()
	if cmdErr != nil {
		logging.Error("nftRule.Add", fmt.Sprintf("%s", cmdErr))
		return cmdErr
	}

	return nil

}
