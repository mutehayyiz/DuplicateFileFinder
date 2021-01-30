package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func ReadFilePathes(paths []string) map[string][]string {
	var file = make(map[string][]string)
	for _, path := range paths {
		err := filepath.Walk(path,
			func(path string, info os.FileInfo, err error) error {

				// don't care hidden directories
				if info.IsDir() && info.Name() == "." {
					return filepath.SkipDir
				}

				// don't care directory names
				if !info.IsDir() {
					size := strconv.FormatInt(info.Size(), 10)
					file[size] = append(file[size], path)
				}
				return nil
			})
		if err != nil {
			fmt.Println(err)
		}
	}
	return file
}

func IgnoreUniques(files map[string][]string) {
	for k, v := range files {
		if len(v) <= 1 {
			delete(files, k)
		}
	}
}

func CalculateHash(bytes []byte) string {
	h := sha256.New()
	h.Write(bytes)
	return hex.EncodeToString(h.Sum(nil))
}

func HashFile(files map[string][]string) map[string][]string {
	var tmp = make(map[string][]string)
	for _, element := range files {
		for _, i := range element {
			bytes, _ := ioutil.ReadFile(i)
			c := CalculateHash(bytes)
			tmp[c] = append(tmp[c], i)
		}
	}
	return tmp
}

func findDuplicates(pathList []string) map[string][]string {
	paths := ReadFilePathes(pathList)
	IgnoreUniques(paths)
	paths = HashFile(paths)
	IgnoreUniques(paths)
	return paths
}

func Write(results map[string][]string) error {
	bytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return errors.New("couldn't marshal configuration")
	}

	err = ioutil.WriteFile("output.json", bytes, 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("couldn't write license to file: %s", err))
	}
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "main",
	Short: "The finder for duplicate files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		results := findDuplicates(args)

		save, _ := cmd.Flags().GetBool("save")
		if save {
			err:=Write(results)
			if err != nil {
				fmt.Println("couldn't save to output.json ", save)

			}
			fmt.Println("saved to output.json ", save)
		} else {
			for index, element := range results {
				fmt.Println(index, "====>")
				for _, e := range element {
					fmt.Println("\t", e)
				}
			}
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("save", "s", false, "save results as json")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
