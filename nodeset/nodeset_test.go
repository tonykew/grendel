// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package nodeset

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testNodeSet struct {
	test   string
	result string
	length int
}

func TestNodeSetSimple(t *testing.T) {
	tests := []testNodeSet{
		testNodeSet{"cws-machin", "cws-machin", 1},
		testNodeSet{"cpn-d13-01", "cpn-d13-01", 1},
		testNodeSet{"cpn-d13-[01]", "cpn-d13-[01]", 1},
		testNodeSet{"cpn-d13-[01-10]", "cpn-d13-[01-10]", 10},
		testNodeSet{"cpn-k[08-09]-[02-24/2]-[01-02]", "cpn-k[08-09]-[02,04,06,08,10,12,14,16,18,20,22,24]-[01-02]", 48},
		testNodeSet{"cpn-q[06-09]-[36,35,32,31,28,27,17,16,13,12,09,08,05,04]-[01-02],cpn-q[06-09]-[20,23],cpn-q[07-08]-[39,40]-[01-02]", "cpn-q[06-09]-[04-05,08-09,12-13,16-17,27-28,31-32,35-36]-[01-02],cpn-q[06-09]-[20,23],cpn-q[07-08]-[39-40]-[01-02]", 128},
	}

	for _, nstest := range tests {
		n1, err := NewNodeSet(nstest.test)
		assert.Nil(t, err)
		assert.Equal(t, nstest.result, n1.String())
		assert.Equal(t, nstest.length, n1.Len())
	}
}

func TestNodeSetIterator(t *testing.T) {
	test := make([]string, 0)
	for i := 1; i < 11; i++ {
		test = append(test, fmt.Sprintf("cpn-d13-%02d", i))
	}

	n1, err := NewNodeSet("cpn-d13-[01-10]")
	assert.Nil(t, err)
	assert.Equal(t, 10, n1.Len())

	result := make([]string, 0)

	it := n1.Iterator()
	for it.Next() {
		result = append(result, it.Value())
	}

	assert.Equal(t, 10, len(result))
	assert.EqualValues(t, test, result)
}

func TestNodeSetJSON(t *testing.T) {
	tests := []testNodeSet{
		testNodeSet{`["cws-machin"]`, "cws-machin", 1},
		testNodeSet{`["cpn-d13-[01-10]","cpn-d14-[01-05]"]`, "cpn-d13-[01-10],cpn-d14-[01-05]", 15},
	}

	for _, nstest := range tests {
		var n1 NodeSet
		err := json.Unmarshal([]byte(nstest.test), &n1)
		assert.Nil(t, err)

		assert.Equal(t, nstest.result, n1.String())
		assert.Equal(t, nstest.length, n1.Len())

		data, err := json.Marshal(&n1)
		assert.Nil(t, err)

		var l1 []string
		err = json.Unmarshal(data, &l1)
		assert.Nil(t, err)

		var l2 []string
		err = json.Unmarshal([]byte(nstest.test), &l2)

		sort.Strings(l1)

		assert.Equal(t, l2, l1)
	}
}
