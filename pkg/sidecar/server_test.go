/*
Copyright 2026 KSmartData

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

package sidecar

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test backup server donor gating", func() {

	var (
		cfg *Config
		srv *server
	)

	BeforeEach(func() {
		cfg = &Config{
			Hostname:       "cluster-mysql-1",
			BackupUser:     "backup-user",
			BackupPassword: "backup-password",
		}
		srv = &server{cfg: cfg}
	})

	requestBackup := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", serverBackupEndpoint, nil)
		req.SetBasicAuth(cfg.BackupUser, cfg.BackupPassword)
		rec := httptest.NewRecorder()
		srv.backupHandler(rec, req)
		return rec
	}

	It("should refuse to serve a backup while the donor replication is not healthy", func() {
		srv.isDonorHealthy = func() error {
			return fmt.Errorf("replication is configured but not healthy")
		}

		rec := requestBackup()
		Expect(rec.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(rec.Body.String()).To(ContainSubstring("donor not ready to serve a backup"))
	})

	It("should proceed past the gate when the donor replication is healthy", func() {
		srv.isDonorHealthy = func() error { return nil }

		// no xtrabackup binary in the unit test environment: reaching the
		// xtrabackup exec failure proves the gate let the request through
		rec := requestBackup()
		Expect(rec.Code).NotTo(Equal(http.StatusServiceUnavailable))
	})

	It("should check authentication before the donor gate", func() {
		gateCalled := false
		srv.isDonorHealthy = func() error {
			gateCalled = true
			return nil
		}

		req := httptest.NewRequest("GET", serverBackupEndpoint, nil)
		rec := httptest.NewRecorder()
		srv.backupHandler(rec, req)

		Expect(rec.Code).To(Equal(http.StatusForbidden))
		Expect(gateCalled).To(BeFalse())
	})
})

var _ = Describe("Test donor replication state evaluation", func() {

	It("should accept a replica with both threads running", func() {
		Expect(evalReplicationThreads("Yes", "Yes")).To(Succeed())
	})

	It("should reject a replica reconnecting to a dead master", func() {
		Expect(evalReplicationThreads("Connecting", "Yes")).To(HaveOccurred())
	})

	It("should reject a replica with stopped threads", func() {
		Expect(evalReplicationThreads("No", "No")).To(HaveOccurred())
	})

	It("should reject a replica with a stopped SQL thread", func() {
		Expect(evalReplicationThreads("Yes", "No")).To(HaveOccurred())
	})

	It("should use the legacy statement before MySQL 8.4", func() {
		Expect(showReplicaStatusQuery(semver.MustParse("5.7.44"))).To(Equal("SHOW SLAVE STATUS"))
		Expect(showReplicaStatusQuery(semver.MustParse("8.0.37"))).To(Equal("SHOW SLAVE STATUS"))
	})

	It("should use the replica spelling starting with MySQL 8.4", func() {
		Expect(showReplicaStatusQuery(semver.MustParse("8.4.9"))).To(Equal("SHOW REPLICA STATUS"))
	})
})

var _ = Describe("Test donor master reachability probe", func() {

	It("should accept a replica whose master answers on its port", func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).To(Succeed())
		defer func() {
			_ = listener.Close()
		}()

		host, port, err := net.SplitHostPort(listener.Addr().String())
		Expect(err).To(Succeed())
		Expect(checkMasterReachable(host, port)).To(Succeed())
	})

	It("should reject a replica whose master is gone", func() {
		// grab a port that is guaranteed unused, then release it
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).To(Succeed())
		host, port, err := net.SplitHostPort(listener.Addr().String())
		Expect(err).To(Succeed())
		Expect(listener.Close()).To(Succeed())

		err = checkMasterReachable(host, port)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unreachable"))
	})

	It("should skip the probe without a replica configuration", func() {
		Expect(checkMasterReachable("", "")).To(Succeed())
	})
})
