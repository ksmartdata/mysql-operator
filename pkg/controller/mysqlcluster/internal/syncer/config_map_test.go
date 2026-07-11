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
	"regexp"
	"strings"
	"testing"

	"github.com/blang/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/bitpoke/mysql-operator/pkg/apis/mysql/v1alpha1"
	"github.com/bitpoke/mysql-operator/pkg/internal/mysqlcluster"
)

// TestBuildMysqlConfDataVersionForks locks the semantics of the 8.4 version
// fork in my.cnf generation. Byte-level compatibility of the < 8.4 output is
// guarded separately by the chainsaw E2E golden tests
// (test/e2e-chainsaw/golden/).
func TestBuildMysqlConfDataVersionForks(t *testing.T) {
	tests := []struct {
		version string
		want    []string
		notWant []string
	}{
		{
			version: "5.7.44",
			want: []string{
				`(?m)^skip-host-cache$`,
				`(?m)^relay-log-info-repository\s*= TABLE$`,
				`(?m)^master-info-repository\s*= TABLE$`,
				`(?m)^log-slave-updates\s*= on$`,
				`(?m)^skip-slave-start\s*= on$`,
			},
			notWant: []string{
				`host_cache_size`,
				`log-replica-updates`,
				`skip-replica-start`,
				`default-authentication-plugin`,
			},
		},
		{
			version: "8.0.37",
			want: []string{
				`(?m)^skip-host-cache$`,
				`(?m)^relay-log-info-repository\s*= TABLE$`,
				`(?m)^master-info-repository\s*= TABLE$`,
				`(?m)^log-slave-updates\s*= on$`,
				`(?m)^skip-slave-start\s*= on$`,
				`(?m)^default-authentication-plugin\s*= mysql_native_password$`,
			},
			notWant: []string{
				`host_cache_size`,
				`log-replica-updates`,
				`skip-replica-start`,
			},
		},
		{
			version: "8.4.9",
			want: []string{
				`(?m)^host_cache_size\s*= 0$`,
				`(?m)^log-replica-updates\s*= on$`,
				`(?m)^skip-replica-start\s*= on$`,
			},
			notWant: []string{
				`skip-host-cache`,
				`relay-log-info-repository`,
				`master-info-repository`,
				`log-slave-updates`,
				`skip-slave-start`,
				`default-authentication-plugin`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			cluster := mysqlcluster.New(&api.MysqlCluster{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec:       api.MysqlClusterSpec{MysqlVersion: tt.version},
			})

			data, err := buildMysqlConfData(cluster)
			if err != nil {
				t.Fatalf("buildMysqlConfData: %v", err)
			}

			for _, pattern := range tt.want {
				if !regexp.MustCompile(pattern).MatchString(data) {
					t.Errorf("my.cnf for %s should match %q, got:\n%s", tt.version, pattern, data)
				}
			}
			for _, pattern := range tt.notWant {
				if regexp.MustCompile(pattern).MatchString(data) {
					t.Errorf("my.cnf for %s should NOT match %q, got:\n%s", tt.version, pattern, data)
				}
			}
		})
	}
}

func TestBuildBashPreStopVersionForks(t *testing.T) {
	old := buildBashPreStop(semver.MustParse("8.0.37"))
	for _, want := range []string{`show slave status\G`, `show slave hosts\G`} {
		if !strings.Contains(old, want) {
			t.Errorf("preStop for < 8.4 should contain %q", want)
		}
	}

	new84 := buildBashPreStop(semver.MustParse("8.4.9"))
	for _, want := range []string{`SHOW REPLICA STATUS\G`, `SHOW REPLICAS`} {
		if !strings.Contains(new84, want) {
			t.Errorf("preStop for 8.4 should contain %q", want)
		}
	}
	for _, notWant := range []string{"show slave", "StmtHolder"} {
		if strings.Contains(new84, notWant) {
			t.Errorf("preStop for 8.4 should NOT contain %q", notWant)
		}
	}
}

func TestBuildSemiSyncConfVersionForks(t *testing.T) {
	// the < 8.4 snippet is a frozen literal: existing clusters using the
	// rpl_semi_sync_enabled annotation must not see a ConfigMap change
	oldWant := `
	plugin-load-add	 = "semisync_master.so;semisync_slave.so"
	rpl_semi_sync_master_enabled 	=	1
	rpl_semi_sync_slave_enabled		=	1
	rpl_semi_sync_master_wait_for_slave_count	=	1
	`
	if got := buildSemiSyncConf(semver.MustParse("8.0.37"), 1); got != oldWant {
		t.Errorf("semi-sync snippet for < 8.4 changed:\ngot:  %q\nwant: %q", got, oldWant)
	}

	new84 := buildSemiSyncConf(semver.MustParse("8.4.9"), 2)
	for _, want := range []string{
		`plugin-load-add	 = "semisync_source.so;semisync_replica.so"`,
		"rpl_semi_sync_source_enabled",
		"rpl_semi_sync_replica_enabled",
		"rpl_semi_sync_source_wait_for_replica_count	=	2",
	} {
		if !strings.Contains(new84, want) {
			t.Errorf("semi-sync snippet for 8.4 should contain %q", want)
		}
	}
	if strings.Contains(new84, "semisync_master") || strings.Contains(new84, "rpl_semi_sync_master") {
		t.Errorf("semi-sync snippet for 8.4 should not contain master/slave names, got:\n%s", new84)
	}
}
