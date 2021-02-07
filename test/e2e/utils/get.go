package utils

import (
	"fmt"
	"strings"
)

type kubectlGetCtx struct {
	kubectlContext
	kind     string
	name     string
	selector string
}

type KubectlGet kubectlGetCtx

func (k *KubectlContext) Get(resource string) *KubectlGet {
	return &KubectlGet{
		kubectlContext: kubectlContext(*k),
		kind:           resource,
	}
}

func (k *KubectlGet) ByName(name string) KubectlGetOutput {
	k.name = name
	return KubectlGetOutput(*k)
}

func (k *KubectlGet) All() KubectlGetOutput {
	return KubectlGetOutput(*k)
}

func (k *KubectlGet) ByLabel(labelSelector string) KubectlGetOutput {
	k.selector = labelSelector
	return KubectlGetOutput(*k)
}

type KubectlGetOutput kubectlGetCtx

func (k KubectlGetOutput) getCMD() string {
	s := k.kubectlContext.getCMD()
	s += fmt.Sprintf(" get %s ", k.kind)
	if k.name != "" {
		s += fmt.Sprintf("%s", k.name)
	} else if k.selector != "" {
		s += fmt.Sprintf("-l %s", k.selector)
	}
	return s
}

func (k KubectlGetOutput) JSON() (string, error) {
	stdout, stderr, _, err := shell(k.getCMD() + " -o json")
	if err != nil {
		return stderr, err
	}
	return strings.TrimSpace(stdout), nil
}

func (k KubectlGetOutput) YAML() (string, error) {
	stdout, stderr, _, err := shell(k.getCMD() + " -o yaml")
	if err != nil {
		return stderr, err
	}
	return strings.TrimSpace(stdout), nil
}

func (k KubectlGetOutput) Exist() (bool, error) {
	stdout, _, _, err := shell(k.getCMD() + " --ignore-not-found")
	if err != nil {
		return false, err
	}
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return false, err
	}
	return true, err
}

func (k KubectlGetOutput) ObjectField(path string) (string, error) {
	fields, err := k.ObjectsField(path)
	if err != nil {
		return "", err
	}
	return fields[0], nil
}

func (k KubectlGetOutput) ObjectsField(path string) ([]string, error) {
	if k.name != "" {
		path = `'{` + path + `}'`
	} else {
		path = `'{range .items[*]}{` + path + `}{"\n"}{end}'`
	}
	stdout, _, _, err := shell(k.getCMD() + " -o jsonpath=" + path)
	if err != nil {
		return nil, err
	}
	return strings.Split(stdout, "\n"), nil
}

func (k KubectlGetOutput) ObjectFields(paths []string) (fields []string, err error) {
	out, err := k.ObjectsFields(paths)
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

func (k KubectlGetOutput) ObjectsFields(paths []string) (out [][]string, err error) {
	path := ""
	if k.name == "" {
		path = `'{range .items[*]}`
	}
	for i, p := range paths {
		if i != 0 {
			path += `{"\t"}`
		}
		path += "{" + p + "}"
	}
	if k.name == "" {
		path += `{"\n"}{end}'`
	}

	stdout, _, _, err := shell(k.getCMD() + " -o jsonpath=" + path)
	if err != nil {
		return nil, err
	}
	objects := strings.Split(stdout, "\n")
	for _, object := range objects {
		fields := strings.Split(object, "\t")
		if len(fields) != len(paths) {
			continue
		}
		out = append(out, fields)
	}
	return out, nil
}
