package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	verbose = false
)

func EnableVerbose() {
	verbose = true
}

func DisableVerbose() {
	verbose = false
}

type KubectlContext kubectlContext

type kubectlContext struct {
	kubeConfig string
	namespace  string
}

func (k kubectlContext) getCMD() string {
	s := "kubectl"
	if k.kubeConfig != "" {
		s += " --kubeconfig=" + k.kubeConfig
	}
	if k.namespace != "" {
		s += " -n " + k.namespace
	}
	return s
}

func Kubectl() *KubectlContext {
	return &KubectlContext{}
}

func (k *KubectlContext) Namespace(ns string) *KubectlContext {
	k.namespace = ns
	return k
}

func (k *KubectlContext) KubeConfigFile(conf string) *KubectlContext {
	k.kubeConfig = conf
	return k
}

func shell(str string) (stdout, stderr string, code int, err error) {
	cmd := exec.Command("sh", "-c", str)
	stdoutBuffer := bytes.NewBuffer(nil)
	stderrBuffer := bytes.NewBuffer(nil)
	if verbose == true {
		fmt.Println("+", str)
		cmd.Stdout, cmd.Stderr = io.MultiWriter(os.Stdout, stdoutBuffer), io.MultiWriter(os.Stdout, stderrBuffer)
	} else {
		cmd.Stdout, cmd.Stderr = stdoutBuffer, stderrBuffer
	}
	err = cmd.Run()
	fmt.Println()
	code = cmd.ProcessState.ExitCode()
	return stdoutBuffer.String(), stderrBuffer.String(), code, err
}
