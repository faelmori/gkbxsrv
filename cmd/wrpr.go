package main

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/cmd/cli"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// GKBXSrv represents the structure of the utils module.
type GKBXSrv struct{}

// RegX registers and returns a new instance of GKBXSrv.
// Returns a pointer to a new instance of GKBXSrv.
func RegX() *GKBXSrv {
	return &GKBXSrv{}
}

// Alias returns the alias of the utils module.
// Returns a string containing the alias of the module.
func (m *GKBXSrv) Alias() string {
	return ""
}

// ShortDescription returns a short description of the utils module.
// Returns a string containing the short description of the module.
func (m *GKBXSrv) ShortDescription() string {
	return "Kubex Ecosystem Services module"
}

// LongDescription returns a long description of the utils module.
// Returns a string containing the long description of the module.
func (m *GKBXSrv) LongDescription() string {
	return "Kubex Ecosystem Services module"
}

// Usage returns the usage of the utils module.
// Returns a string containing the usage of the module.
func (m *GKBXSrv) Usage() string {
	return "gkbxsrv [command] [args]"
}

// Examples returns examples of usage of the utils module.
// Returns a slice of strings containing examples of usage of the module.
func (m *GKBXSrv) Examples() []string {
	return []string{"gkbxsrv utils [args]"}
}

// Active checks if the utils module is active.
// Returns a boolean indicating if the module is active.
func (m *GKBXSrv) Active() bool {
	return true
}

func (m *GKBXSrv) Module() string {
	return "gkbxsrv"
}

func (m *GKBXSrv) Execute() error {
	if err := m.Command().Execute(); err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}
	return nil
}

func (m *GKBXSrv) concatenateExamples() string {
	examples := ""
	for _, example := range m.Examples() {
		examples += string(example) + "\n  "
	}
	return examples
}

func (m *GKBXSrv) Command() *cobra.Command {
	cmd := m.utilzCmd()

	// Data command
	dataCmd := &cobra.Command{
		Use:         "data",
		Aliases:     []string{"d"},
		Short:       "Data module",
		Annotations: m.getDescriptions([]string{"Data module is a set of tools to help you manage data structures.", "Data module"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	cmdDataList := cli.DataCmdsList()
	dataCmd.AddCommand(cmdDataList...)
	cmd.AddCommand(dataCmd)

	// Paths command
	pathsCmd := &cobra.Command{
		Use:         "fs",
		Aliases:     []string{"fileSystem", "path", "pth"},
		Short:       "Paths module",
		Annotations: m.getDescriptions([]string{"Paths module is a set of tools to help you manage file paths.", "Paths module"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	pathsCmdList := cli.PathsCmdsList()
	pathsCmd.AddCommand(pathsCmdList...)
	cmd.AddCommand(pathsCmd)

	// Users command
	usersCmd := &cobra.Command{
		Use:         "users",
		Aliases:     []string{"u"},
		Short:       "Users module",
		Annotations: m.getDescriptions([]string{"Users module is a set of tools to help you manage users.", "Users module"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	usersCmdList := cli.UsersCmdsList()
	usersCmd.AddCommand(usersCmdList...)
	cmd.AddCommand(usersCmd)

	// Ports command
	portsCmd := &cobra.Command{
		Use:         "ports",
		Aliases:     []string{"p"},
		Short:       "Ports module",
		Annotations: m.getDescriptions([]string{"Ports module is a set of tools to help you manage ports.", "Ports module"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	portsCmdList := cli.PortsCmdsList()
	portsCmd.AddCommand(portsCmdList...)
	cmd.AddCommand(portsCmd)

	// SSH command
	sshCmd := &cobra.Command{
		Use:         "ssh",
		Aliases:     []string{"s"},
		Short:       "SSH module",
		Annotations: m.getDescriptions([]string{"SSH module is a set of tools to help you manage SSH connections.", "SSH module"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	sshCmdList := cli.SshCmdsList()
	sshCmd.AddCommand(sshCmdList...)
	cmd.AddCommand(sshCmd)

	// Term command
	termCmd := &cobra.Command{
		Use:         "term",
		Aliases:     []string{"t"},
		Short:       "Term module",
		Annotations: m.getDescriptions([]string{"Term module is a set of tools to help you manage terminal sessions.", "Term module"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	termCmdList := cli.TermCmdsList()
	termCmd.AddCommand(termCmdList...)
	cmd.AddCommand(termCmd)

	// Utils command
	utilsCmd := &cobra.Command{
		Use:         "utils",
		Aliases:     []string{"u"},
		Short:       "Utilities module",
		Annotations: m.getDescriptions([]string{"Utilities module is a set of tools to help you with your daily tasks.", "Utilities module"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	utilsCmd.AddCommand(cli.UtilsCmdsList()...)
	cmd.AddCommand(utilsCmd)

	cmd.AddCommand(cli.GdbaseCommands()...)
	cmd.AddCommand(cli.BrokerCommands()...)

	for _, c := range cmd.Commands() {
		setUsageDefinition(c)
	}

	setUsageDefinition(cmd)

	return cmd
}

func (m *GKBXSrv) getDescriptions(descriptionArg []string, hideBanner bool) map[string]string {
	var description, banner string
	if descriptionArg != nil {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = descriptionArg[0]
		} else {
			description = descriptionArg[1]
		}
	} else {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = m.LongDescription()
		} else {
			description = m.ShortDescription()
		}
	}
	if !hideBanner {
		banner = `   ______      __ __      __              ___________
  / ____/___  / //_/_  __/ /_  ___  _  __/ ____/ ___/
 / / __/ __ \/ ,< / / / / __ \/ _ \| |/_/ /_   \__ \ 
/ /_/ / /_/ / /| / /_/ / /_/ /  __/>  </ __/  ___/ / 
\____/\____/_/ |_\__,_/_.___/\___/_/|_/_/    /____/
`
	} else {
		banner = ""
	}
	return map[string]string{"banner": banner, "description": description}
}

func (m *GKBXSrv) utilzCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "gkbxsrv",
		Example:     m.concatenateExamples(),
		Annotations: m.getDescriptions([]string{m.LongDescription(), m.ShortDescription()}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	return cmd
}
