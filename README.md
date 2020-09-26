# netshort

## Overview

**netshort** is a Netlify shortener build using Go. Inspired by [netlify-shortener](https://github.com/kentcdodds/netlify-shortener).

## Why

Simple reason, I love Go. I prefer using Go for CLI rather than NodeJS. Beside that, when I want deploy [this](https://github.com/kentcdodds/netlify-shortener) in my DO server, I got this error :

```
Error: Both xsel and fallback failed
    at makeError (/home/gavinda/tes-shortener/node_modules/clipboardy/lib/linux.js:16:11)
    at xselWithFallbackSync (/home/gavinda/tes-shortener/node_modules/clipboardy/lib/linux.js:43:10)
    at Object.copySync (/home/gavinda/tes-shortener/node_modules/clipboardy/lib/linux.js:53:3)
    at Object.exports.writeSync (/home/gavinda/tes-shortener/node_modules/clipboardy/index.js:45:14)
    at Object.<anonymous> (/home/gavinda/tes-shortener/node_modules/netlify-shortener/dist/index.js:56:12)
    at Module._compile (internal/modules/cjs/loader.js:1137:30)
    at Object.Module._extensions..js (internal/modules/cjs/loader.js:1157:10)
    at Module.load (internal/modules/cjs/loader.js:985:32)
    at Function.Module._load (internal/modules/cjs/loader.js:878:14)
    at Function.executeUserEntryPoint [as runMain] (internal/modules/run_main.js:71:12) {
  xselError: Error: Command failed: xsel --clipboard --input
  xsel: Can't open display: (null)
  : Inappropriate ioctl for device
```
So, I make the decision to build my own shortener CLI. Then, the **netshort** was born.

## Prerequisite

1. Make sure you have your app in different directory included **_redirects** file inside. If you don't have, please create it first. For example :
`mkdir shortener && cd shortener && touch _redirects`
1. Add the basic redirection. Add this line to your **_redirects** file
`/*             [your_website_or_social_media]`
1. Make your repository in github (public or private up to you) and add it to your Netlify account. You read [this blog](https://www.netlify.com/blog/2016/10/27/a-step-by-step-guide-deploying-a-static-site-or-single-page-app/) for step by step setup your Netlify.
1. Commit and push your **_redirects** file to your github repo.
1. [optional] Setup the local git config. **netshort** will use this to auto push to your repository. For example:
`git config user.name "YOUR_NAME"`
`git config user.email "YOUR_EMAIL"`
`git config credential.helper store`
`git pull`
But if you don't mind to use the current global git config, you can skip this step.
1. Copy `netshort.yaml` config to your $HOME/netshort.yaml. You can also put it different location, but you need to add flags `--config [YOUR_CONFIG_PATH]` whenever you run **netshort**.

## Config

- `app.path` is used to defined your **_redirects** file location (mandatory)
- `shortlink.length` is used to generate the length of short link (optional, default : 5)

## Install

1. Download the binary [here](https://github.com/rabbitmeow/netshort/releases)
2. For macOS/Linux change permission of binary. Example :
`chmod +x netshort_darwin_amd64`
3. Additional step for macOS Catalina :

- Go to your download folder, right click the file. Then choose "Open"
- A popup will appear, click the "Open" button
- Close the terminal

4. If you are using macOS/Linux, you can move the binary to `/usr/local/bin` to make it global to access. Example :
`mv ./netshort_darwin_amd64 /usr/local/bin/netshort`
5. **netshort** ready to use

## Command

- `netshort shorten [long_url] [custom_short_link]`
This used for add the url to your **_redirect** file in your app directory (defined in config file). The second args is optional. **netshort** will generate for you if you don't have the second args
- `netshort list`
This used for listing your link. It will read all your link in **_redirects** file

## Tips

You can combine with [telegram-bot](https://github.com/botgram/shell-bot) to execute via telegram chat. So, you can shorten link anywhere, anytime only using a telegram app.
Demo :
![netshort-telegram-demo](https://api.monosnap.com/file/download?id=7M0NULpN5OfJebzx6TMB4BrVZTek7Q)