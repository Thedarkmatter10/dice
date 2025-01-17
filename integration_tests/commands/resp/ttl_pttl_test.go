// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package resp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTLPTTL(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		{
			name:       "TTL Simple Value",
			commands:   []string{"SET foo bar", "GETEX foo ex 5", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", "bar", int64(5)},
			assertType: []string{"equal", "equal", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "PTTL Simple Value",
			commands:   []string{"SET foo bar", "GETEX foo px 5000", "GETEX foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", "bar", int64(5000)},
			assertType: []string{"equal", "equal", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "TTL & PTTL Non-Existent Key",
			commands:   []string{"TTL foo", "PTTL foo"},
			expected:   []interface{}{int64(-2), int64(-2)},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "TTL & PTTL without Expiry",
			commands:   []string{"SET foo bar", "GET foo", "TTL foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), int64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "TTL & PTTL with Persist",
			commands:   []string{"SET foo bar", "GETEX foo persist", "TTL foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), int64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "TTL & PTTL with Expire and Expired Key",
			commands:   []string{"SET foo bar", "GETEX foo ex 5", "GET foo", "TTL foo", "PTTL foo", "TTL foo", "PTTL foo"},
			expected:   []interface{}{"OK", "bar", "bar", int64(5), int64(5000), int64(-2), int64(-2)},
			assertType: []string{"equal", "equal", "equal", "assert", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"foo"}, store)
			FireCommand(conn, "DEL foo")
			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
				if tc.assertType[i] == "equal" {
					assert.Equal(t, tc.expected[i], result)
				} else if tc.assertType[i] == "assert" {
					assert.True(t, result.(int64) <= tc.expected[i].(int64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
