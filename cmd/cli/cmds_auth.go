package cli

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/utils"
	databases "github.com/faelmori/kbxutils/api"
	l "github.com/faelmori/logz"
	"github.com/faelmori/xtui/types"
	"reflect"

	. "github.com/faelmori/xtui/components"
	"github.com/spf13/cobra"
	"strings"
)

func AuthenticationRootCommand(cmd *cobra.Command) *cobra.Command {
	authCmd := &cobra.Group{
		ID:    "auth",
		Title: "Authentication Commands",
	}
	cmd.AddGroup(authCmd)
	cmd.AddCommand(authCommands()...)
	return cmd
}

func authCommands() []*cobra.Command {
	authUser := authenticateUserCommand()

	return []*cobra.Command{authUser}
}
func authenticateUserCommand() *cobra.Command {

	authenticateUserCmd := &cobra.Command{
		GroupID:     "auth",
		Use:         "auth",
		Aliases:     []string{"authenticate", "login", "signin"},
		Annotations: getDescriptions([]string{"Authenticate user to database.", "Authenticate user"}, false),
		Run: func(cmd *cobra.Command, args []string) {
			if username == "" || password == "" {
				dbHost := ""
				if host != "" {
					dbHost = host
				} else if dsn != "" {
					dbHost = strings.Split(dsn, "@")[1]
				}
				if !quiet {
					utils.ClearScreen()
					inputs := make([]types.FormInputObject[any], 4)
					//data := reflect.ValueOf(&database).Elem().Interface().(string)

					inputs[0] = &types.Input[any]{
						Ph:                 "Database",
						Tp:                 reflect.TypeOf(""),
						Val:                nil,
						Req:                true,
						Min:                3,
						Max:                50,
						Err:                "Database é obrigatório",
						ValidationRulesVal: nil, /*[]types.FormInputValidationString{
						"TESTE",
						"TESTE2",
						func(s string) error {
							if len(s) < 3 {
								return fmt.Errorf("database deve ter no mínimo 3 caracteres")
							}
							return nil
						},*/
					}
					inputs[1] = &types.Input[any]{
						Ph:                 "Porta",
						Tp:                 reflect.TypeOf(""),
						Val:                nil,
						Req:                true,
						Min:                3,
						Max:                5,
						Err:                "Porta é obrigatória",
						ValidationRulesVal: nil, /*func(s string) error {
							p, pErr := strconv.Atoi(s)
							if pErr != nil {
								return fmt.Errorf("porta deve ser um número")
							}
							if p < 100 || p > 65535 {
								return fmt.Errorf("porta deve ser um número entre 100 e 65535")
							}
							return nil
						},*/
					}
					inputs[2] = &types.Input[any]{
						Ph:                 "Nome de Usuário",
						Tp:                 reflect.TypeOf(""),
						Val:                nil,
						Req:                true,
						Min:                3,
						Max:                50,
						Err:                "Nome de Usuário é obrigatório",
						ValidationRulesVal: nil, /*func(s string) error {
							if len(s) < 3 {
								return fmt.Errorf("nome de Usuário deve ter no mínimo 3 caracteres")
							}
							return nil
						}*/
					}
					inputs[3] = &types.Input[any]{
						Ph:                 "Senha",
						Tp:                 reflect.TypeOf(""),
						Val:                nil,
						Req:                true,
						Min:                6,
						Max:                50,
						Err:                "Senha é obrigatória",
						ValidationRulesVal: nil, /*func(s string) error {
							if len(s) < 6 {
								return fmt.Errorf("senha deve ter no mínimo 6 caracteres")
							}
							return nil
						},*/
					}

					var fields = types.FormFields{Title: "text", Fields: inputs}
					tuizResult, tuizErr := ShowForm(types.FormConfig{
						Title:      fmt.Sprintf("Autenticar Usuário %s", dbHost),
						FormFields: fields,
					})
					if tuizErr != nil {
						l.Error(fmt.Sprintf("Error: %v", tuizErr), map[string]any{
							"context": "tuiz",
							"error":   tuizErr,
						})
					}
					username = tuizResult["field0"]
					password = tuizResult["field2"]
				} else {
					l.Error("Username and password are required", map[string]any{
						"context": "tuiz",
					})
					return
				}
			}
			dbaseObj := databases.NewDatabaseService(configFile)
			_, dbaseConnErr := dbaseObj.OpenDB()
			if dbaseConnErr != nil {
				l.Error("Error connecting to database", map[string]any{
					"context": "tuiz",
					"error":   dbaseConnErr,
				})
				return
			}
		},
	}

	authenticateUserCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	authenticateUserCmd.Flags().StringVarP(&username, "username", "u", "kubex_adm", "username")
	authenticateUserCmd.Flags().StringVarP(&password, "password", "p", "{auto}", "password")
	authenticateUserCmd.Flags().StringVarP(&dbType, "type", "t", "postgres", "database type")
	authenticateUserCmd.Flags().StringVarP(&host, "host", "H", "localhost", "host")
	authenticateUserCmd.Flags().StringVarP(&port, "port", "P", "5432", "port")
	authenticateUserCmd.Flags().StringVarP(&database, "database", "d", "kubex_db", "database")
	authenticateUserCmd.Flags().StringVarP(&path, "path", "f", "", "path")
	authenticateUserCmd.Flags().StringVarP(&dsn, "dsn", "D", "", "data source name")
	authenticateUserCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	authenticateUserCmd.Flags().StringVarP(&outputType, "output-type", "o", "json", "The output type for the user")
	authenticateUserCmd.Flags().StringVarP(&outputTarget, "output-target", "T", "", "The output target for the user")

	if markHiddenErr := authenticateUserCmd.Flags().MarkHidden("quiet"); markHiddenErr != nil {
		fmt.Printf("Error marking flag as hidden: %v", markHiddenErr)
	}
	if markHiddenErr := authenticateUserCmd.Flags().MarkHidden("path"); markHiddenErr != nil {
		fmt.Printf("Error marking flag as hidden: %v", markHiddenErr)
	}
	if markHiddenErr := authenticateUserCmd.Flags().MarkHidden("dsn"); markHiddenErr != nil {
		fmt.Printf("Error marking flag as hidden: %v", markHiddenErr)
	}

	return authenticateUserCmd
}
