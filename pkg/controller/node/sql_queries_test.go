/*
Copyright 2019 Pressinfra SRL

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package node

import (
	"strings"
	"testing"

	"github.com/blang/semver"
)

var (
	v8037 = semver.MustParse("8.0.37")
	v849  = semver.MustParse("8.4.9")
)

// The < 8.4 queries are frozen literals: they must stay exactly what the
// operator has always run against existing clusters.
func TestReplicationQueriesLegacyFrozen(t *testing.T) {
	wantChangeMaster := `
      STOP SLAVE;
	  CHANGE MASTER TO MASTER_AUTO_POSITION=1,
		MASTER_HOST=?,
		MASTER_USER=?,
		MASTER_PASSWORD=?,
		MASTER_CONNECT_RETRY=?;
	`
	if got := changeMasterToQuery(v8037); got != wantChangeMaster {
		t.Errorf("changeMasterToQuery for < 8.4 changed:\ngot:  %q\nwant: %q", got, wantChangeMaster)
	}

	if got := startReplicationQuery(v8037); got != "START SLAVE;" {
		t.Errorf("startReplicationQuery for < 8.4 changed: %q", got)
	}

	wantFallback := `
		  reset slave;
		  start slave IO_THREAD;
		  stop slave IO_THREAD;
		  reset slave;
		  start slave;
		`
	if got := startReplicationFallbackQuery(v8037); got != wantFallback {
		t.Errorf("startReplicationFallbackQuery for < 8.4 changed:\ngot:  %q\nwant: %q", got, wantFallback)
	}

	if got := resetBinaryLogsQuery(v8037); got != "RESET MASTER" {
		t.Errorf("resetBinaryLogsQuery for < 8.4 changed: %q", got)
	}
}

func TestReplicationQueries84(t *testing.T) {
	for query, wants := range map[string][]string{
		changeMasterToQuery(v849):           {"STOP REPLICA;", "CHANGE REPLICATION SOURCE TO SOURCE_AUTO_POSITION=1", "SOURCE_HOST=?", "SOURCE_CONNECT_RETRY=?"},
		startReplicationQuery(v849):         {"START REPLICA;"},
		startReplicationFallbackQuery(v849): {"RESET REPLICA;", "START REPLICA IO_THREAD;", "STOP REPLICA IO_THREAD;"},
		resetBinaryLogsQuery(v849):          {"RESET BINARY LOGS AND GTIDS"},
	} {
		for _, want := range wants {
			if !strings.Contains(query, want) {
				t.Errorf("8.4 query should contain %q, got:\n%s", want, query)
			}
		}
		for _, notWant := range []string{"SLAVE", "slave", "MASTER", "master"} {
			if strings.Contains(query, notWant) {
				t.Errorf("8.4 query should NOT contain %q, got:\n%s", notWant, query)
			}
		}
	}
}
