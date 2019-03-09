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

type nftFamily int

const (
	nftFAMILYIP   nftFamily = 0
	nftFAMILYIP6  nftFamily = 1
	nftFAMILYINET nftFamily = 2
)

func (family nftFamily) String() string {

	names := [...]string{
		"ip",
		"ip6",
		"inet",
	}

	return names[family]
}
