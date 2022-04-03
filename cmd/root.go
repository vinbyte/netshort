/*
Copyright © 2020 Gavinda Kinandana <hai@gavinda.dev>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "netshort",
	Short: "netshort is Netlify link shortener",
	Long: `A simple link shortener specifically for Netlify.
You can use this app to build your own link shortener.
Built with love by Gavinda Kinandana using Go.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/netshort.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name "netshort" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("netshort")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func checkAppDir() bool {
	dir := viper.GetString("app.path")
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir + "/_redirects")
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func readFile(checkIsDuplicate bool, shortlink string, isAutoRegenerateLink bool) bool {
	file, err := os.Open(viper.GetString("app.path") + "/_redirects")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	isDuplicate := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if checkIsDuplicate {
			split := strings.Split(scanner.Text(), " ")
			if split[0] == "/"+ShortLink {
				isDuplicate = true
				file.Close()
				if isAutoRegenerateLink {
					ShortLink = randomizeShortLink(linkLength)
					readFile(checkIsDuplicate, ShortLink, isAutoRegenerateLink)
				}
				break
			}
		} else {
			fmt.Println(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return isDuplicate
}
