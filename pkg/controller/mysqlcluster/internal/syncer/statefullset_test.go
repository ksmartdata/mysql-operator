/*
Copyright 2018 Pressinfra SRL

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

package mysqlcluster

import (
	"testing"

	"github.com/blang/semver"
)

// < 8.4 must get no extra arguments (frozen container spec); >= 8.4 needs TLS
// for the percona-toolkit connections against caching_sha2_password users.
func TestPtToolkitTLSArgs(t *testing.T) {
	for _, v := range []string{"5.7.44", "8.0.37"} {
		if got := ptToolkitTLSArgs(semver.MustParse(v)); got != nil {
			t.Errorf("ptToolkitTLSArgs(%s) = %v, want nil", v, got)
		}
	}

	got := ptToolkitTLSArgs(semver.MustParse("8.4.9"))
	if len(got) != 2 || got[0] != "--mysql_ssl" || got[1] != "1" {
		t.Errorf(`ptToolkitTLSArgs(8.4.9) = %v, want ["--mysql_ssl" "1"]`, got)
	}
}
