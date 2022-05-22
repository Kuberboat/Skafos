package core

import (
	"fmt"

	kubeCore "p9t.io/kuberboat/pkg/api/core"
)

func GetPodSpecificPauseName(pod *kubeCore.Pod) string {
	return fmt.Sprintf("%v_%v", pod.UUID.String(), "pause")
}

func IsSameHostAddr(addrA, addrB string) bool {
	if (addrA == "localhost" && addrB == "127.0.0.1") ||
		(addrA == "127.0.0.1" && addrB == "localhost") {
		return true
	}
	return addrA == addrB
}
