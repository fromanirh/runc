package affinity

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/opencontainers/runc/libcontainer/configs"
)

const (
	NilCPUID int = -1
)

func SetAffinityToEnvVar(cpuid int) string {
	return fmt.Sprintf("FORCE_AFFINITY=%d", cpuid)
}

func GetAffinity(config configs.Config, env []string) (int, error) {
	cpuID := NilCPUID
	if config.Cgroups != nil {
		cpuIDs, err := parseLinuxCpuset(config.Cgroups.Resources.CpusetCpus)
		if err != nil {
			return -1, err
		}
		cpuID = cpuIDs[0]
	} else if env != nil {
		cpuID = getAffinityFromEnv(env)
	}
	return cpuID, nil
}

func getAffinityFromEnv(env []string) int {
	for _, envVar := range env {
		if strings.HasPrefix(envVar, "FORCE_AFFINITY") {
			return getAffinityFromEnvVar(envVar)
		}
	}
	return NilCPUID
}

func getAffinityFromEnvVar(env string) int {
	var cpuid int
	n, err := fmt.Sscanf(env, "FORCE_AFFINITY=%d", &cpuid)
	if n != 1 || err != nil {
		return NilCPUID
	}
	return cpuid
}

func parseLinuxCpuset(s string) ([]int, error) {
	cpus := []int{}
	if len(s) == 0 {
		return cpus, nil
	}

	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if !strings.Contains(item, "-") {
			// single cpu: "2"
			cpuid, err := strconv.Atoi(item)
			if err != nil {
				return cpus, err
			}
			cpus = append(cpus, cpuid)
			continue
		}

		// range of cpus: "0-3"
		cpuRange := strings.SplitN(item, "-", 2)
		cpuBegin, err := strconv.Atoi(cpuRange[0])
		if err != nil {
			return cpus, err
		}
		cpuEnd, err := strconv.Atoi(cpuRange[1])
		if err != nil {
			return cpus, err
		}
		for cpuid := cpuBegin; cpuid <= cpuEnd; cpuid++ {
			cpus = append(cpus, cpuid)
		}
	}

	sort.Ints(cpus)
	return cpus, nil
}
