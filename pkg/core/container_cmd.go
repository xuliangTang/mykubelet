package core

import "os/exec"

type ContainerCmd struct {
	Cmd           *exec.Cmd `json:"cmd"`
	ContainerName string    `json:"container_name"`
	ExitCode      int       `json:"exit_code"`
	ExitError     error     `json:"exit_error"`
}

// Run 运行容器cmd
func (this *ContainerCmd) Run() {
	err := this.Cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			this.ExitCode = exitError.ExitCode()
		} else {
			this.ExitCode = -9999 // 其他错误
			this.ExitError = err
		}
	}
}
