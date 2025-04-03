package cli

import (
	"fmt"
	databases "github.com/faelmori/gkbxsrv/services"
	"github.com/spf13/cobra"
)

func RolesRootCommand() *cobra.Command {
	rolesCmd := &cobra.Command{
		Use:         "roles",
		Aliases:     []string{"role", "r", "rol"},
		Annotations: getDescriptions([]string{"Role commands for the gospider module.", "Role commands"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("you must specify a subcommand")
		},
	}

	rolesCmd.AddCommand(rolesCommands()...)

	return rolesCmd
}

func rolesCommands() []*cobra.Command {
	insRole := insertRoleCommand()

	return []*cobra.Command{insRole}
}
func insertRoleCommand() *cobra.Command {
	insertRoleCmd := &cobra.Command{
		Use:         "insert-role",
		Aliases:     []string{"ins-role", "role"},
		Annotations: getDescriptions([]string{"Insert a new role into the database.", "Insert role"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbaseObj := databases.NewDatabaseService(configFile)
			_, dbaseConnErr := dbaseObj.OpenDB()
			if dbaseConnErr != nil {
				return dbaseConnErr
			}

			return nil
		},
	}

	insertRoleCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	insertRoleCmd.Flags().StringVarP(&username, "username", "u", "kubex_adm", "username")
	insertRoleCmd.Flags().StringVarP(&password, "password", "p", "{auto}", "password")
	insertRoleCmd.Flags().StringVarP(&dbType, "type", "t", "postgres", "database type")
	insertRoleCmd.Flags().StringVarP(&host, "host", "H", "localhost", "host")
	insertRoleCmd.Flags().StringVarP(&port, "port", "P", "5432", "port")
	insertRoleCmd.Flags().StringVarP(&database, "database", "d", "kubex_db", "database")
	insertRoleCmd.Flags().StringVarP(&path, "path", "f", "", "path")
	insertRoleCmd.Flags().StringVarP(&dsn, "dsn", "D", "", "data source name")
	insertRoleCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")

	return insertRoleCmd
}
