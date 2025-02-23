package cmd

import (
	"fmt"
	"github.com/faelmori/gokubexfs/cmd/cli"
	"github.com/spf13/cobra"
	"os"
	"reflect"
	"strings"
)

// Utilz representa a estrutura do módulo utils.
type Utilz struct{}

// RegX registra e retorna uma nova instância de Utilz.
// Retorna um ponteiro para uma nova instância de Utilz.
func RegX() *Utilz {
	return &Utilz{}
}

// Alias retorna o alias do módulo utils.
// Retorna uma string contendo o alias do módulo.
func (m *Utilz) Alias() string {
	return "utilz"
}

// ShortDescription retorna uma descrição curta do módulo utils.
// Retorna uma string contendo a descrição curta do módulo.
func (m *Utilz) ShortDescription() string {
	return "Utilities module"
}

// LongDescription retorna uma descrição longa do módulo utils.
// Retorna uma string contendo a descrição longa do módulo.
func (m *Utilz) LongDescription() string {
	return "Utilities module is a set of tools to help you with your daily tasks."
}

// Usage retorna a forma de uso do módulo utils.
// Retorna uma string contendo a forma de uso do módulo.
func (m *Utilz) Usage() string {
	return "kbx utils [command] [args]"
}

// Examples retorna exemplos de uso do módulo utils.
// Retorna um slice de strings contendo exemplos de uso do módulo.
func (m *Utilz) Examples() []string {
	return []string{"kbx util [command] [args]", "kbx utils getOrCreate [args]"}
}

// Active verifica se o módulo utils está ativo.
// Retorna um booleano indicando se o módulo está ativo.
func (m *Utilz) Active() bool {
	return true
}

func (m *Utilz) Module() string {
	return "utils"
}

func (m *Utilz) Execute() error {
	cmd := m.Command()
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("erro ao executar o comando: %w", err)
	}
	return nil
}

func (m *Utilz) concatenateExamples() string {
	examples := ""
	for _, example := range m.Examples() {
		examples += string(example) + "\n  "
	}
	return examples
}

func (m *Utilz) Command() *cobra.Command {
	cmd := utilzCmd(m)

	// Comando data
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

	// Comando path
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

	// Comando users
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

	// Comando ports
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

	// Comando ssh
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

	// Comando term
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

	// Comando utils
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

	for _, c := range cmd.Commands() {
		setUsageDefinition(c)
	}

	setUsageDefinition(cmd)

	return cmd
}

func (m *Utilz) getDescriptions(descriptionArg []string, hideBanner bool) map[string]string {
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

func (m *Utilz) utilzCmdsList() ([]*cobra.Command, error) {
	//cmdUtilz := utilzCmd(m)

	return []*cobra.Command{
		//cmdUtilz,
	}, nil
}

func (m *Utilz) utilzCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "utils",
		Aliases:     []string{m.Alias(), "util", "utl", "u"},
		Example:     m.concatenateExamples(),
		Annotations: m.getDescriptions([]string{m.LongDescription(), m.ShortDescription()}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Verifica e executa as flags
			cmdFlag, _ := cmd.Flags().GetString("cmd")
			argsFlag, _ := cmd.Flags().GetStringArray("args")
			//quietFlag, _ := cmd.Flags().GetBool("quiet")

			newArgs := argsFlag
			newArgs = append(newArgs, args...)
			//quietFlagStr := ""
			//if quietFlag {
			//	quietFlagStr = "quiet"
			//}

			if cmdFlag != "" {
				switch cmdFlag {
				case "users":
					return nil
				case "data":
					return nil
				case "paths":
					return nil
				case "ports":
					return nil
				default:
					return fmt.Errorf("command not found")
				}
			}

			return fmt.Errorf("invalid command or flag")
		},
	}

	// Flags para deps
	cmd.Flags().StringP("cmd", "c", "", "Executa o comando especificado")
	cmd.Flags().StringArrayP("args", "a", []string{}, "Argumentos para o comando especificado")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso")

	return cmd
}

func utilzCmd(m *Utilz) *cobra.Command {
	cmd := &cobra.Command{
		Use:         m.Module(),
		Aliases:     []string{"util", "utils"},
		Annotations: m.getDescriptions([]string{m.LongDescription(), m.ShortDescription()}, os.Getenv("KBX_HIDE_BANNER") == "true"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}
	return cmd
}

func getDepsList() ([]string, error) {
	if len(os.Args) == 0 {
		return nil, fmt.Errorf("nenhuma dependência informada")
	}
	for i, dep := range os.Args {
		if reflect.TypeOf(dep).String() == "[]string" {
			return os.Args[i+1:], nil
		}
	}
	return nil, fmt.Errorf("nenhuma dependência informada")
}
