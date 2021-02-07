package utils

import (
	"fmt"

	"github.com/alessio/shellescape"
)

type kubectlExecCtx struct {
	kubectlContext
	name      string
	container string
}

type KubectlExec kubectlExecCtx

func (k *KubectlContext) ExecInPod(name string) *KubectlExec {
	return &KubectlExec{
		kubectlContext: kubectlContext(*k),
		name:           name,
	}
}

func (e *KubectlExec) InContainer(c string) *KubectlExec {
	e.container = c
	return e
}

func (e *KubectlExec) Shell(cmd string) (string, error) {
	s := e.kubectlContext.getCMD()
	s += fmt.Sprintf(" exec %s", shellescape.Quote(e.name))
	if e.container != "" {
		s += fmt.Sprintf(" -c %s", shellescape.Quote(e.container))
	}
	s += fmt.Sprintf(" -- sh -c %s", cmd)
	stdout, _, _, err := shell(s)
	if err != nil {
		return "", err
	}
	return stdout, nil
}
