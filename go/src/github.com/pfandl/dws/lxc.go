package dws

import (
	"gopkg.in/lxc/go-lxc.v2"
	"os/exec"
	"strings"
)

type VirtualMachine struct {
	Container      *lxc.Container
	IpV4Address    string
	IpV4MacAddress string
	UtsName        string
}

type RepoQuery struct {
	PackageManager string
	Command        string
	Arguments      []string
}

var (
	PackageQueries = []RepoQuery{
		// Debian derivatives
		{
			"apt-get",
			"",
			nil,
		},
		// RedHat Centos
		{
			"yum",
			"repoquery",
			nil,
		},
		// RedHat Fedora
		{
			"dnf",
			"dnf",
			[]string{
				"repoquery",
				"-l",
				"lxc-templates",
			},
		},
		// Fallback
		{
			"ls",
			"ls",
			[]string{
				"/usr/share/lxc/templates",
			},
		},
	}
)

func GetTemplates() ([]string, error) {
	var cmd *string = nil
	var arg *[]string = nil
	for _, query := range PackageQueries {
		err := exec.Command(query.PackageManager).Run()
		if err == nil {
			cmd = &query.Command
			arg = &query.Arguments
			break
		}
	}
	if cmd == nil {
		return nil, ErrLxcNoTemplatesInstalled
	}
	out, err := exec.Command(*cmd, *arg...).Output()
	if err != nil {
		return nil, err
	}
	res := strings.Split(string(out), "\n")
	if len(res) == 0 {
		return nil, ErrLxcNoTemplatesFound
	}
	return res, nil
}
