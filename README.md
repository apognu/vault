# Vault

[![Build Status](https://travis-ci.org/apognu/vault.svg?branch=master)](https://travis-ci.org/apognu/vault)

Personal project of a git-based encrypted password store. Each "password" is actually a JSON-encoded map of strings, encrypted through AES-GCM.

The whole store directory is located under $HOME/.vault (overridable through the VAULT_PATH environment variable) and can be pushed to a remote git repository though the ```vault``` command.

**This is a draft in progress, I would be very cautious with using it to store your most precious passwords. Plus, the storage format is bound to change until 1.0, data will be unusable when it does.**

## Create the vault

The following command creates an empty vault and adds one passphrase to be used in it. Later on, you will be able to add more passphrase that can unlock the vault.

```
$ vault init
Enter passphrase:
Confirm:
INFO[0000] vault created successfully
```

## Key management

The user passphrase does not directly encrypt the store's secrets. Instead, on vault creation, a master key is randomly generated and encrypted with a key derived from the user password (through PBKDF2). This encrypted master key is stored in a file containing metadata about the store, directly alongside the secrets.

The store's master key can be encrypted with any number of passphrases, so several passphrases (or people) could be used to unlock the store. This also allows for seamlessly changing the passphrase used to unlock the store.

Also, a solution for rotating the store's master key is in the pipes.

The available keys in a store can be listed with the following command:

```
$ vault key list
vault key list
 - #0 (Tue, 09 Aug 2017, 16:25) Initial key created on vault creation
       721a9b52bfceacc503c056e3b9b93cfa
 - #1 (Tue, 09 Aug 2017, 21:34) Added key for whatever reason
       5d41402abc4b2a76b9719d911017c592
```

And deleted through this:

```
$ vault key delete 1
INFO[0000] key was successfully deleted
```

For obvious reasons, the last key stored in the store's metadata cannot be deleted. You'll have to create another one beforehand.

A new key can be added to the vault, with a custom comment, with this command:

```
$ vault key add -c 'New key'
Enter passphrase: 
New passphrase: 
Confirm: 
INFO[0007] key was successfully added
```

The command will prompt you for one of the existing passphrases, and then to enter and confirm the one you want to add.

## Add a secret

```
$ vault add dir/subdir/website.com username=apognu password=Str0ngP@ss
Enter passphrase:
INFO[0010] secret 'dir/subdir/website.com' created successfully
```

```vault``` is attribute-agnostic, there is no special handling of, for instance, the ```password``` attribute. You can add any number of attributes to an entry.

### Eyes-only attributes

One special kind of attribute is for _eyes-only_. They only differ in that they are not printed on the console by default, and they are input interactively. Any attribute set without a value will trigger the prompt and will never be printed without the ```-p``` option.

```
$ vault add website.com username=apognu password=
Value for 'password':
Enter passphrase:
INFO[0010] secret 'website.com' created successfully
```

### Generated passwords

One can generate random passwords (now with [A-Za-z0-9]) with the syntax ```attr=-```. By default, a random 16-character password will be generated for that attribute. Generated attributes will automatically be set as eyes-only.

```
$ vault add websites.com username=apognu password=-
```

One can generate passwords with a different size with the ```-l``` option.

### File attributes

An entire file can be embedded into an attribute with the syntax ```attr=@/path/to/file```. By default, any file attribute will not be printed on the console, and will require the use of ```-c``` or ```-p``` to be used.

```
$ vault add ssh/keys pubkey=@/home/apognu/.ssh/id_rsa.pub privkey=@/home/apognu/.ssh/id_rsa
INFO[0010] secret 'ssh/keys' created successfully
$ vault show ssh/keys
Store » ssh » keys
  privkey = <file content>
   pubkey = <file content>
```

When you use the ```-w``` option in combination with showing a secret containing file attributes, all the file attributes of that secret will be written to files in a directory named after the secret path.

```
$ vault show my/secret/file
Store » my » secret » file
  file = <file content>
$ vault show my/secret/file -w
INFO[0000] attribute written to 'vault-my-secret-file/file'
```

For now, all file attributes of the secrets are written to the output directory. A future version of vault may allow for selecting which attributes to consider for writing.

## Print a secret

```
$ vault show dir/subdir/website.com
Store » dir » subdir » website.com
       url = http://example.com/login
  username = apognu
  password = <redacted>
```

The ```-p``` option can be used to display the redacted attributes.

The ```-c``` option can be used to copy one attribute to the clipboard. By default, if the entry contains ony one eyes-only attribute, it will be used. If there are more than one eyes-only attribute, the attribute named ```password``` will be copied. If you would like to copy another attribute to your clipboard, use the ```-a``` option.

## Edit a secret

The syntax for modifying an existing secret is exactly the same as the one used to create one, with one addition: an optional list of attributes to delete.

```
$ vault edit website.com -d url username=newlogin password=
```

This command will delete thre ```url``` attribute from the secret, change the ```username``` attribute to ```newlogin``` and prompt for the value of the eyes-only attribute ```password```

## Delete a secret

```
$ vault delete dir/subdir/website.com
```

## Seal and unseal the vault

By default, your passphrase will always be asked interactively whenever you create, edit or delete a secret. This can quickly become cumbersome and prone to error. To mitigate this, a user can ```unseal``` his vault.

Unsealing one's vault will remember (see below) your passphrase for as long as the vault is left unsealed (or until next reboot), so that any command requiring a passphrase can use it to encrypt and decrypt data.

This works by **storing the user's hashed passphrase** in a ```0400``` file under /run/user/<uid> or /tmp.

To unseal your vault:

```
$ vault unseal
```

To seal it again:

```
$ vault seal
```

## Git integration

On vault create, it is automatically set up in a local git repository. You can link it to a remote repository like so:

```
$ vault git remote <url>
```

From now on, every change to your vault will automatically result in a commit, which you then can push with:

```
$ vault git push
```

Similarly, if the remote repository was changed for any reason (like you using your vault from another computer), you can fetch the latest and brightest with:

```
$ vault git pull
```

Remember, the directory ```$HOME/.vault``` is a regular git repository, you can used the ```git``` command as you like.
