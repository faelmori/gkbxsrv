package cli

import (
	"fmt"
	"github.com/faelmori/gokubexfs/internal/models/forms"
	. "github.com/faelmori/gokubexfs/models"
	"github.com/faelmori/gokubexfs/services"
	databases "github.com/faelmori/gokubexfs/services"
	"github.com/faelmori/kbx/mods/logz"

	. "github.com/faelmori/kbx/mods/ui/components"
	"github.com/faelmori/kbx/mods/ui/types"
	"github.com/faelmori/kbx/mods/utils"
	"github.com/spf13/cobra"
)

func UserRootCommand() *cobra.Command {
	userCmd := &cobra.Command{
		Use:         "user",
		Aliases:     []string{"users", "u", "usr"},
		Annotations: getDescriptions([]string{"User commands for the gospyder module.", "User commands"}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return logz.ErrorLog("No command specified for users.", "gospyder")
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
				return logz.ErrorLog(fmt.Sprintf("Failed to initialize the database: %s", gDBRepoConnErr.Error()), "GDBase")
			} else {
				_ = logz.InfoLog(fmt.Sprintf("Database initialized: %s", dbB.Name()), "GDBase")
			}
			userR := NewUserRepo(&dbB)
			userRepo := *userR

			var user *User
			if username == "" || email == "" || password == "" || name == "" {
				utils.ClearScreen()
				_ = logz.InfoLog("No user data provided, starting wizard...", "GDBase")
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
			} else {
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
			}

			if validateErr := user.Validate(); validateErr != nil {
				return logz.ErrorLog(fmt.Sprintf("Failed to validate user: %s", validateErr.Error()), "GDBase")
			}

			dckr := services.NewDockerService()
			if dckr.IsDockerRunning() {
				if dbErr := dckr.SetupDatabaseServices(); dbErr != nil {
					return logz.ErrorLog(fmt.Sprintf("Failed to setup database services: %s", dbErr.Error()), "GDBase")
				}
			}
			if createdUser, createdUserErr := userRepo.Create(user); createdUserErr != nil || createdUser == nil {
				return logz.ErrorLog(fmt.Sprintf("Failed to create user: %s", createdUserErr.Error()), "GDBase")
			} else {
				_ = logz.InfoLog(fmt.Sprintf("User created: %s", createdUser.GetID()), "GDBase")
				if outputTarget != "" {
					fsSrv := services.NewFileSystemService(configFile)
					fs := *fsSrv
					writeErr := fs.WriteToFile(outputTarget, createdUser, &outputType)
					if writeErr != nil {
						return logz.ErrorLog(fmt.Sprintf("Failed to write user to file: %s", writeErr.Error()), "GDBase")
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
		_ = logz.ErrorLog(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), "GoSpyder")
	}
	if markHiddenErr := insertUserCmd.Flags().MarkHidden("role"); markHiddenErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), "GoSpyder")
	}
	if markHiddenErr := insertUserCmd.Flags().MarkHidden("avatar"); markHiddenErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), "GoSpyder")
	}
	if markHiddenErr := insertUserCmd.Flags().MarkHidden("data"); markHiddenErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Error marking flag as hidden: %v", markHiddenErr), "GoSpyder")
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
					return logz.ErrorLog(fmt.Sprintf("Failed to setup database services: %s", dbErr.Error()), "GDBase")
				}
			}
			gDBRepo := databases.NewDatabaseService(configFile)
			gDBRepoConn, gDBRepoConnErr := gDBRepo.OpenDB()
			if gDBRepoConnErr != nil {
				return logz.ErrorLog(fmt.Sprintf("Failed to initialize the database: %s", gDBRepoConnErr.Error()), "GDBase")
			}
			userRepo := NewUserRepo(gDBRepoConn)
			if userRepo != nil {
				if users, usersErr := userRepo.List(where); usersErr != nil {
					return logz.ErrorLog(fmt.Sprintf("Failed to find users: %s", usersErr.Error()), "GDBase")
				} else {
					if len(users.GetRows()) < 1 {
						return logz.ErrorLog(fmt.Sprintf("No users found with selected criteria (%p): %s", where, usersErr.Error()), "GDBase")
					}
					return StartTableScreen(&users, nil)
				}
			} else {
				return logz.ErrorLog("Failed to get user repository", "GDBase")
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
