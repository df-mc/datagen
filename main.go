package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/df-mc/datagen/dragonfly"
	"github.com/df-mc/datagen/pocketmine"
	_ "github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
)

func main() {
	_ = os.RemoveAll("output")

	dialer := minecraft.Dialer{
		TokenSource: tokenSource(),
	}
	conn, err := dialer.Dial("raknet", "127.0.0.1:19132")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.DoSpawn(); err != nil {
		panic(err)
	}
	dragonfly.HandleGameData(conn.GameData())
	pocketmine.HandleGameData(conn.GameData())

	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			break
		}

		switch p := pk.(type) {
		case *packet.AvailableActorIdentifiers:
			pocketmine.HandleAvailableActorIdentifiers(p)
		case *packet.BiomeDefinitionList:
			dragonfly.HandleBiomeDefinitionList(p)
			pocketmine.HandleBiomeDefinitionList(p)
		case *packet.CraftingData:
			dragonfly.HandleCraftingData(p)
			pocketmine.HandleCraftingData(p)
		case *packet.CreativeContent:
			dragonfly.HandleCreativeContent(p)
			pocketmine.HandleCreativeContent(p)
		}
	}
}

// tokenSource returns a token source for using with a gophertunnel client. It either reads it from the
// token.tok file if cached or requests logging in with a device code.
func tokenSource() oauth2.TokenSource {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	token := new(oauth2.Token)
	tokenData, err := os.ReadFile("token.tok")
	if err == nil {
		_ = json.Unmarshal(tokenData, token)
	} else {
		token, err = auth.RequestLiveToken()
		check(err)
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		token, err = auth.RequestLiveToken()
		check(err)
		src = auth.RefreshTokenSource(token)
	}
	go func() {
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c

		tok, _ := src.Token()
		b, _ := json.Marshal(tok)
		_ = os.WriteFile("token.tok", b, 0644)
		os.Exit(0)
	}()
	return src
}
