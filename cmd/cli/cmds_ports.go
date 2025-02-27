package cli

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/utils"
	"github.com/spf13/cobra"
)

// PortsCmdsList retorna uma lista de comandos Cobra relacionados a portas.
// Retorna um slice de ponteiros para comandos Cobra.
func PortsCmdsList() []*cobra.Command {
	cmdCheckPortOpen := portsCmdCheckPortOpen()
	cmdListOpenPorts := portsCmdListOpenPorts()
	cmdClosePort := portsCmdClosePort()
	cmdOpenPort := portsCmdOpenPort()

	return []*cobra.Command{
		cmdCheckPortOpen,
		cmdListOpenPorts,
		cmdClosePort,
		cmdOpenPort,
	}
}

// portsCmdCheckPortOpen cria um comando Cobra para verificar se uma porta está aberta.
// Retorna um ponteiro para o comando Cobra configurado.
func portsCmdCheckPortOpen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-open",
		Short: "Verifica se uma porta está aberta",
		Long:  "Verifica se uma porta específica está aberta no localhost",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("é necessário fornecer o número da porta")
			}
			port := args[0]

			open, err := utils.CheckPortOpen(port)
			if err != nil {
				return err
			}
			fmt.Printf("A porta %s está aberta? %t\n", port, open)
			return nil
		},
	}

	return cmd
}

// portsCmdListOpenPorts cria um comando Cobra para listar todas as portas abertas.
// Retorna um ponteiro para o comando Cobra configurado.
func portsCmdListOpenPorts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-open",
		Short: "Lista todas as portas abertas",
		Long:  "Lista todas as portas abertas no localhost",
		RunE: func(cmd *cobra.Command, args []string) error {
			ports, err := utils.ListOpenPorts()
			if err != nil {
				return err
			}
			fmt.Println("Portas abertas:", ports)
			return nil
		},
	}

	return cmd
}

// portsCmdClosePort cria um comando Cobra para fechar uma porta específica.
// Retorna um ponteiro para o comando Cobra configurado.
func portsCmdClosePort() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Fecha uma porta específica",
		Long:  "Fecha uma porta específica no localhost",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("é necessário fornecer o número da porta")
			}
			port := args[0]

			err := utils.ClosePort(port)
			if err != nil {
				return err
			}
			fmt.Printf("A porta %s foi fechada\n", port)
			return nil
		},
	}

	return cmd
}

// portsCmdOpenPort cria um comando Cobra para abrir uma porta específica.
// Retorna um ponteiro para o comando Cobra configurado.
func portsCmdOpenPort() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Abre uma porta específica",
		Long:  "Abre uma porta específica no localhost",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("é necessário fornecer o número da porta")
			}
			port := args[0]

			err := utils.OpenPort(port)
			if err != nil {
				return err
			}
			fmt.Printf("A porta %s foi aberta\n", port)
			return nil
		},
	}

	return cmd
}
