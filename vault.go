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
	app.UsageTemplate(kingpin.SeparateOptionalFlagsUsageTemplate)

	appServer := app.Command("server", "run the HTTP interface")
	appServerListen := appServer.Flag("listen", "address on which to listen on").Short('l').Default("127.0.0.1:8080").TCP()
	appServerAPIKey := appServer.Flag("apikey", "API key to use for all requests").Short('k').Required().String()

	appInit := app.Command("init", "initiate the vault")

	appKey := app.Command("key", "vault key management")
	appKeyList := appKey.Command("list", "list all keys available in the vault")
	appKeyAdd := appKey.Command("add", "add a key that unlocks the vault")
	appKeyAddComment := appKeyAdd.Flag("comment", "description of this key").Short('c').Required().String()
	appKeyDelete := appKey.Command("delete", "delete a key from the vault")
	appKeyDeleteID := appKeyDelete.Arg("id", "ID of the key to delete").Required().Int()
	appKeyRotate := appKey.Command("rotate", "[EXPERIMENTAL] rotate the vault master key")

	appList := app.Command("list", "list all secrets")
	appListPath := appList.Arg("path", "secret path").Default("/").String()

	appShow := app.Command("show", "show all secrets")
	appShowPath := appShow.Arg("path", "secret path").Required().String()
	appShowPrint := appShow.Flag("print", "print 'password' attribute on console").Short('p').Bool()
	appShowClipboard := appShow.Flag("clip", "copy 'password' attribute into clipboard").Short('c').Bool()
	appShowClipAttr := appShow.Flag("clip-attributes", "attribute to copy to clipboard").Short('a').Default("").String()
	appShowWrite := appShow.Flag("write", "write file attributes on the filesystem").Short('w').Bool()
	appShowWriteFiles := appShow.Flag("file", "which file attributes to write").Short('f').Strings()
	appShowWriteStdout := appShow.Flag("stdout", "print file attribute to STDOUT").Short('s').Bool()

	appAdd := app.Command("add", "add a secret")
	appAddPath := appAdd.Arg("path", "secret path").Required().String()
	appAddAttrs := appAdd.Arg("attributes", "secret attributes").Required().StringMap()
	appAddGeneratorLength := appAdd.Flag("length", "length of generated passwords").Short('l').Default("16").Int()
	appAddGeneratorSymbols := appAdd.Flag("symbols", "include special characters in the generated password").Default("false").Bool()

	appEdit := app.Command("edit", "edit an existing secret")
	appEditPath := appEdit.Arg("path", "path to the secret to edit").Required().String()
	appEditDeletedAttrs := appEdit.Flag("delete", "attributes to delete from the secret").Short('d').Strings()
	appEditAttrs := appEdit.Arg("attributes", "secret attributes").StringMap()
	appEditGeneratorLength := appEdit.Flag("length", "length of generated passwords").Short('l').Default("16").Int()
	appEditGeneratorSymbols := appEdit.Flag("symbols", "include special characters in the generated password").Default("false").Bool()

	appRename := app.Command("rename", "rename a secret")
	appRenamePath := appRename.Arg("path", "path to the secret to rename").Required().String()
	appRenameNewPath := appRename.Arg("newpath", "new path to secret").Required().String()

	appDelete := app.Command("delete", "delete a secret")
	appDeletePath := appDelete.Arg("path", "secret path").Required().String()

	appGit := app.Command("git", "archive the store in a git repository")
	appGitClone := appGit.Command("clone", "clone an existing store repository")
	appGitCloneURL := appGitClone.Arg("url", "remote store repository URL").Required().String()
	appGitRemote := appGit.Command("remote", "set the remote git repository to push to")
	appGitRemoteURL := appGitRemote.Arg("url", "git repository URL").Required().String()
	appGitPush := appGit.Command("push", "push the state of the store")
	appGitPull := appGit.Command("pull", "pull the state of the store")

	appUnseal := app.Command("unseal", "unseal store until next reboot")
	appSeal := app.Command("seal", "seal store")

	args := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch args {
	case appServer.FullCommand():
		StartServer(*appServerListen, *appServerAPIKey)

	case appInit.FullCommand():
		crypt.InitVault()

	case appGitClone.FullCommand():
		util.GitClone(*appGitCloneURL)
	}

	util.AssertVaultExists()

	switch args {
	case appKeyList.FullCommand():
		crypt.ListKeys()
	case appKeyAdd.FullCommand():
		crypt.AddKey(*appKeyAddComment)
	case appKeyDelete.FullCommand():
		crypt.DeleteKey(*appKeyDeleteID)
	case appKeyRotate.FullCommand():
		crypt.RotateKey()

	case appList.FullCommand():
		listSecrets(*appListPath)
	case appShow.FullCommand():
		showSecret(*appShowPath, *appShowPrint, *appShowClipboard, *appShowClipAttr, *appShowWrite, *appShowWriteFiles, *appShowWriteStdout)
	case appAdd.FullCommand():
		addSecret(*appAddPath, *appAddAttrs, *appAddGeneratorLength, *appAddGeneratorSymbols, false, []string{})
	case appEdit.FullCommand():
		editSecret(*appEditPath, *appEditAttrs, *appEditDeletedAttrs, *appEditGeneratorLength, *appEditGeneratorSymbols)
	case appRename.FullCommand():
		renameSecret(*appRenamePath, *appRenameNewPath)
	case appDelete.FullCommand():
		deleteSecret(*appDeletePath)

	case appGitRemote.FullCommand():
		util.GitRemote(*appGitRemoteURL)
	case appGitPush.FullCommand():
		util.GitPush()
	case appGitPull.FullCommand():
		util.GitPull()

	case appUnseal.FullCommand():
		crypt.Unseal()
	case appSeal.FullCommand():
		crypt.Seal(false)
	}
}
