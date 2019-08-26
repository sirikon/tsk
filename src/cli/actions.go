package cli

import (
	"io"
	"os"
	"os/exec"

	"github.com/sirikon/tsk/src/application"
)

type CLI struct {
	CWD           string
	Out           io.Writer
	Err           io.Writer
	In            io.Reader
	ColorsEnabled bool
}

func (c *CLI) Run(args []string) int {
	if len(args) == 0 {
		return c.index()
	}
	return c.runCommand(args)
}

// Index .
func (c *CLI) index() (result int) {
	defer HandlePanic(&result, c.Out)
	p := c.getPrinter()
	project := c.getProject()

	p.header(project.TskFile)
	for _, command := range project.Commands {
		p.command(command, project, 0)
	}
	p.emptyLine()

	return 0
}

func (c *CLI) runCommand(args []string) (result int) {
	defer HandlePanic(&result, c.Out)
	p := c.getPrinter()
	project := c.getProject()

	command := findCommand(args, project.Commands)

	if command == nil || !command.Runnable {
		p.line("Command not found")
		return 1
	}

	return c.execCommand(command, project)
}

func (c *CLI) getProject() *application.Project {
	project, err := application.GetProject(c.CWD); HandleErr(err)
	return project
}

func (c *CLI) getPrinter() *printer {
	return &printer{
		colorsEnabled: c.ColorsEnabled,
		out:           c.Out,
	}
}

func (c *CLI) execCommand(command *application.Command, project *application.Project) int {
	cmd := exec.Command("sh", command.Path)
	cmd.Dir = project.RootFolder
	cmd.Stdout = c.Out
	cmd.Stderr = c.Err
	cmd.Stdin = c.In
	cmd.Env = append(os.Environ(), buildEnvVars(project.TskFile)...)
	err := cmd.Run()
	if err != nil {
		return 1
	}
	return 0
}


// BuildEnvVars .
func buildEnvVars(tskFile *application.TskFile) []string {
	result := make([]string, 0)
	for key, value := range tskFile.Env {
		result = append(result, key+"="+value)
	}
	return result
}

func findCommand(args []string, commands []*application.Command) *application.Command {
	if len(args) == 0 {
		return nil
	}
	commandName := args[0]
	isLast := len(args) == 1
	for _, c := range commands {
		if c.Name == commandName {
			if isLast {
				return c
			}

			return findCommand(args[1:], c.SubCommands)
		}
	}

	return nil
}
