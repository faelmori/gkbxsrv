package cli

import (
	"fmt"
	"github.com/faelmori/gokubexfs/internal/utils"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

// PathsCmdsList retorna uma lista de comandos Cobra relacionados a caminhos.
// Retorna um slice de ponteiros para comandos Cobra.
func PathsCmdsList() []*cobra.Command {
	cmdEnsureDir := pathsCmdEnsureDir()
	cmdEnsureFile := pathsCmdEnsureFile()
	cmdSanitizePath := pathsCmdSanitizePath()
	cmdGetTempDir := pathsCmdGetTempDir()
	cmdEnsureTempDir := pathsCmdEnsureTempDir()
	cmdWatchKubexFiles := watchKubexFiles()

	return []*cobra.Command{
		cmdEnsureDir,
		cmdEnsureFile,
		cmdSanitizePath,
		cmdGetTempDir,
		cmdEnsureTempDir,
		cmdWatchKubexFiles,
	}
}

// pathsCmdEnsureDir cria um comando Cobra para garantir a existência de um diretório.
// Retorna um ponteiro para o comando Cobra configurado.
func pathsCmdEnsureDir() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ensure-folder",
		Aliases: []string{"ensure-dir", "ensure-directory", "ensure-folder", "ensure-fold", "ensure-foldr", "ensureDir", "ensureDirectory", "ensureFolder", "ensureFold", "ensureFoldr", "ensureFldr"},
		Short:   "Garante a existência de um diretório, criando-o se necessário",
		Long:    "Garante a existência de um diretório, criando-o se necessário e aplicando as permissões e usuário/grupo fornecidos",
		RunE: func(cmd *cobra.Command, args []string) error {
			pathFlagValue, _ := cmd.Flags().GetString("path")
			permFlagValue, _ := cmd.Flags().GetString("perm")
			userFlagValue, _ := cmd.Flags().GetString("user")
			groupFlagValue, _ := cmd.Flags().GetString("group")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var path, perm, user, group string
			var userGroup []string
			if pathFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer o caminho do diretório")
				}
				path = args[0]
			} else {
				path = pathFlagValue
			}
			perm = "0777"
			if permFlagValue != "" {
				perm = permFlagValue
			}
			user = ""
			if userFlagValue != "" {
				user = userFlagValue
			}
			group = ""
			if groupFlagValue != "" {
				group = groupFlagValue
			}
			if user != "" && group == "" {
				return fmt.Errorf("é necessário fornecer o grupo do diretório")
			} else if user == "" && group != "" {
				return fmt.Errorf("é necessário fornecer o usuário do diretório")
			} else {
				userGroup = []string{user, group}
			}
			fileModPerm, _ := strconv.ParseUint(perm, 8, 32)

			if quietFlagValue {
				return utils.EnsureDir(path, os.FileMode(fileModPerm), userGroup)
			}

			err := utils.EnsureDir(path, os.FileMode(fileModPerm), userGroup)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Directory ensured")
			}

			return nil
		},
	}

	cmd.Flags().StringP("path", "p", "", "Caminho do diretório")
	cmd.Flags().StringP("perm", "m", "0777", "Permissões do diretório")
	cmd.Flags().StringP("user", "u", "", "Usuário do diretório")
	cmd.Flags().StringP("group", "g", "", "Grupo do diretório")
	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}

// pathsCmdEnsureFile cria um comando Cobra para garantir a existência de um arquivo.
// Retorna um ponteiro para o comando Cobra configurado.
func pathsCmdEnsureFile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure-file",
		Short: "Garante a existência de um arquivo",
		Long:  "Garante a existência de um arquivo, criando-o se necessário",
		RunE: func(cmd *cobra.Command, args []string) error {
			pathFlagValue, _ := cmd.Flags().GetString("path")
			permFlagValue, _ := cmd.Flags().GetString("perm")
			userFlagValue, _ := cmd.Flags().GetString("user")
			groupFlagValue, _ := cmd.Flags().GetString("group")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var path, perm, user, group string
			var userGroup []string
			if pathFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer o caminho do diretório")
				}
				path = args[0]
			} else {
				path = pathFlagValue
			}
			perm = "0777"
			if permFlagValue != "" {
				perm = permFlagValue
			}
			user = ""
			if userFlagValue != "" {
				user = userFlagValue
			}
			group = ""
			if groupFlagValue != "" {
				group = groupFlagValue
			}
			if user != "" && group == "" {
				return fmt.Errorf("é necessário fornecer o grupo do diretório")
			} else if user == "" && group != "" {
				return fmt.Errorf("é necessário fornecer o usuário do diretório")
			} else {
				userGroup = []string{user, group}
			}
			fileModPerm, _ := strconv.ParseUint(perm, 8, 32)

			if quietFlagValue {
				return utils.EnsureFile(path, os.FileMode(fileModPerm), userGroup)
			}

			err := utils.EnsureFile(path, os.FileMode(fileModPerm), userGroup)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("File ensured")
			}

			return nil
		},
	}

	cmd.Flags().StringP("path", "p", "", "Caminho do diretório")
	cmd.Flags().StringP("perm", "m", "0777", "Permissões do diretório")
	cmd.Flags().StringP("user", "u", "", "Usuário do diretório")
	cmd.Flags().StringP("group", "g", "", "Grupo do diretório")
	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}

// pathsCmdSanitizePath cria um comando Cobra para sanitizar um caminho.
// Retorna um ponteiro para o comando Cobra configurado.
func pathsCmdSanitizePath() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sanitize",
		Short: "Sanitiza um caminho",
		Long:  "Sanitiza um caminho fornecido, garantindo que ele esteja dentro do caminho base sem injetar caracteres maliciosos",
		RunE: func(cmd *cobra.Command, args []string) error {
			basePathFlagValue, _ := cmd.Flags().GetString("basePath")
			baseNameFlagValue, _ := cmd.Flags().GetString("baseName")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var basePath, baseName string
			if basePathFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer o caminho base")
				}
				basePath = args[0]
			} else {
				basePath = basePathFlagValue
			}
			if baseNameFlagValue != "" {
				baseName = baseNameFlagValue
			}

			sanitizedPath, err := utils.SanitizePath(basePath, baseName)
			if quietFlagValue {
				if err != nil {
					return nil
				} else {
					fmt.Printf("%s\n", sanitizedPath)
				}
			} else {
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Path sanitized:", sanitizedPath)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringP("basePath", "p", "", "Caminho base ou absoluto")
	cmd.Flags().StringP("baseName", "n", "", "Nome base (opcional)")
	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}

// pathsCmdGetTempDir cria um comando Cobra para obter o diretório temporário da aplicação.
// Retorna um ponteiro para o comando Cobra configurado.
func pathsCmdGetTempDir() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-temp",
		Aliases: []string{"tempdir", "tempDir", "temp-directory", "tempDirectory", "temp-folder", "tempFolder", "temp-fold", "tempFold", "tempFldr"},
		Short:   "Obtém o path absoluto do diretório temporário da aplicação",
		Long:    "Obtém o path absoluto do diretório temporário da aplicação utilizado para cache e arquivos temporários",
		RunE: func(cmd *cobra.Command, args []string) error {
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")
			tempDir, err := utils.GetTempDir()
			if quietFlagValue {
				if err != nil {
					return nil
				} else {
					fmt.Printf("%s\n", tempDir)
				}
			} else {
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("TempDir path:", tempDir)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}

// pathsCmdEnsureTempDir cria um comando Cobra para garantir a existência do diretório temporário da aplicação.
// Retorna um ponteiro para o comando Cobra configurado.
func pathsCmdEnsureTempDir() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure-temp",
		Short: "Garante a existência do diretório temporário configurado para a aplicação",
		Long:  "Garante a existência do diretório temporário configurado para a aplicação, criando-o se necessário",
		RunE: func(cmd *cobra.Command, args []string) error {
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			err := utils.EnsureTempDir()

			if quietFlagValue {
				if err != nil {
					return nil
				}
			} else {
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("TempDir ensured")
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}

// watchKubexFiles cria um comando Cobra para assistir os arquivos Kubex.
// Retorna um ponteiro para o comando Cobra configurado.
func watchKubexFiles() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubex-files",
		Aliases: []string{"wkubex", "watch-kubex", "watchKubex", "kubex-tree"},
		Short:   "Watch for changes in the Kubex files",
		Long:    "Watch for changes in the Kubex files and reload the service",
		RunE: func(cmd *cobra.Command, args []string) error {
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			err := utils.WatchKubexFiles()

			if quietFlagValue {
				if err != nil {
					return nil
				}
			} else {
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Kubex files watched")
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Execução silenciosa")

	return cmd
}
