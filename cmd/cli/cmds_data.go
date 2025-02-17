package cli

import (
	"crypto/rsa"
	"fmt"
	"github.com/faelmori/gokubexfs/internal/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

// DataCmdsList retorna uma lista de comandos Cobra relacionados a dados.
// Retorna um slice de ponteiros para comandos Cobra.
func DataCmdsList() []*cobra.Command {
	cmdContains := dataCmdContains()
	cmdClearScreen := dataCmdClearScreen()
	cmdEncryptData := dataCmdEncryptData()
	cmdDecryptData := dataCmdDecryptData()
	cmdHashData := dataCmdHashData()
	cmdValidateHash := dataCmdValidateHash()
	cmdCompressData := dataCmdCompressData()
	cmdDecompressData := dataCmdDecompressData()
	cmdEncodeData := dataCmdEncodeData()
	cmdDecodeData := dataCmdDecodeData()
	cmdSignData := dataCmdSignData()
	cmdVerifyData := dataCmdVerifyData()

	return []*cobra.Command{
		cmdContains,
		cmdClearScreen,
		cmdEncryptData,
		cmdDecryptData,
		cmdHashData,
		cmdValidateHash,
		cmdCompressData,
		cmdDecompressData,
		cmdEncodeData,
		cmdDecodeData,
		cmdSignData,
		cmdVerifyData,
	}
}

// dataCmdContains cria um comando Cobra para verificar se um elemento está presente em uma coleção.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdContains() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contains",
		Short: "Busca algum elemento em uma coleção ou objeto",
		Long:  "Busca algum elemento em uma coleção ou objeto (slice, array, map, string)",
		RunE: func(cmd *cobra.Command, args []string) error {
			element, _ := cmd.Flags().GetString("element")
			collection, _ := cmd.Flags().GetStringArray("collection")
			quiet, _ := cmd.Flags().GetBool("quiet")
			if len(collection) == 0 {
				if len(args) >= 1 {
					if element != "" {
						collection = args[0:]
					} else {
						return fmt.Errorf("element is required in flag if collection is not provided as flag")
					}
				} else {
					return fmt.Errorf("collection is required")
				}
			}
			if element == "" {
				if len(args) >= 1 {
					if len(collection) >= 1 {
						element = args[0]
					} else {
						return fmt.Errorf("element is required")
					}
				} else {
					return fmt.Errorf("element is required")
				}
			}

			contains := utils.Contains(collection, element)
			if quiet {
				fmt.Printf("%t\n", contains)
				return nil
			}
			if contains {
				fmt.Println("Elemento encontrado na coleção")
			} else {
				fmt.Println("Elemento não encontrado na coleção")
			}
			return nil
		},
	}

	cmd.Flags().StringArrayP("collection", "c", []string{}, "Collection to be verified")
	cmd.Flags().StringP("element", "e", "", "Element to look for in the collection")
	cmd.Flags().BoolP("ignore-case", "i", false, "Ignore case when comparing strings")
	cmd.Flags().BoolP("quiet", "q", false, "Quiet mode (no output)")

	return cmd
}

// dataCmdClearScreen cria um comando Cobra para limpar a tela do terminal.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdClearScreen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Limpa a tela",
		Long:  "Limpa a tela do terminal",
		RunE: func(cmd *cobra.Command, args []string) error {
			xFlagValue, _ := cmd.Flags().GetBool("x")
			if !xFlagValue {
				utils.ClearScreen()
			} else {
				_ = exec.Command("clear").Run()
			}
			return nil
		},
	}

	cmd.Flags().BoolP("x", "x", false, "Clear screen buffer too")

	return cmd
}

// dataCmdEncryptData cria um comando Cobra para encriptar dados usando uma chave fornecida.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdEncryptData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encripta dados",
		Long:  "Encripta dados usando uma chave fornecida",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			keyFlagValue, _ := cmd.Flags().GetString("key")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, key, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if keyFlagValue == "" {
				return fmt.Errorf("é necessário fornecer a chave para encriptação (flag --key)")
			} else {
				key = keyFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			encryptedData, err := utils.EncryptData(data, key)
			if err != nil {
				return fmt.Errorf("erro ao encriptar dados: %v", err)
			}

			if output != "" {
				err = os.WriteFile(output, []byte(encryptedData), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever dados encriptados no arquivo: %v", err)
				}
			} else {
				if !quietFlagValue {
					fmt.Printf("%s\n", encryptedData)
				} else {
					return fmt.Errorf("quiet mode requires output argument")
				}
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("key", "k", "", "Chave para encriptação")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para os dados encriptados")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdDecryptData cria um comando Cobra para desencriptar dados usando uma chave fornecida.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdDecryptData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Descriptografia de dados",
		Long:  "Descriptografa dados encriptados usando uma chave fornecida",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			keyFlagValue, _ := cmd.Flags().GetString("key")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, key, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para desencriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if keyFlagValue == "" {
				return fmt.Errorf("é necessário fornecer a chave para desencriptografia (flag --key)")
			} else {
				key = keyFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			decryptedData, err := utils.DecryptData(data, key)
			if err != nil {
				return err
			}
			if output != "" {
				err = os.WriteFile(output, []byte(decryptedData), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever dados desencriptados no arquivo: %v", err)
				}
			} else {
				if !quietFlagValue {
					fmt.Printf("%s\n", decryptedData)
				} else {
					return fmt.Errorf("quiet mode requires output argument")
				}
			}
			return nil
		},
	}

	return cmd
}

// dataCmdHashData cria um comando Cobra para gerar um hash dos dados fornecidos.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdHashData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash",
		Short: "Gera um hash dos dados",
		Long:  "Gera um hash SHA-256 dos dados fornecidos",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			hash := utils.HashData(data)

			if output != "" {
				err := os.WriteFile(output, []byte(hash), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever hash dos dados no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Hash dos dados escrito em %s\n", output)
				}
			} else {
				fmt.Printf("%s\n", hash)
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para o hash dos dados")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdValidateHash cria um comando Cobra para validar se os dados correspondem ao hash fornecido.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdValidateHash() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Valida um hash",
		Long:  "Valida se os dados correspondem ao hash fornecido",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			hashFlagValue, _ := cmd.Flags().GetString("hash")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, hash, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if hashFlagValue == "" {
				if len(args) < 2 {
					return fmt.Errorf("é necessário fornecer o hash para validação")
				}
				hash = args[1]
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			valid := utils.ValidateHash(data, hash)

			if output != "" {
				err := os.WriteFile(output, []byte(fmt.Sprintf("%t", valid)), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever validação do hash no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Validação do hash escrita em %s\n", output)
				}
			} else {
				fmt.Printf("%t\n", valid)
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("hash", "h", "", "Hash a ser validado")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para a validação do hash")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdCompressData cria um comando Cobra para comprimir dados usando gzip.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdCompressData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compress",
		Short: "Comprime dados",
		Long:  "Comprime dados usando gzip",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			compressedData, err := utils.CompressData(data)
			if err != nil {
				return err
			}

			if output != "" {
				err = os.WriteFile(output, []byte(compressedData), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever dados comprimidos no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Dados comprimidos escritos em %s\n", output)
				}
			} else {
				fmt.Printf("%s\n", compressedData)
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para os dados comprimidos")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdDecompressData cria um comando Cobra para descomprimir dados previamente comprimidos.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdDecompressData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decompress",
		Short: "Descomprime dados",
		Long:  "Descomprime dados previamente comprimidos",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			decompressedData, err := utils.DecompressData(data)
			if err != nil {
				return err
			}
			if output != "" {
				err = os.WriteFile(output, []byte(decompressedData), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever dados comprimidos no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Dados comprimidos escritos em %s\n", output)
				}
			} else {
				fmt.Printf("%s\n", decompressedData)
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para os dados descomprimidos")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdEncodeData cria um comando Cobra para codificar dados em Base64.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdEncodeData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode",
		Short: "Codifica dados",
		Long:  "Codifica dados em Base64",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			encodedData := utils.EncodeData(data)

			if output != "" {
				err := os.WriteFile(output, []byte(encodedData), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever dados codificados no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Dados codificados escritos em %s\n", output)
				}
			} else {
				fmt.Printf("%s\n", encodedData)
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para os dados codificados")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdDecodeData cria um comando Cobra para decodificar dados de Base64.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdDecodeData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode",
		Short: "Decodifica dados",
		Long:  "Decodifica dados de Base64",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			decodedData, err := utils.DecodeData(data)
			if err != nil {
				return err
			}

			if output != "" {
				err = os.WriteFile(output, []byte(decodedData), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever dados decodificados no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Dados decodificados escritos em %s\n", output)
				}
			} else {
				fmt.Printf("%s\n", decodedData)
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para os dados decodificados")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdSignData cria um comando Cobra para assinar dados digitalmente usando uma chave privada.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdSignData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Assina dados",
		Long:  "Assina dados digitalmente usando uma chave privada",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			keyFlagValue, _ := cmd.Flags().GetString("key")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, key, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if keyFlagValue == "" {
				return fmt.Errorf("é necessário fornecer a chave para encriptação (flag --key)")
			} else {
				key = keyFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			privateKey, err := utils.LoadPrivateKey(key)
			if err != nil {
				return err
			}

			signedData, err := utils.SignData(data, privateKey)
			if err != nil {
				return err
			}

			if output != "" {
				err = os.WriteFile(output, []byte(signedData), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever dados assinados no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Dados assinados escritos em %s\n", output)
				}
			} else {
				fmt.Printf("%s\n", signedData)
			}

			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("key", "k", "", "Chave privada para assinatura")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para os dados assinados")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}

// dataCmdVerifyData cria um comando Cobra para verificar a assinatura digital dos dados usando uma chave pública.
// Retorna um ponteiro para o comando Cobra configurado.
func dataCmdVerifyData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verifica assinatura",
		Long:  "Verifica a assinatura digital dos dados usando uma chave pública",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFlagValue, _ := cmd.Flags().GetString("input")
			pubKeyFlagValue, _ := cmd.Flags().GetString("key")
			signatureFlagValue, _ := cmd.Flags().GetString("signature")
			outputFlagValue, _ := cmd.Flags().GetString("output")
			quietFlagValue, _ := cmd.Flags().GetBool("quiet")

			var data, pubKey, signature, output string
			if inputFlagValue == "" {
				if len(args) < 1 {
					return fmt.Errorf("é necessário fornecer os dados para encriptação")
				}
				data = args[0]
			} else {
				data = inputFlagValue
			}
			if pubKeyFlagValue == "" {
				return fmt.Errorf("é necessário fornecer a chave para encriptação (flag --key)")
			} else {
				pubKey = pubKeyFlagValue
			}
			if signatureFlagValue == "" {
				return fmt.Errorf("é necessário fornecer a chave para encriptação (flag --key)")
			} else {
				signature = signatureFlagValue
			}
			if outputFlagValue != "" {
				output = outputFlagValue
			}

			publicKey, err := utils.LoadPublicKey(pubKey)
			if err != nil {
				return err
			}

			err = utils.VerifyData(data, signature, publicKey.(*rsa.PublicKey))
			if err != nil {
				return err
			}

			if output != "" {
				err = os.WriteFile(output, []byte("true"), 0644)
				if err != nil {
					return fmt.Errorf("erro ao escrever validação da assinatura no arquivo: %v", err)
				}
				if !quietFlagValue {
					fmt.Printf("Validação da assinatura escrita em %s\n", output)
				}
			} else {
				fmt.Printf("%t\n", true)
			}
			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "", "Arquivo de entrada para os dados a serem encriptados")
	cmd.Flags().StringP("key", "k", "", "Chave pública para verificação")
	cmd.Flags().StringP("signature", "s", "", "Assinatura digital dos dados")
	cmd.Flags().StringP("output", "o", "", "Arquivo de saída para a validação da assinatura")
	cmd.Flags().BoolP("quiet", "q", false, "Modo silencioso (sem saída)")

	return cmd
}
