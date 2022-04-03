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
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LongURL is the link that you want to short
var LongURL string

// ShortLink is the short link
var ShortLink string
var linkLength int
var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var isAlphaNum = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

const (
	ErrInvalidURL           = "URL not valid"
	ErrRedirectFileNotFound = "_redirect file does not exist (or is a directory)"
	ErrAlphanumericOnly     = "short link is alphanumeric only"
	ErrShortLinkExist       = "Your short link already exist"
	ErrNoWhitespaceAtLine   = "no whitespace detected at _redirects file line %d"

	AdditionalWhitespaceLength = 10
)

// shortenCmd represents the shorten command
var shortenCmd = &cobra.Command{
	Use:   "shorten <long_url> <custom_short_link>",
	Short: "Shorten the long url and provide a short link",
	Long: `netshort will process your <long_url> in first args and generate the short link. 
If the <custom_short_link> submitted, it will use it as short link. 
The result will put in the _redirects file in your app path specified at config file.
Then push it to your git repository.
For example:

netshort shorten https://google.com
This will generate a random short link with length specified at config file (default: 5).

netshort shorten https://google.com goo
This will use /goo as a short link.`,
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		// validating args
		err := validateParam(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		isAutoGenerateShortLink, err := generateShortLink(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// start prepend to _redirects file
		isDuplicate := readFile(true, ShortLink, isAutoGenerateShortLink)
		if isDuplicate {
			fmt.Println(ErrShortLinkExist)
			os.Exit(1)
		}

		err = updateRedirectFile()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		pushToGitRepo()

		fmt.Println("/" + ShortLink + " -> " + LongURL)
	},
}

func init() {
	rootCmd.AddCommand(shortenCmd)
	rand.Seed(time.Now().UnixNano())
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func validateParam(args []string) (err error) {
	LongURL = args[0]
	if !isValidURL(LongURL) {
		err = errors.New(ErrInvalidURL)
		return
	}
	if !checkAppDir() {
		err = errors.New(ErrRedirectFileNotFound)
	}

	return
}

func generateShortLink(args []string) (isAutoRegenerateLink bool, err error) {
	isAutoRegenerateLink = false
	if len(args) > 1 {
		if !isAlphaNum(args[1]) {
			err = errors.New(ErrAlphanumericOnly)
			return
		}
		ShortLink = args[1]
	} else {
		linkLength = 5
		if viper.GetInt("shortlink.length") != 0 {
			linkLength = viper.GetInt("shortlink.length")
		}
		ShortLink = randomizeShortLink(linkLength)
		isAutoRegenerateLink = true
	}

	return
}

func updateRedirectFile() (err error) {
	//give whitespace

	// read the _redirects file
	readF, err := os.Open(viper.GetString("app.path") + "/_redirects")
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(readF)
	currentFileContents := []string{}
	newFileContents := ""
	// find the longest short link
	line := 0
	longestShortLink := 0
	for scanner.Scan() {
		line++
		// adjust the whitespace
		tmp := scanner.Text()
		split := strings.Fields(tmp)
		if len(split) < 2 {
			err = fmt.Errorf(ErrNoWhitespaceAtLine, line)
			return
		}
		shortLink := split[0]
		if len(shortLink) > longestShortLink {
			longestShortLink = len(shortLink)
		}
		currentFileContents = append(currentFileContents, tmp)
	}
	readF.Close()

	totalLength := longestShortLink + AdditionalWhitespaceLength
	newEntry := fmt.Sprintf("%-"+strconv.Itoa(totalLength)+"v%s\n", "/"+ShortLink, LongURL)
	newFileContents += newEntry

	for _, c := range currentFileContents {
		split := strings.Fields(c)
		shortLink := split[0]
		longLink := split[1]
		new := fmt.Sprintf("%-"+strconv.Itoa(totalLength)+"v%s\n", shortLink, longLink)
		newFileContents += new
	}

	writeF, err := os.OpenFile(viper.GetString("app.path")+"/_redirects", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	if _, err = writeF.Write([]byte(newFileContents)); err != nil {
		panic(err)
	}
	writeF.Close()

	return
}

func pushToGitRepo() (err error) {
	pushCommand := `cd ` + viper.GetString("app.path") + ` && git add _redirects && git commit -m "add /` + ShortLink + `" && git push origin master`
	var pushCmd *exec.Cmd
	var stdout, stderr bytes.Buffer
	if runtime.GOOS == "windows" {
		pushCmd = exec.Command("cmd", "/C", pushCommand)
	} else {
		pushCmd = exec.Command("bash", "-c", pushCommand)
	}
	pushCmd.Stdout = &stdout
	pushCmd.Stderr = &stderr
	err = pushCmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	out := stdout.String() + stderr.String()
	fmt.Println(out + "\n")

	return
}

func randomizeShortLink(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
