package cli

import (
	"fmt"
	"github.com/faelmori/gokubexfs/internal/utils"
	"github.com/spf13/cobra"
	"strings"
)

// UsersCmdsList retorna uma lista de comandos Cobra relacionados a usuários.
// Retorna um slice de ponteiros para comandos Cobra.
func UsersCmdsList() []*cobra.Command {
	cmdGetPrimaryUser := usersCmdGetPrimaryUser()
	cmdGetPrimaryGroup := usersCmdGetPrimaryGroup()
	cmdGetGroups := usersCmdGetGroups()

	return []*cobra.Command{
		cmdGetPrimaryUser,
		cmdGetPrimaryGroup,
		cmdGetGroups,
	}
}

// usersCmdGetPrimaryUser cria um comando Cobra para obter o usuário principal do sistema.
// Retorna um ponteiro para o comando Cobra configurado.
func usersCmdGetPrimaryUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Obtém o usuário principal",
		Long:  "Obtém o usuário principal do sistema",
		RunE: func(cmd *cobra.Command, args []string) error {
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			user, err := utils.GetPrimaryUser()

			if quietFlagValue {
				if err != nil {
					cmd.SilenceErrors = true
					return fmt.Errorf("")
				} else {
					fmt.Println(user)
				}
			} else {
				if err != nil {
					return err
				} else {
					fmt.Println("Primary user:", user)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}

// usersCmdGetPrimaryGroup cria um comando Cobra para obter o grupo principal do usuário no sistema.
// Retorna um ponteiro para o comando Cobra configurado.
func usersCmdGetPrimaryGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Obtém o grupo principal",
		Long:  "Obtém o grupo principal do usuário no sistema",
		RunE: func(cmd *cobra.Command, args []string) error {
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			group, err := utils.GetPrimaryGroup()

			if quietFlagValue {
				if err != nil {
					cmd.SilenceErrors = true
					return err
				}
				fmt.Println(group)
			} else {
				if err != nil {
					return err
				}
				fmt.Println("Primary group:", group)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}

// usersCmdGetGroups cria um comando Cobra para obter os grupos aos quais o usuário pertence.
// Retorna um ponteiro para o comando Cobra configurado.
func usersCmdGetGroups() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Obtém os grupos",
		Long:  "Obtém os grupos aos quais o usuário pertence",
		RunE: func(cmd *cobra.Command, args []string) error {
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			groups, err := utils.GetGroups()

			if quietFlagValue {
				if err != nil {
					cmd.SilenceErrors = true
					return err
				}
				fmt.Println(strings.Join(groups, ","))
			} else {
				if err != nil {
					return err
				}
				fmt.Println("Primary user groups:", strings.Join(groups, ","))
			}
			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}
