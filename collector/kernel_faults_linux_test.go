// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !nokernel_faults

package collector

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type testKernelFaultsCollector struct {
	kc Collector
}

func (c testKernelFaultsCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.kc.Update(ch); err != nil {
		panic(err)
	}
}

func (c testKernelFaultsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func TestKernelFaultsCollector(t *testing.T) {
	prev := *sysPath
	t.Cleanup(func() { *sysPath = prev })
	*sysPath = "fixtures/sys"

	c, err := NewKernelFaultsCollector(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	expected := `
# HELP node_kernel_faults_hard_lockups_total Total number of hard lockups detected since the system booted.
# TYPE node_kernel_faults_hard_lockups_total counter
node_kernel_faults_hard_lockups_total 1
# HELP node_kernel_faults_oops_total Total number of kernel oopses since the system booted.
# TYPE node_kernel_faults_oops_total counter
node_kernel_faults_oops_total 2
# HELP node_kernel_faults_rcu_stalls_total Total number of RCU stalls detected since the system booted.
# TYPE node_kernel_faults_rcu_stalls_total counter
node_kernel_faults_rcu_stalls_total 3
# HELP node_kernel_faults_soft_lockups_total Total number of soft lockups detected since the system booted.
# TYPE node_kernel_faults_soft_lockups_total counter
node_kernel_faults_soft_lockups_total 4
# HELP node_kernel_faults_warnings_total Total number of kernel WARN() calls since the system booted.
# TYPE node_kernel_faults_warnings_total counter
node_kernel_faults_warnings_total 5
`
	if err := testutil.CollectAndCompare(testKernelFaultsCollector{c}, strings.NewReader(expected)); err != nil {
		t.Fatal(err)
	}
}

func TestKernelFaultsCollectorNoData(t *testing.T) {
	prev := *sysPath
	t.Cleanup(func() { *sysPath = prev })
	*sysPath = "fixtures/sys/nonexistent"

	c, err := NewKernelFaultsCollector(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan prometheus.Metric, 8)
	if err := c.Update(ch); !errors.Is(err, ErrNoData) {
		t.Fatalf("want ErrNoData, got %v", err)
	}
	if got := len(ch); got != 0 {
		t.Fatalf("want 0 metrics, got %d", got)
	}
}
