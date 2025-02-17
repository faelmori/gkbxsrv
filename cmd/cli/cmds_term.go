package cli

import (
	"fmt"
	"github.com/faelmori/gokubexfs/internal/utils"
	"github.com/spf13/cobra"
)

// TermCmdsList retorna uma lista de comandos Cobra relacionados ao terminal.
// Retorna um slice de ponteiros para comandos Cobra.
func TermCmdsList() []*cobra.Command {
	figCmd := termFigletCmd()

	return []*cobra.Command{
		figCmd,
	}
}

// termFigletCmd cria um comando Cobra para exibir um texto como um banner.
// Retorna um ponteiro para o comando Cobra configurado.
func termFigletCmd() *cobra.Command {
	var figletTitle string

	fCmd := &cobra.Command{
		Use:     "figlet",
		Aliases: []string{"banner", "bannerText", "fig"},
		Short:   "Exibe um texto como um banner",
		Long:    "Exibe um texto como um banner usando o comando figlet",
		RunE: func(cmd *cobra.Command, args []string) error {
			if figletTitle == "" {
				return fmt.Errorf("é necessário fornecer o texto")
			}
			fErr := utils.Figlet(figletTitle)
			if fErr != nil {
				return fmt.Errorf("erro ao exibir o texto: %v", fErr)
			}
			return nil
		},
	}

	// Define a flag para o texto a ser exibido como um banner
	fCmd.Flags().StringVarP(&figletTitle, "text", "t", "", "Texto a ser exibido como um banner")

	return fCmd
}
