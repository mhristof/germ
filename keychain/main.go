package keychain

import (
	"fmt"

	"github.com/keybase/go-keychain"
	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/log"
)

type KeyChain struct {
	Service     string
	AccessGroup string
}

func (k *KeyChain) Add(name, value string) {
	item := keychain.NewGenericPassword(k.Service, name, name, []byte(value), k.AccessGroup)
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	err := keychain.AddItem(item)
	if err == keychain.ErrorDuplicateItem {
		log.WithFields(log.Fields{
			"name": name,
		}).Fatal("Duplicate secret")
	}
}

func (k *KeyChain) List() []string {
	accounts, err := keychain.GetGenericPasswordAccounts(k.Service)
	if err != nil {
		log.WithFields(log.Fields{
			"k.Service": k.Service,
		}).Fatal("Cannot retrieve the accounts")
	}

	return accounts
}

func (k *KeyChain) Delete(name string) {
	log.WithFields(log.Fields{
		"name": name,
	}).Debug("Deleting keychain object")

	err := keychain.DeleteGenericPasswordItem(k.Service, name)
	if err != nil {
		log.WithFields(log.Fields{
			"name": name,
			"err":  err,
		}).Fatal("Failed to delete")
	}
}

func (k *KeyChain) Profiles() []iterm.Profile {
	var ret []iterm.Profile
	for _, account := range k.List() {
		prof := iterm.NewProfile(fmt.Sprintf("custom/%s", account), map[string]string{})

		prof.KeyboardMap["0x61-0x80000"] = iterm.KeyboardMap{
			Action: 12,
			Text:   fmt.Sprintf("eval $(/usr/bin/security find-generic-password  -s %s -w -a %s)", k.Service, account),
		}

		ret = append(ret, *prof)

	}

	return ret
}
