package main

import (
	_ "embed"

	gosdkconfigcmd "github.com/bizshuk/gosdk/cmd/config"
)

// embeddedSettings is the editable configuration seed shipped with autop.
// The SDK's config default command writes it to ~/.config/autop/settings.json
// when the operator explicitly runs `autop config default`.
//
//go:embed settings.example.json
var embeddedSettings []byte

func init() {
	gosdkconfigcmd.MustRegisterDefault("settings.json", embeddedSettings)
}
