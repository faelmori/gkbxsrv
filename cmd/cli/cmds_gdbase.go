package cli

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/logz"
	databases "github.com/faelmori/gkbxsrv/services"
	"github.com/spf13/cobra"
)

var fs databases.FilesystemService
var dbCfg *databases.DatabaseService
var defaultConfitFile string

func GdbaseCommands() []*cobra.Command {
	return []*cobra.Command{
		gdbaseCommand(),
	}
}

func gdbaseCommand() *cobra.Command {
	var configFile, username, password, dbType, host, port, database, path, dsn string
	var quiet bool

	if defaultConfitFile == "" {
		if fs == nil {
			tfs := databases.NewFileSystemService("")
			fs = *tfs
		}
		defaultConfitFile = fs.GetConfigFilePath()
	}

	var gdbaseExp = []string{
		"gkbxsrv database auth --username=foo --password=bar --host=localhost --port=5432 --database=kubex_db",
		"gkbxsrv database user add-user --username='foo' --password='bar' --name='Foo' --email='foo@bar.com'",
	}

	cmd := &cobra.Command{
		Use:     "database",
		Aliases: []string{"db", "sql", "gkbxsrv", "database"},
		Example: concatenateExamples(gdbaseExp),
		Annotations: getDescriptions([]string{
			"Local DB server and manager for many things",
			"Local DB server and manager for many things",
		}, true),
		RunE: func(cmd *cobra.Command, args []string) error {

			dbaseObj := databases.NewDatabaseService(configFile)
			_, dbaseConnErr := dbaseObj.OpenDB()
			if dbaseConnErr != nil {
				logz.Logger.Error(fmt.Sprintf("Error connecting to database: %v", dbaseConnErr), nil)
				return dbaseConnErr
			}

			dbCfg = &dbaseObj

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", defaultConfitFile, "config file")
	cmd.Flags().StringVarP(&username, "username", "u", "kubex_adm", "username")
	cmd.Flags().StringVarP(&password, "password", "p", "{auto}", "password")
	cmd.Flags().StringVarP(&dbType, "type", "t", "postgres", "database type")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "host")
	cmd.Flags().StringVarP(&port, "port", "P", "5432", "port")
	cmd.Flags().StringVarP(&database, "database", "d", "kubex_db", "database")
	cmd.Flags().StringVarP(&path, "path", "f", "", "path")
	cmd.Flags().StringVarP(&dsn, "dsn", "D", "", "data source name")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")

	if markHiddenErr := cmd.Flags().MarkHidden("quiet"); markHiddenErr != nil {
		logz.Logger.Error(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), nil)
	}
	if markHiddenErr := cmd.Flags().MarkHidden("path"); markHiddenErr != nil {
		logz.Logger.Error(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), nil)
	}
	if markHiddenErr := cmd.Flags().MarkHidden("dsn"); markHiddenErr != nil {
		logz.Logger.Error(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), nil)
	}

	cmd = AuthenticationRootCommand(cmd)
	cmd.AddCommand(UserRootCommand())
	cmd.AddCommand(RolesRootCommand())

	return cmd
}
func concatenateExamples(examplesList []string) string {
	examples := ""
	for _, example := range examplesList {
		examples += string(example) + "\n  "
	}
	return examples
}
