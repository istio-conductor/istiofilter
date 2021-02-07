package utils

import (
	"fmt"

	"github.com/alessio/shellescape"
)

type kubectlLogCtx struct {
	kubectlContext
	name      string
	container string
}

type KubectlLog kubectlExecCtx

func (k *KubectlContext) Log(name string, container string) *KubectlLog {
	return &KubectlLog{
		kubectlContext: kubectlContext(*k),
		name:           name,
		container:      container,
	}
}

func (e *KubectlLog) Last(size int) (string, error) {
	s := e.kubectlContext.getCMD()
	s += fmt.Sprintf(" log %s", shellescape.Quote(e.name))
	if e.container != "" {
		s += fmt.Sprintf(" -c %s", shellescape.Quote(e.container))
	}
	s += fmt.Sprintf("--tail %d", size)
	stdout, _, _, err := shell(s)
	if err != nil {
		return "", err
	}
	return stdout, nil
}
