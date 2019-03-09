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

package pluginNFT

import (
	"core/clog"
	"core/config"
	"core/msgbus"
	"encoding/json"
	"flag"
	"fmt"
	"time"
)

var useSudo = true

var logging clog.Logger
var plugin msgbus.Plugin

type nftJSONConfig struct {
	Tables map[string]*nftTable `json:"tables"`
}

type nftJSONChain struct {
	Name      string    `json:"name"`
	Hook      string    `json:"hook"`
	Policy    nftPolicy `json:"policy"`
	RuleCount float64   `json:"rulecount"`
}

type nftJSONRule struct {
	ChainName string `json:"chainName"`
	// The index start with 1, because 0 means the rule is new
	Position  int        `json:"position"`
	Enabled   bool       `json:"enabled"`
	Policy    nftPolicy  `json:"policy"`
	Statement [][]string `json:"statements"`
}

var nftConfig nftJSONConfig

var nftSkipApplyRules bool
var applyTimer *time.Timer

/*
ParseCmdLine read the cmd-line arguments and set global values from it
*/
func ParseCmdLine() {
	flag.BoolVar(&nftSkipApplyRules, "nftSkipOnStart", false, "If true, dont apply rules on application start")
}

/*
Init the nft-modules
*/
func Init() {

	logging = clog.New("NFT")

	// load from config
	nftConfig, _ = loadFromConfig()

	if nftSkipApplyRules == false {
		nftConfig.applyAll()
	}

	/*
		chain := nftConfig.Tables["tGopilot"].Chains["input"]
		rule := chain.ruleNew(nftPolicyAccept)
		rule.statementAdd([]string{"ct", "state", "related,established"})
		nftConfig.saveConfig()
	*/
	/*
		table := tableNew("tGopliot", nftFAMILYINET)
		table.Apply()

		chain := table.chainNew("Input", "input", nftPolicyAccept)
		chain.Apply()
		rule := chain.ruleNew(nftPolicyAccept)
		rule.statementAdd([]string{"iif", "lo"})
		rule.Apply()

		rule = chain.ruleNew(nftPolicyAccept)
		rule.statementAdd([]string{"ct", "state", "related,established"})
		rule.Apply()

		rule = chain.ruleNew(nftPolicyDrop)
		rule.statementAdd([]string{"ct", "state", "invalid"})
		rule.Apply()

		rule = chain.ruleNew(nftPolicyAccept)
		rule.statementAdd([]string{"ip", "protocol", "icmp"})
		rule.statementAdd([]string{"icmp", "type", "echo-request"})
		rule.statementAdd([]string{"ct", "state", "new"})
		rule.Apply()

		chain = table.chainNew("Forward", "forward", nftPolicyAccept)
		chain.Apply()

		chain = table.chainNew("Output", "output", nftPolicyAccept)
		chain.Apply()
	*/
	// register plugin on messagebus
	plugin = msgbus.NewPlugin("NFT")
	plugin.Register()
	plugin.ListenForGroup("nft", onMessage)
}

func loadFromConfig() (nftJSONConfig, error) {

	needToSave := false

	// get the config
	var jsonObject map[string]interface{}
	jsonObject, _ = config.GetJSONObject("nft")
	if jsonObject == nil {
		jsonObject = make(map[string]interface{})
		needToSave = true
	}

	// the final struct
	var jsonConfig nftJSONConfig

	// interface -> string
	jsonString, err := json.Marshal(jsonObject)
	if err != nil {
		logging.Error("Marshal", err.Error())
		return jsonConfig, err
	}

	// string -> struct
	err = json.Unmarshal([]byte(jsonString), &jsonConfig)
	if err != nil {
		logging.Error("Unmarshal", err.Error())
		return jsonConfig, err
	}

	if len(jsonConfig.Tables) == 0 {
		jsonConfig.Tables = make(map[string]*nftTable)
		needToSave = true
	}

	// table
	var newTable nftTable
	if _, ok := jsonConfig.Tables["tGopilot"]; !ok {

		newTable = tableNew("tGopilot", nftFAMILYINET)
		jsonConfig.Tables["tGopilot"] = &newTable
		needToSave = true
	}
	table := jsonConfig.Tables["tGopilot"]

	// chain - input
	if _, ok := table.Chains["input"]; !ok {
		table.chainNew("input", "input", nftPolicyAccept)
		needToSave = true
	}

	// chain - forward
	if _, ok := table.Chains["forward"]; !ok {
		table.chainNew("forward", "forward", nftPolicyAccept)
		needToSave = true
	}

	// chain - output
	if _, ok := table.Chains["output"]; !ok {
		table.chainNew("output", "output", nftPolicyAccept)
		needToSave = true
	}

	// save it ?
	if needToSave == true {
		jsonConfig.saveConfig()
	}

	// because the name and ids are keys inside the json, we need to at this missing infos in the structs
	for tableName, table := range jsonConfig.Tables {

		// fill out values which are not in the json
		table.name = tableName

		for chainName, chain := range table.Chains {

			// fill out values which are not in the json
			chain.table = table
			chain.name = chainName

			for ruleIndex, rule := range chain.Rules {

				// fill out values which are not in the json
				rule.chain = chain
				rule.index = ruleIndex

			}
		}
	}

	return jsonConfig, nil
}

func (nftconfig *nftJSONConfig) saveConfig() error {

	// to json
	groupObjectBytes, err := json.Marshal(nftconfig)
	if err != nil {
		logging.Error("Marshal", err.Error())
		return err
	}

	// from json
	var jsonObject map[string]interface{}
	err = json.Unmarshal(groupObjectBytes, &jsonObject)
	if err != nil {
		logging.Error("Unmarshal", err.Error())
		return err
	}

	// save it
	config.SetJSONObject("nft", jsonObject)
	config.Save()

	return nil
}

func (config *nftJSONConfig) applyAll() error {

	// flush
	RulesetFlush()

	var err error

	for _, table := range config.Tables {

		// Add Table
		err = table.Apply()
		if err != nil {
			return err
		}

		for _, chain := range table.Chains {

			// Add chain
			err = chain.Apply()
			if err != nil {
				return err
			}

			for _, rule := range chain.Rules {

				// Add rule
				err = rule.Apply()
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

func onMessage(message *msgbus.Msg, group, command, payload string) {

	// If the timer is not finished, we can confirm from the ui
	// which means that we dont kick out ourselfe :)
	// also we can save the rules
	if command == "apply" {
		err := nftConfig.applyAll()
		if err != nil {
			message.Answer(&plugin, "error", fmt.Sprintf("%s", err))
			return
		}

		applyTimer = time.AfterFunc(time.Second*20, func() {
			logging.Error("apply", "No confirm after 20 seconds, load last confirmed rules")

			// load saved rules
			nftConfig, _ = loadFromConfig()
			nftConfig.applyAll()

			applyTimer = nil
			message.Answer(&plugin, "confirmOk", "")
			return
		})

		message.Answer(&plugin, "confirmWait", "")
		return
	}

	if command == "confirm" {
		if applyTimer != nil {
			applyTimer.Stop()
			nftConfig.saveConfig()
		}

		applyTimer = nil
		message.Answer(&plugin, "confirmOk", "")
		return
	}

	if command == "confirmCancel" {

		// stop timer
		if applyTimer != nil {
			applyTimer.Stop()
		}

		// reload and apply from config
		nftConfig, _ = loadFromConfig()
		nftConfig.applyAll()

		applyTimer = nil
		message.Answer(&plugin, "confirmCancelOk", "")
		return
	}

	// from here, the timer should not be active
	if applyTimer != nil {
		logging.Error("applyTimer", "Timer is active, you must confirm your change, or wait until the timer is finished")
		message.Answer(&plugin, "error", "Timer is active, you must confirm your change, or wait until the timer is finished")
		return
	}

	if command == "getChains" {

		if table, ok := nftConfig.Tables["tGopilot"]; ok {

			for chainName, chain := range table.Chains {

				// copy it locally to remove rules
				var jsonChain nftJSONChain
				jsonChain.Name = chainName
				jsonChain.Hook = chain.Hook
				jsonChain.Policy = chain.Policy
				jsonChain.RuleCount = float64(len(chain.Rules))

				groupObjectBytes, err := json.Marshal(jsonChain)
				if err != nil {
					logging.Error("getChains", fmt.Sprintf("%s", err))
					continue
				}

				message.Answer(&plugin, "chain", string(groupObjectBytes))
			}
			return
		}

		message.Answer(&plugin, "error", fmt.Sprintf("Table '%s' dont exist", payload))
		return
	}

	if command == "getRules" {

		for _, table := range nftConfig.Tables {

			if chain, ok := table.Chains[payload]; ok {

				for ruleIndex, rule := range chain.Rules {

					var jsonRule nftJSONRule
					jsonRule.ChainName = rule.chain.name
					jsonRule.Position = ruleIndex + 1
					jsonRule.Enabled = rule.Enabled
					jsonRule.Policy = rule.Policy
					jsonRule.Statement = rule.Statement

					groupObjectBytes, err := json.Marshal(jsonRule)
					if err != nil {
						logging.Error("getRules", fmt.Sprintf("%s", err))
						continue
					}

					message.Answer(&plugin, "rule", string(groupObjectBytes))
				}

				return
			}

		}

		message.Answer(&plugin, "error", fmt.Sprintf("Table '%s' dont exist", payload))
		return
	}

	if command == "updateRule" {

		var jsonRule nftJSONRule
		err := json.Unmarshal([]byte(payload), &jsonRule)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// try to get the rule out from the table/chain
		if table, ok := nftConfig.Tables["tGopilot"]; ok {
			if chain, ok := table.Chains[jsonRule.ChainName]; ok {

				// create the rule, we would like to overwrite
				var newNftRule nftRule
				newNftRule.chain = chain
				newNftRule.index = 0
				newNftRule.Enabled = jsonRule.Enabled
				newNftRule.Policy = jsonRule.Policy
				newNftRule.Statement = jsonRule.Statement

				// position is not in this chain
				if jsonRule.Position > len(chain.Rules) {
					message.Answer(&plugin, "error", fmt.Sprintf("Rule with index '%v' not found", jsonRule.Position-1))
					return
				}

				// position is > 0, replace rule inside the chain
				if jsonRule.Position > 0 {
					newNftRule.index = jsonRule.Position - 1
					chain.Rules[newNftRule.index] = &newNftRule
				} else {
					newNftRule.index = len(chain.Rules)
					chain.Rules = append(chain.Rules, &newNftRule)
				}

				// send ok
				message.Answer(&plugin, "updateRuleOk", "")

				// send changed rule
				groupObjectBytes, err := json.Marshal(jsonRule)
				if err != nil {
					logging.Error("getRules", fmt.Sprintf("%s", err))
					message.Answer(&plugin, "error", fmt.Sprintf("%s", err))
					return
				}

				message.Answer(&plugin, "rule", string(groupObjectBytes))
				return
			}

			message.Answer(&plugin, "error", fmt.Sprintf("Chain '%s' not found", jsonRule.ChainName))
			return
		}
		message.Answer(&plugin, "error", fmt.Sprintf("Table '%s' not found", "tGopilot"))
		return
	}

	if command == "deleteRule" {

		var jsonRule nftJSONRule
		err := json.Unmarshal([]byte(payload), &jsonRule)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// try to get the rule out from the table/chain
		if table, ok := nftConfig.Tables["tGopilot"]; ok {
			if chain, ok := table.Chains[jsonRule.ChainName]; ok {
				chain.Rules = append(chain.Rules[:jsonRule.Position-1], chain.Rules[jsonRule.Position:]...)
				message.Answer(&plugin, "deleteRuleOk", "")
				return
			}
		}

	}

	if command == "moveRuleUp" {

		var jsonRule nftJSONRule
		err := json.Unmarshal([]byte(payload), &jsonRule)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// we can only move up, when the position is > 1
		if jsonRule.Position <= 1 {
			message.Answer(&plugin, "error", fmt.Sprintf("Can not move rule up"))
			return
		}

		// try to get the rule out from the table/chain
		if table, ok := nftConfig.Tables["tGopilot"]; ok {
			if chain, ok := table.Chains[jsonRule.ChainName]; ok {

				upperRule := chain.Rules[jsonRule.Position-2]
				myRule := chain.Rules[jsonRule.Position-1]

				chain.Rules[jsonRule.Position-2] = myRule
				chain.Rules[jsonRule.Position-1] = upperRule

				message.Answer(&plugin, "moveRuleUpOk", "")
				return
			}
		}

		return
	}

	if command == "moveRuleDown" {

		var jsonRule nftJSONRule
		err := json.Unmarshal([]byte(payload), &jsonRule)
		if err != nil {
			message.Answer(&plugin, "error", err.Error())
			return
		}

		// try to get the rule out from the table/chain
		if table, ok := nftConfig.Tables["tGopilot"]; ok {
			if chain, ok := table.Chains[jsonRule.ChainName]; ok {

				// we can only move up, when the position is > 1
				if jsonRule.Position > len(chain.Rules) {
					message.Answer(&plugin, "error", fmt.Sprintf("Can not move rule down"))
					return
				}

				myRule := chain.Rules[jsonRule.Position-1]
				nextRule := chain.Rules[jsonRule.Position]

				chain.Rules[jsonRule.Position-1] = nextRule
				chain.Rules[jsonRule.Position] = myRule

				message.Answer(&plugin, "moveRuleDownOk", "")
				return
			}
		}

		return
	}

}
