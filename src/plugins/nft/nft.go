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
	"core/clog"
	"encoding/json"
	"plugins/core"
)

var useSudo = true

var logging clog.Logger

type nftJSONConfig struct {
	Tables map[string]*nftTable `json:"tables"`
}

var nftConfig nftJSONConfig

/*
Init the nft-modules
*/
func Init() {

	logging = clog.New("NFT")

	// load from config
	nftConfig, _ = loadFromConfig()

	nftConfig.applyAll()
	/*
		chain := nftConfig.Tables["tGopilot"].Chains["input"]
		rule := chain.ruleNew(nftPolicyAccept)
		rule.statementAdd([]string{"ct", "state", "related,established"})
		saveConfig(nftConfig)
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
}

func loadFromConfig() (nftJSONConfig, error) {

	needToSave := false

	// get the config
	jsonObject, _ := core.GetJsonObject("nft")
	if jsonObject == nil {
		jsonObject, _ = core.NewJsonObject("nft")
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
		saveConfig(jsonConfig)
	}

	// because the name and ids are keys inside the json, we need to at this missing infos in the structs
	for tableName, table := range jsonConfig.Tables {

		// fill out values which are not in the json
		table.name = tableName

		for chainName, chain := range table.Chains {

			// fill out values which are not in the json
			chain.table = table
			chain.name = chainName

			for ruleUUID, rule := range chain.Rules {

				// fill out values which are not in the json
				rule.uuid = ruleUUID
				rule.chain = chain

			}
		}
	}

	return jsonConfig, nil
}

func saveConfig(newConfig nftJSONConfig) error {

	// to json
	groupObjectBytes, err := json.Marshal(newConfig)
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
	core.SetJsonObject("nft", jsonObject)
	core.ConfigSave()
	return nil
}

func (config *nftJSONConfig) applyAll() error {

	// flush
	RulesetFlush()

	for _, table := range config.Tables {

		// Add Table
		table.Apply()

		for _, chain := range table.Chains {

			// Add chain
			chain.Apply()

			for _, rule := range chain.Rules {

				// Add rule
				rule.Apply()
			}
		}
	}

	return nil
}
