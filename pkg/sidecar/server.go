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

package sidecar

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/blang/semver"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/bitpoke/mysql-operator/pkg/util/constants"
)

const (
	backupStatusTrailer = "X-Backup-Status"
	backupSuccessful    = "Success"
	backupFailed        = "Failed"
)

type server struct {
	cfg *Config
	http.Server

	// isDonorHealthy gates the backup endpoint on the local replication
	// state; a field so tests can stub the MySQL round-trip
	isDonorHealthy func() error
}

func newServer(cfg *Config, stop <-chan struct{}) *server {
	mux := http.NewServeMux()
	srv := &server{
		cfg: cfg,
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", serverPort),
			Handler: mux,
		},
		isDonorHealthy: func() error { return checkDonorReplication(cfg) },
	}

	// Add handle functions
	mux.HandleFunc(serverProbeEndpoint, srv.healthHandler)
	mux.Handle(serverBackupEndpoint, maxClients(http.HandlerFunc(srv.backupHandler), 1))

	// Shutdown gracefully the http server
	go func() {
		<-stop // wait for stop signal
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Error(err, "failed to stop http server")

		}
	}()

	return srv
}

// nolint: unparam
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Error(err, "failed writing request")
	}
}

func (s *server) backupHandler(w http.ResponseWriter, r *http.Request) {

	if !s.isAuthenticated(r) {
		http.Error(w, "Not authenticated!", http.StatusForbidden)
		return
	}

	// xtrabackup holds LOCK INSTANCE FOR BACKUP for the whole stream; on a
	// replica whose master just died that lock lands exactly in orchestrator's
	// DeadMaster recovery window and makes its STOP SLAVE / RESET SLAVE ALL
	// on this node fail (orchestrator proceeds with the promotion anyway, and
	// the leftover replication config re-attaches this node to the resurrected
	// old master). Refuse to serve until replication is out of flux; the
	// requester's init container retries.
	if err := s.isDonorHealthy(); err != nil {
		log.Info("refusing to serve backup: donor replication state is not safe", "reason", err.Error())
		http.Error(w, fmt.Sprintf("donor not ready to serve a backup: %s", err), http.StatusServiceUnavailable)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "HTTP server does not support streaming!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Trailer", backupStatusTrailer)

	// nolint: gosec
	xtrabackup := exec.Command(xtrabackupCommand, s.cfg.XtrabackupArgs()...)
	xtrabackup.Stderr = os.Stderr

	stdout, err := xtrabackup.StdoutPipe()
	if err != nil {
		log.Error(err, "failed to create stdout pipe")
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	defer func() {
		// don't care
		_ = stdout.Close()
	}()

	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(w, stdout); err != nil {
		log.Error(err, "failed to copy buffer")
		http.Error(w, "buffer copy failed", http.StatusInternalServerError)
		return
	}

	if err := xtrabackup.Wait(); err != nil {
		log.Error(err, "failed waiting for xtrabackup to finish")
		w.Header().Set(backupStatusTrailer, backupFailed)
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	// success
	w.Header().Set(backupStatusTrailer, backupSuccessful)
	flusher.Flush()
}

func (s *server) isAuthenticated(r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	return ok && user == s.cfg.BackupUser && pass == s.cfg.BackupPassword
}

// maxClients limit an http endpoint to allow just n max concurrent connections
func maxClients(h http.Handler, n int) http.Handler {
	sema := make(chan struct{}, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()

		h.ServeHTTP(w, r)
	})
}

func prepareURL(svc string, endpoint string) string {
	if !strings.Contains(svc, ":") {
		svc = fmt.Sprintf("%s:%d", svc, serverPort)
	}
	return fmt.Sprintf("http://%s%s", svc, endpoint)
}

func transportWithTimeout(connectTimeout time.Duration) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// requestABackup connects to specified host and endpoint and gets the backup
func requestABackup(cfg *Config, host, endpoint string) (*http.Response, error) {
	log.Info("initialize a backup", "host", host, "endpoint", endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	// always waiting for a cluster
	err := wait.PollImmediateUntil(time.Minute, func() (done bool, err error) {
		return isServiceAvailable(host), nil
	}, ctx.Done())

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", prepareURL(host, endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %s", err)
	}

	// set authentication user and password
	req.SetBasicAuth(cfg.BackupUser, cfg.BackupPassword)

	client := &http.Client{}
	client.Transport = transportWithTimeout(serverConnectTimeout)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		status := "unknown"
		if resp != nil {
			status = resp.Status
		}
		return nil, fmt.Errorf("fail to get backup: %s, code: %s", err, status)
	}

	return resp, nil
}

func checkBackupTrailers(resp *http.Response) error {
	if values, ok := resp.Trailer[backupStatusTrailer]; !ok || !stringInSlice(backupSuccessful, values) {
		// backup is failed, remove from remote
		return fmt.Errorf("backup failed to be taken: no 'Success' trailer found")
	}

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// checkDonorReplication reports whether this node's replication state is safe
// to serve a backup stream from: either it has no replica configuration at all
// (a master or a bootstrap node), or both replication threads are running. A
// replica whose IO thread is reconnecting or stopped is mid-failover or
// broken: a clone taken from it captures a topology that is being torn down.
func checkDonorReplication(cfg *Config) error {
	// the replication user is the same one xtrabackup runs with and holds
	// REPLICATION CLIENT explicitly (sys_operator only implies it via SUPER,
	// which MySQL 8.4 drops)
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/?timeout=5s&readTimeout=5s",
		cfg.ReplicationUser, cfg.ReplicationPassword, mysqlPort)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection to the local mysql: %s", err)
	}
	defer func() {
		if cErr := db.Close(); cErr != nil {
			log.Error(cErr, "failed to close mysql connection")
		}
	}()

	rows, err := db.Query(showReplicaStatusQuery(cfg.MySQLVersion))
	if err != nil {
		return fmt.Errorf("failed to query replication status: %s", err)
	}
	defer func() {
		if cErr := rows.Close(); cErr != nil {
			log.Error(cErr, "failed to close rows")
		}
	}()

	if !rows.Next() {
		// no replica configuration: a master or a standalone node
		return rows.Err()
	}

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to read replication status columns: %s", err)
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	if err := rows.Scan(scanArgs...); err != nil {
		return fmt.Errorf("failed to scan replication status: %s", err)
	}

	columnValue := func(names ...string) string {
		for i, col := range columns {
			for _, name := range names {
				if col == name {
					return string(values[i])
				}
			}
		}
		return ""
	}

	return evalReplicationThreads(
		columnValue("Slave_IO_Running", "Replica_IO_Running"),
		columnValue("Slave_SQL_Running", "Replica_SQL_Running"),
	)
}

func evalReplicationThreads(ioRunning, sqlRunning string) error {
	if ioRunning == "Yes" && sqlRunning == "Yes" {
		return nil
	}
	return fmt.Errorf("replication is configured but not healthy: io_running=%q, sql_running=%q",
		ioRunning, sqlRunning)
}

func showReplicaStatusQuery(v semver.Version) string {
	if v.GTE(constants.MySQL84) {
		return "SHOW REPLICA STATUS"
	}
	return "SHOW SLAVE STATUS"
}
