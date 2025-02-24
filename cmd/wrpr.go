package cmd

//import (
//	"fmt"
//	"github.com/faelmori/gokubexfs/cmd/cli"
//	"github.com/faelmori/gokubexfs/services"
//	databases "github.com/faelmori/gokubexfs/services"
//	"github.com/faelmori/kbx/mods/logz"
//	"github.com/spf13/cobra"
//	"os"
//	"strings"
//)
//
//var fs services.FilesystemService
//
//type GDBase struct {
//	dbCfg             *databases.DatabaseService
//	defaultConfitFile string
//}
//
//func (m *GDBase) Alias() string {
//	return "db"
//}
//func (m *GDBase) ShortDescription() string {
//	return "Local DB server and manager for many things"
//}
//func (m *GDBase) LongDescription() string {
//	return "Local DB server and manager for many things"
//}
//func (m *GDBase) Usage() string {
//	return "gdbase [command] [args]"
//}
//func (m *GDBase) Examples() []string {
//	return []string{
//		"kbx gokubexfs auth --username=foo --password=bar --host=localhost --port=5432 --database=kubex_db",
//		"kbx gokubexfs user add-user --username='foo' --password='bar' --name='Foo' --email='foo@bar.com'",
//	}
//}
//func (m *GDBase) Active() bool {
//	return true
//}
//func (m *GDBase) Module() string {
//	return "gdbase"
//}
//func (m *GDBase) Execute() error {
//	rootCmd := m.Command()
//	return rootCmd.Execute()
//}
//func (m *GDBase) Command() *cobra.Command {
//	var configFile, username, password, dbType, host, port, database, path, dsn string
//	var quiet bool
//
//	cmd := &cobra.Command{
//		Use:         m.Module(),
//		Aliases:     []string{m.Alias(), "db", "sql", "gokubexfs", "database"},
//		Example:     m.concatenateExamples(),
//		Annotations: m.getDescriptions(nil, true),
//		RunE: func(cmd *cobra.Command, args []string) error {
//			dbaseObj := databases.NewDatabaseService(configFile)
//			_, dbaseConnErr := dbaseObj.OpenDB()
//			if dbaseConnErr != nil {
//				return logz.ErrorLog(fmt.Sprintf("Error connecting to database: %v", dbaseConnErr), "GoSpyder")
//			}
//
//			m.dbCfg = &dbaseObj
//
//			return nil
//		},
//	}
//
//	cmd.Flags().StringVarP(&configFile, "config", "c", m.defaultConfitFile, "config file")
//	cmd.Flags().StringVarP(&username, "username", "u", "kubex_adm", "username")
//	cmd.Flags().StringVarP(&password, "password", "p", "{auto}", "password")
//	cmd.Flags().StringVarP(&dbType, "type", "t", "postgres", "database type")
//	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "host")
//	cmd.Flags().StringVarP(&port, "port", "P", "5432", "port")
//	cmd.Flags().StringVarP(&database, "database", "d", "kubex_db", "database")
//	cmd.Flags().StringVarP(&path, "path", "f", "", "path")
//	cmd.Flags().StringVarP(&dsn, "dsn", "D", "", "data source name")
//	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
//
//	if markHiddenErr := cmd.Flags().MarkHidden("quiet"); markHiddenErr != nil {
//		_ = logz.ErrorLog(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), "GoSpyder")
//	}
//	if markHiddenErr := cmd.Flags().MarkHidden("path"); markHiddenErr != nil {
//		_ = logz.ErrorLog(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), "GoSpyder")
//	}
//	if markHiddenErr := cmd.Flags().MarkHidden("dsn"); markHiddenErr != nil {
//		_ = logz.ErrorLog(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), "GoSpyder")
//	}
//
//	cmd = cli.AuthenticationRootCommand(cmd)
//	cmd.AddCommand(cli.UserRootCommand())
//	cmd.AddCommand(cli.RolesRootCommand())
//
//	setUsageDefinition(cmd)
//
//	return cmd
//}
//
//func RegX() *GDBase {
//	if fs == nil {
//		fsSrv := services.NewFileSystemService("")
//		fs = *fsSrv
//	}
//
//	defaultConfigFile := fs.GetConfigFilePath()
//
//	return &GDBase{
//		defaultConfitFile: defaultConfigFile,
//	}
//}
//func (m *GDBase) concatenateExamples() string {
//	examples := ""
//	for _, example := range m.Examples() {
//		examples += string(example) + "\n  "
//	}
//	return examples
//}
//
//func (m *GDBase) getDescriptions(descriptionArg []string, _ bool) map[string]string {
//	var description, banner string
//	if descriptionArg != nil {
//		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
//			description = descriptionArg[0]
//		} else {
//			description = descriptionArg[1]
//		}
//	} else {
//		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
//			description = m.LongDescription()
//		} else {
//			description = m.ShortDescription()
//		}
//	}
//	//if !hideBanner {
//	banner = ` ____       ____  ____
// / ___|     |  _ \| __ )  __ _ ___  ___
//| |  _ _____| | | |  _ \ / _| / __|/ _ \
//| |_| |_____| |_| | |_) | (_| \__ \  __/
// \____|     |____/|____/ \__,_|___/\___|
//`
//	//} else {
//	//banner = ""
//	//}
//	return map[string]string{"banner": banner, "description": description}
//}
