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

package pluginnft

import (
	"fmt"
	"os/exec"
)

/*
RulesetFlush flush all rules of netfilter
*/
func RulesetFlush() error {

	// the command
	cmd := exec.Command("sudo", "nft", "flush", "ruleset")
	logging.Info("RulesetFlush", fmt.Sprintf("%s", cmd.Args))

	// run it
	cmdErr := cmd.Run()
	if cmdErr != nil {
		logging.Error("RulesetFlush", fmt.Sprintf("%s", cmdErr))
		return cmdErr
	}

	return nil

}
