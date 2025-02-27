package cli

import (
	"fmt"
	//"github.com/faelmori/gkbxsrv/internal/models/forms"
	. "github.com/faelmori/gkbxsrv/models"
	"github.com/faelmori/gkbxsrv/services"
	databases "github.com/faelmori/gkbxsrv/services"
	//"github.com/faelmori/gkbxsrv/utils"
	//. "github.com/faelmori/kbx/mods/ui/components"
	//"github.com/faelmori/kbx/mods/ui/types"
	"github.com/spf13/cobra"
)

func UserRootCommand() *cobra.Command {
	userCmd := &cobra.Command{
		Use:         "user",
		Aliases:     []string{"users", "u", "usr"},
		Annotations: getDescriptions([]string{"User commands for the gospyder module.", "User commands"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("no command specified for users")
		},
	}

	userCmd.AddCommand(userCommands()...)

	return userCmd
}

func userCommands() []*cobra.Command {
	viewUser := viewUserCommand()
	insUser := insertUserCommand()

	return []*cobra.Command{viewUser, insUser}
}
func insertUserCommand() *cobra.Command {
	var userDataMap = make(map[string]any)
	var userDataMapStr = make(map[string]string)

	insertUserCmd := &cobra.Command{
		Use:     "add-user",
		Aliases: []string{"insert-user", "ins-user", "insUser", "addUser"},
		Short:   "Insert a new user into the database",
		Long:    "Insert a new user into the database using a wizard if no arguments are provided",
		RunE: func(cmd *cobra.Command, args []string) error {
			userMap := make(map[string]any)

			gDBRepo := databases.NewDatabaseService(configFile)
			gDBRepoConn, gDBRepoConnErr := gDBRepo.OpenDB()
			dbB := *gDBRepoConn
			if gDBRepoConnErr != nil {
				return fmt.Errorf(fmt.Sprintf("Failed to initialize the database: %s", gDBRepoConnErr.Error()))
			}
			userR := NewUserRepo(&dbB)
			userRepo := *userR

			var user *User
			/*if username == "" || email == "" || password == "" || name == "" {
				utils.ClearScreen()
				userForm := forms.NewUserForm()
				userFields := userForm.GetFields(userDataMapStr)
				kbdzInputsData, kbdzInputsErr := KbdzInputs(
					types.TuizConfigz{
						Tt: "Inserting New User - Wizard",
						Fds: &types.TuizFields{
							Tt:  "User Data",
							Fds: userFields,
						},
					},
				)
				if kbdzInputsErr != nil {
					return logz.ErrorLog(fmt.Sprintf("Failed to get user data: %s", kbdzInputsErr.Error()), "GDBase")
				}
				for key, val := range kbdzInputsData {
					userMap[key] = val
				}
				user = UserFactory(userMap)
			} else {*/
			if len(userDataMap) > 0 {
				userMap = userDataMap
			} else {
				userMap = map[string]any{
					"username": username,
					"email":    email,
					"password": password,
					"name":     name,
					"roles":    role,
					"phone":    phone,
					"document": document,
					"address":  address,
					"city":     city,
					"state":    state,
					"country":  country,
					"zip":      zip,
					"avatar":   avatar,
					"birth":    birth,
				}
			}
			user = UserFactory(userMap)
			//}

			if validateErr := user.Validate(); validateErr != nil {
				return validateErr
			}

			dckr := services.NewDockerService()
			if dckr.IsDockerRunning() {
				if dbErr := dckr.SetupDatabaseServices(); dbErr != nil {
					return dbErr
				}
			}
			if createdUser, createdUserErr := userRepo.Create(user); createdUserErr != nil || createdUser == nil {
				return createdUserErr
			} else {
				if outputTarget != "" {
					fsSrv := services.NewFileSystemService(configFile)
					fs := *fsSrv
					writeErr := fs.WriteToFile(outputTarget, createdUser, &outputType)
					if writeErr != nil {
						return writeErr
					}
				}
			}
			return nil
		},
	}

	insertUserCmd.Flags().StringP("file", "f", "", "The datasource name for the database to initialize")
	insertUserCmd.Flags().StringVarP(&username, "username", "u", "", "The username for the user")
	insertUserCmd.Flags().StringVarP(&email, "email", "e", "", "The email for the user")
	insertUserCmd.Flags().StringVarP(&password, "password", "p", "", "The password for the user")
	insertUserCmd.Flags().StringVarP(&name, "name", "n", "", "The name for the user")
	insertUserCmd.Flags().StringVarP(&role, "role", "r", "", "The role for the user")
	insertUserCmd.Flags().StringVarP(&birth, "birth", "b", "", "The birth for the user")
	insertUserCmd.Flags().StringVarP(&phone, "phone", "t", "", "The phone for the user")
	insertUserCmd.Flags().StringVarP(&document, "document", "d", "", "The document for the user")
	insertUserCmd.Flags().StringVarP(&address, "address", "a", "", "The address for the user")
	insertUserCmd.Flags().StringVarP(&city, "city", "c", "", "The city for the user")
	insertUserCmd.Flags().StringVarP(&state, "state", "s", "", "The state for the user")
	insertUserCmd.Flags().StringVarP(&country, "country", "o", "", "The country for the user")
	insertUserCmd.Flags().StringVarP(&zip, "zip", "z", "", "The zip for the user")
	insertUserCmd.Flags().StringVarP(&avatar, "avatar", "v", "", "The avatar for the user")
	insertUserCmd.Flags().StringVarP(&outputType, "output-type", "O", "json", "The output type for the user")
	insertUserCmd.Flags().StringVarP(&outputTarget, "output-target", "T", "", "The output target for the user")
	insertUserCmd.Flags().StringToStringVarP(&userDataMapStr, "data", "x", nil, "The data for the user")
	insertUserCmd.Flags().BoolP("quiet", "q", false, "Quiet mode")

	if markHiddenErr := insertUserCmd.Flags().MarkHidden("quiet"); markHiddenErr != nil {
		fmt.Printf("Error marking flag as hidden: %v", markHiddenErr)
	}
	if markHiddenErr := insertUserCmd.Flags().MarkHidden("role"); markHiddenErr != nil {
		fmt.Printf(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr))
	}
	if markHiddenErr := insertUserCmd.Flags().MarkHidden("avatar"); markHiddenErr != nil {
		fmt.Printf(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr))
	}
	if markHiddenErr := insertUserCmd.Flags().MarkHidden("data"); markHiddenErr != nil {
		fmt.Printf("Error marking flag as hidden: %v", markHiddenErr)
	}

	return insertUserCmd
}
func viewUserCommand() *cobra.Command {
	var where = make(map[string]string)

	viewUsersCmd := &cobra.Command{
		Use:     "list-users",
		Aliases: []string{"view-users", "lu"},
		Short:   "View users in the database",
		Long:    "View users in the database using a terminal table screen",
		RunE: func(cmd *cobra.Command, args []string) error {
			dckr := services.NewDockerService()
			if dckr.IsDockerRunning() {
				if dbErr := dckr.SetupDatabaseServices(); dbErr != nil {
					return dbErr
				}
			}
			gDBRepo := databases.NewDatabaseService(configFile)
			gDBRepoConn, gDBRepoConnErr := gDBRepo.OpenDB()
			if gDBRepoConnErr != nil {
				return gDBRepoConnErr
			}
			userRepo := NewUserRepo(gDBRepoConn)
			if userRepo != nil {
				if users, usersErr := userRepo.List(where); usersErr != nil {
					return usersErr
				} else {
					if len(users.GetRows()) < 1 {
						return fmt.Errorf("no users found")
					}
					//return StartTableScreen(&users, nil)
					return nil
				}
			} else {
				return fmt.Errorf("failed to initialize the user repository")
			}
		},
	}

	viewUsersCmd.Flags().StringP("file", "f", "", "The datasource name for the database to initialize")
	viewUsersCmd.Flags().StringVarP(&outputType, "output-type", "O", "json", "The output type for the user")
	viewUsersCmd.Flags().StringVarP(&outputTarget, "output-target", "T", "", "The output target for the user")
	viewUsersCmd.Flags().StringToStringVarP(&where, "where", "w", map[string]string{"active": "true"}, "The where clause for the user")
	viewUsersCmd.Flags().BoolP("quiet", "q", false, "Quiet mode")

	return viewUsersCmd
}
