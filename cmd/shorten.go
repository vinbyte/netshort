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
		LongURL = args[0]
		if !isValidURL(LongURL) {
			fmt.Println("URL not valid")
			os.Exit(1)
		}
		if !checkAppDir() {
			fmt.Println("_redirect file does not exist (or is a directory)")
			os.Exit(1)
		}
		isAutoRegenerateLink := false
		if len(args) > 1 {
			if !isAlphaNum(args[1]) {
				fmt.Println("Short link is alphanumeric only")
				os.Exit(1)
			}
			ShortLink = args[1]
		} else {
			linkLength = 5
			if viper.GetInt("shortlink.length") != 0 {
				linkLength = viper.GetInt("shortlink.length")
			}
			ShortLink = generateShortLink(linkLength)
			isAutoRegenerateLink = true
		}
		// start prepend to _redirects file
		isDuplicate := readFile(true, ShortLink, isAutoRegenerateLink)
		if isDuplicate {
			fmt.Println("Your short link already exist")
			os.Exit(1)
		}
		//give whitespace
		result := fmt.Sprintf("%-"+strconv.Itoa(len(ShortLink)+10)+"v%s\n", "/"+ShortLink, LongURL)
		readF, err := os.Open(viper.GetString("app.path") + "/_redirects")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		scanner := bufio.NewScanner(readF)
		currentContents := ""
		for scanner.Scan() {
			// adjust the whitespace
			tmp := scanner.Text()
			split := strings.Fields(tmp)
			addLength := (len(ShortLink) + 10) - len(split[0])
			new := fmt.Sprintf("%-"+strconv.Itoa(len(split[0])+addLength)+"v%s\n", split[0], split[1])
			currentContents += new
		}
		readF.Close()
		writeF, err := os.OpenFile(viper.GetString("app.path")+"/_redirects", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if _, err = writeF.Write([]byte(result + currentContents)); err != nil {
			panic(err)
		}
		writeF.Close()
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

func generateShortLink(n int) string {
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
