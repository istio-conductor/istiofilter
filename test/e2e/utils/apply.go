package utils

import "fmt"

func (k *KubectlContext) Apply(file string) (stdout string, err error) {
	s := (kubectlContext)(*k).getCMD()
	s += fmt.Sprintf(" apply -f %s", file)
	stdout, _, _, err = shell(s)
	return stdout, err
}

func (k *KubectlContext) DeleteByFile(file string) (stdout string, err error) {
	s := (kubectlContext)(*k).getCMD()
	s += fmt.Sprintf(" delete -f %s", file)
	stdout, _, _, err = shell(s)
	return stdout, err
}
func (k *KubectlContext) WaitAllDeployments() {
	s := (kubectlContext)(*k).getCMD()
	s += fmt.Sprintf(" wait --for=condition=available --timeout=60s --all deployments")
}
