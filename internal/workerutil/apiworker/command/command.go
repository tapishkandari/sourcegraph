package command

//
// TODO - redo
//

type Cmd struct {
	image string
	// TODO(efritz) - currently treated as arguments and doesn't
	// support the intended image/commands setup, where each command
	// would be equivalent to a line in a bash script. Need to figure
	// out the best way to supply it to the underlying shell.
	command []string
	wd      string
	env     map[string]string
}

func NewCmd(image string, command ...string) *Cmd {
	return &Cmd{
		image:   image,
		command: command,
		env:     map[string]string{},
	}
}

func (cmd *Cmd) SetWd(wd string) *Cmd {
	cmd.wd = wd
	return cmd
}

func (cmd *Cmd) AddEnv(key, value string) *Cmd {
	cmd.env[key] = value
	return cmd
}
