package keychain

import (
	"fmt"

	"github.com/keybase/go-keychain"
	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
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
		log.Fatal().Err(err).Str("name", name).Msg("duplicate secret")
	}
}

func (k *KeyChain) List() []string {
	accounts, err := keychain.GetGenericPasswordAccounts(k.Service)
	if err != nil {
		log.Fatal().Err(err).Str("k.Service", k.Service).Msg("cannot retrive the accounts")
	}

	return accounts
}

func (k *KeyChain) Delete(name string) {
	log.Debug().Str("name", name).Msg("deleting keychain object")

	err := keychain.DeleteGenericPasswordItem(k.Service, name)
	if err != nil {
		log.Fatal().Err(err).Str("name", name).Msg("cannot delete item")
	}
}

func (k *KeyChain) Profiles() []iterm.Profile {
	var ret []iterm.Profile
	for _, account := range k.List() {
		prof := iterm.NewProfile(fmt.Sprintf("custom/%s", account), map[string]string{})

		prof.KeyboardMap[iterm.KeyboardSortcutAltA] = iterm.KeyboardMap{
			Action: 12,
			Text:   fmt.Sprintf("eval $(/usr/bin/security find-generic-password  -s %s -w -a %s)", k.Service, account),
		}

		ret = append(ret, *prof)

	}

	return ret
}
