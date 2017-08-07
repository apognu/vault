# Vault

Personal project of a git-based encrypted password store. Each "password" is actually a JSON-encoded map of strings, encrypted through AES-GCM with a key derived from a passphrase, allowing to store more than just one password for each entry.

The whole store directory is located under $HOME/.vault (overridable through the VAULT_PATH environment variable) and can be pushed to a remote git repository though the ```vault``` command.

**This is a draft in progress, I would be very cautious with using it to store your most precious passwords.**

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

This works by very uglily **storing the user's hashed passphrase** in a ```0400``` file under /tmp. I know, this sucks, and will probably be enhanced in future versions. But we're stuck with it for now.

To unseal your vault:

```
$ vault unseal
```

To seal it again:

```
$ vault seal
```

## Git integration

You can enable git integration by running the following command:

```
$ vault git init
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
