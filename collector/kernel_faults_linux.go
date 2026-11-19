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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
)

const kernelFaultsCollectorSubsystem = "kernel_faults"

type kernelFaultsFile struct {
	name string
	desc *prometheus.Desc
}

func newKernelFaultsFile(file, metric, help string) kernelFaultsFile {
	return kernelFaultsFile{
		name: file,
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, kernelFaultsCollectorSubsystem, metric),
			help, nil, nil,
		),
	}
}

// each counter is in /sys/kernel and present when the kernel was built with the matching option, presence is detected per file
var kernelFaultsFiles = []kernelFaultsFile{
	newKernelFaultsFile("hardlockup_count", "hard_lockups_total", "Total number of hard lockups detected since the system booted."),
	newKernelFaultsFile("oops_count", "oops_total", "Total number of kernel oopses since the system booted."),
	newKernelFaultsFile("rcu_stall_count", "rcu_stalls_total", "Total number of RCU stalls detected since the system booted."),
	newKernelFaultsFile("softlockup_count", "soft_lockups_total", "Total number of soft lockups detected since the system booted."),
	newKernelFaultsFile("warn_count", "warnings_total", "Total number of kernel WARN() calls since the system booted."),
}

type kernelFaultsCollector struct {
	logger *slog.Logger
}

func init() {
	registerCollector(kernelFaultsCollectorSubsystem, defaultDisabled, NewKernelFaultsCollector)
}

// NewKernelFaultsCollector returns a collector exposing the kernel fault counters (oops, lockups, warn, and RCU stalls) from /sys/kernel
func NewKernelFaultsCollector(logger *slog.Logger) (Collector, error) {
	return &kernelFaultsCollector{logger: logger}, nil
}

func (c *kernelFaultsCollector) Update(ch chan<- prometheus.Metric) error {
	found := false
	for _, f := range kernelFaultsFiles {
		path := sysFilePath(filepath.Join("kernel", f.name))
		val, err := readUintFromFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				c.logger.Debug("kernel fault counter not available, skipping", "file", f.name)
				continue
			}
			return fmt.Errorf("failed to read %s: %w", path, err)
		}
		found = true
		ch <- prometheus.MustNewConstMetric(f.desc, prometheus.CounterValue, float64(val))
	}
	if !found {
		return ErrNoData
	}
	return nil
}
