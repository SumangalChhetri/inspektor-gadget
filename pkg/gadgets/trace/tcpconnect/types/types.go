// Copyright 2022 The Inspektor Gadget authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"github.com/kinvolk/inspektor-gadget/pkg/columns"
	eventtypes "github.com/kinvolk/inspektor-gadget/pkg/types"
)

type Event struct {
	eventtypes.Event

	Pid       uint32 `json:"pid,omitempty" column:"pid,minWidth:7"`
	UID       uint32 `json:"uid,omitempty" column:"uid,minWidth:6,hide"`
	Comm      string `json:"comm,omitempty" column:"comm,maxWidth:16"`
	IPVersion int    `json:"ipversion,omitempty" column:"ip,width:2,fixed"`
	// For Saddr and Daddr:
	// Min: XXX.XXX.XXX.XXX (IPv4) = 15
	// Max: 0000:0000:0000:0000:0000:ffff:XXX.XXX.XXX.XXX (IPv4-mapped IPv6 address) = 45
	Saddr     string `json:"saddr,omitempty" column:"saddr,minWidth:15,maxWidth:45"`
	Daddr     string `json:"daddr,omitempty" column:"daddr,minWidth:15,maxWidth:45"`
	Dport     uint16 `json:"dport,omitempty" column:"dport,minWidth:type"`
	MountNsID uint64 `json:"mountnsid,omitempty" column:"mntns,width:12,hide"`
}

func GetColumns() *columns.Columns[Event] {
	return columns.MustCreateColumns[Event]()
}

func Base(ev eventtypes.Event) Event {
	return Event{
		Event: ev,
	}
}

func (e Event) GetBaseEvent() eventtypes.Event {
	return e.Event
}
