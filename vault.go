package main

import (
	"os"

	"github.com/apognu/vault/crypt"
	"github.com/apognu/vault/util"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := kingpin.New("vault", "Simple encrypted data store")
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.CompactUsageTemplate)

	appInit := app.Command("init", "initiate the vault")

	appList := app.Command("list", "list all secrets")
	appListPath := appList.Arg("path", "secret path").Default("/").String()

	appShow := app.Command("show", "show all secrets")
	appShowPath := appShow.Arg("path", "secret path").Required().String()
	appShowPrint := appShow.Flag("print", "print 'password' attribute on console").Short('p').Bool()
	appShowClipboard := appShow.Flag("clip", "copy 'password' attribute into clipboard").Short('c').Bool()
	appShowClipAttr := appShow.Flag("clip-attributes", "attribute to copy to clipboard").Short('a').Default("").String()

	appAdd := app.Command("add", "add a secret")
	appAddPath := appAdd.Arg("path", "secret path").Required().String()
	appAddAttrs := appAdd.Arg("attributes", "secret attributes").Required().StringMap()
	appAddGeneratorLength := appAdd.Flag("length", "length of generated passwords").Short('l').Default("16").Int()

	appEdit := app.Command("edit", "edit an existing secret")
	appEditPath := appEdit.Arg("path", "path to the secret to edit").Required().String()
	appEditDeletedAttrs := appEdit.Flag("delete", "attributes to delete from the secret").Short('d').Strings()
	appEditAttrs := appEdit.Arg("attributes", "secret attributes").StringMap()
	appEditGeneratorLength := appEdit.Flag("length", "length of generated passwords").Short('l').Default("16").Int()

	appDelete := app.Command("delete", "delete a secret")
	appDeletePath := appDelete.Arg("path", "secret path").Required().String()

	appGit := app.Command("git", "archive the store in a git repository")
	appGitClone := appGit.Command("clone", "clone an existing store repository")
	appGitCloneURL := appGitClone.Arg("url", "remote store repository URL").Required().String()
	appGitInit := appGit.Command("init", "initialize git local repository")
	appGitRemote := appGit.Command("remote", "set the remote git repository to push to")
	appGitRemoteURL := appGitRemote.Arg("url", "git repository URL").Required().String()
	appGitPush := appGit.Command("push", "push the state of the store")
	appGitPull := appGit.Command("pull", "pull the state of the store")

	appUnseal := app.Command("unseal", "unseal store until next reboot")
	appSeal := app.Command("seal", "seal store")

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case appInit.FullCommand():
		crypt.InitVault()
	}

	util.AssertVaultExists()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case appList.FullCommand():
		listSecrets(*appListPath)
	case appShow.FullCommand():
		showSecret(*appShowPath, *appShowPrint, *appShowClipboard, *appShowClipAttr)
	case appAdd.FullCommand():
		addSecret(*appAddPath, *appAddAttrs, []string{}, *appAddGeneratorLength, false, []string{})
	case appEdit.FullCommand():
		editSecret(*appEditPath, *appEditAttrs, *appEditDeletedAttrs, *appEditGeneratorLength)
	case appDelete.FullCommand():
		deleteSecret(*appDeletePath)
	case appGitClone.FullCommand():
		gitClone(*appGitCloneURL)
	case appGitInit.FullCommand():
		gitInit()
	case appGitRemote.FullCommand():
		gitRemote(*appGitRemoteURL)
	case appGitPush.FullCommand():
		gitPush()
	case appGitPull.FullCommand():
		gitPull()
	case appUnseal.FullCommand():
		crypt.Unseal()
	case appSeal.FullCommand():
		crypt.Seal()
	}
}
