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
	"testing"

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
			},
			notWant: []string{`host_cache_size`},
		},
		{
			version: "8.0.37",
			want: []string{
				`(?m)^skip-host-cache$`,
				`(?m)^relay-log-info-repository\s*= TABLE$`,
				`(?m)^master-info-repository\s*= TABLE$`,
			},
			notWant: []string{`host_cache_size`},
		},
		{
			version: "8.4.9",
			want: []string{
				`(?m)^host_cache_size\s*= 0$`,
			},
			notWant: []string{
				`skip-host-cache`,
				`relay-log-info-repository`,
				`master-info-repository`,
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
