package daemonset

import "fmt"

func NodeLabelName(moduleName string) string {
	return fmt.Sprintf("oot.node.kubernetes.io/%s.ready", moduleName)
}
