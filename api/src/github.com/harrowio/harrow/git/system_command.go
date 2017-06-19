package git

import (
	"fmt"
	"sort"
	"strings"
)

type SystemCommand struct {
	Exec string   // name of the executable
	Args []string // arguments to pass to the executable
	Env  []string // environment in which the command runs
	Dir  string   // working directory of the command
}

func NewSystemCommand(exec string, args ...string) *SystemCommand {
	return &SystemCommand{
		Exec: exec,
		Args: args,
		Dir:  ".",
		Env:  []string{},
	}
}

func (self *SystemCommand) WorkingDirectory(dir string) *SystemCommand {
	self.Dir = dir
	return self
}

func (self *SystemCommand) SetEnv(name, value string) *SystemCommand {
	self.Env = append(self.Env, fmt.Sprintf("%s=%s", name, value))
	sort.Strings(self.Env)
	return self
}

func (self *SystemCommand) String() string {
	return fmt.Sprintf("(cd %s; %s %s %s)",
		self.Dir,
		strings.Join(self.Env, " "),
		self.Exec,
		strings.Join(self.QuotedArgs(), " "),
	)
}

func (self *SystemCommand) QuotedArgs() []string {
	result := []string{}
	for _, arg := range self.Args {
		quoted := arg
		if strings.ContainsAny(arg, "'\" \t\n\r?*!") {
			quoted = "'" + strings.Replace(arg, "'", "'\\''", -1) + "'"
		}
		result = append(result, quoted)
	}

	return result
}
