package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	// Define as flags que podem ser passadas por linha de comando
	pFolder := flag.String("folder", "", "folder of files to be concatenate")
	pFilter := flag.String("filter", "", "filter do select files inside the folder")
	pOutputFile := flag.String("out", "", "output file with concatenated data")
	pRemoveConcatenatedFiles := flag.Bool("remove", false, "remove concatenated files")
	pErrorNoFiles := flag.Bool("errornofiles", false, "generate error if no files was selected")
	pAppendOutputFile := flag.Bool("append", true, "append data to the end of output file")
	flag.Parse()
	// define o horario de inicio da execução
	started := time.Now()
	// ajusta o caminho da pasta, se não foi informado define como o caminho do
	// programa
	if *pFolder == "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		*pFolder = dir
	}
	// encerra com erro se não foi informado o filtro
	if *pFilter == "" {
		log.Fatal("filter not provided")
	}
	// encerra com erro se não foi informado o arquivo de saída
	if *pOutputFile == "" {
		log.Fatal("output file not provided")
	}
	// define a mascara para filtar os arquivos através de expressão regular
	*pFilter = regexp.QuoteMeta(*pFilter)
	*pFilter = strings.ReplaceAll(*pFilter, "\\*", ".+")
	pattern, err := regexp.Compile(*pFilter)
	if err != nil {
		log.Fatal("invalid filter provided")
	}
	// define a lista de arquivos que serão concatenados
	var files []os.FileInfo
	// lista os arquivos
	err = filepath.Walk(*pFolder, func(path string, info os.FileInfo, err error) error {
		// retorna se houve erro
		if err != nil {
			return err
		}
		// ignora pastas
		if info.IsDir() {
			return nil
		}
		// valida se deve selecionar o arquivo
		if pattern.Match([]byte(info.Name())) {
			files = append(files, info)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("unable to list files, %s", err)
	}
	// define os flags para criar o arquivo de saída
	// substitui o conteúdo do arquivo ou adiciona ao final
	flags := os.O_CREATE | os.O_WRONLY
	if *pAppendOutputFile {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	// criar o arquivo de saída que irá conter os dados concatenados
	out, err := os.OpenFile(*pOutputFile, flags, 0775)
	if err != nil {
		log.Fatalf("unable to create output file, %s", err)
	}
	defer out.Close()
	// define a lista de arquivos que foram concatenados com sucesso
	var filesWithSuccess []os.FileInfo
	// concatena os arquivos selecionados no arquivo de saída
	for k, v := range files {
		// define o caminho completo para o arquivo atual
		filePath := path.Join(*pFolder, v.Name())
		log.Printf("[%d] concatenating file {%s}\n", k, filePath)
		// abre o arquivo para leitura
		in, err := os.OpenFile(filePath, os.O_RDONLY, 0775)
		if err != nil {
			log.Printf("unable to open file, %s", err)
			continue
		}
		// copia o conteúdo do arquivo atual para o arquivo de saída
		n, err := io.Copy(out, in)
		if err != nil {
			in.Close()
			if n == 0 {
				log.Printf("unable to copy data to output file, %s", err)
				continue
			} else {
				log.Fatalf("unable to copy data to output file, %s", err)
			}
		}
		// adiciona o arquivo na lista de sucesso
		filesWithSuccess = append(filesWithSuccess, v)
		log.Printf("[%d] file {%s} concatenated successfully, bytes reads {%d}\n", k, filePath, n)
		in.Close()
	}
	// verifica se deve remover os arquivos concatenados
	if *pRemoveConcatenatedFiles {
		for k, v := range filesWithSuccess {
			// define o caminho completo para o arquivo atual
			filePath := path.Join(*pFolder, v.Name())
			err := os.Remove(filePath)
			if err != nil {
				log.Fatalf("unable to remove file, %s", err)
			}
			log.Printf("[%d] file {%s} removed successfully\n", k, filePath)
		}
	}
	// calcula o tempo em segundos da execução
	elapsed := time.Since(started).Seconds()
	// faz o tratamento caso não existam arquivos
	if len(files) == 0 {
		if *pErrorNoFiles {
			log.Fatalf("no files found to be concatenated")
		}
		log.Printf("no files found to be concatenated")
	} else {
		log.Printf("%d files of %d was concatenated in {%s} in %.2f seconds\n", len(filesWithSuccess), len(files), *pOutputFile, elapsed)
	}
}
