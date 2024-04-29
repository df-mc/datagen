package main

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/datagen/data"
	"github.com/df-mc/datagen/models"
	_ "github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

	requiredItemList := make(map[string]models.RequiredItemEntry)
	for _, item := range conn.GameData().Items {
		data.ItemNameToNetworkID[item.Name] = int32(item.RuntimeID)
		data.ItemNetworkIDToName[int32(item.RuntimeID)] = item.Name
		requiredItemList[item.Name] = models.RequiredItemEntry{
			RuntimeID:      item.RuntimeID,
			ComponentBased: item.ComponentBased,
		}
	}
	writeNBT("output/dragonfly/server/world/item_runtime_ids.nbt", data.ItemNameToNetworkID)

	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			break
		}

		switch p := pk.(type) {
		case *packet.CraftingData:
			var shaped []models.ShapedRecipe
			var shapeless, smithing, smithingTrim []models.ShapelessRecipe
			for _, recipe := range p.Recipes {
				switch recipe := recipe.(type) {
				case *protocol.ShapelessRecipe:
					shapeless = append(shapeless, models.NewShapelessRecipe(*recipe))
				case *protocol.ShapedRecipe:
					shaped = append(shaped, models.NewShapedRecipe(*recipe))
				case *protocol.SmithingTransformRecipe:
					smithing = append(smithing, models.NewShapelessRecipe(protocol.ShapelessRecipe{
						Input:  []protocol.ItemDescriptorCount{recipe.Base, recipe.Addition, recipe.Template},
						Output: []protocol.ItemStack{recipe.Result},
						Block:  recipe.Block,
					}))
				case *protocol.SmithingTrimRecipe:
					smithingTrim = append(smithingTrim, models.NewShapelessRecipe(protocol.ShapelessRecipe{
						Input: []protocol.ItemDescriptorCount{recipe.Base, recipe.Addition, recipe.Template},
						Block: recipe.Block,
					}))
				}
			}
			writeNBT("output/dragonfly/server/item/recipe/crafting_data.nbt", models.CraftingRecipes{Shaped: shaped, Shapeless: shapeless})
			writeNBT("output/dragonfly/server/item/recipe/smithing_data.nbt", smithing)
			writeNBT("output/dragonfly/server/item/recipe/smithing_trim_data.nbt", smithingTrim)
		case *packet.CreativeContent:
			var entries []models.CreativeItem
			for _, c := range p.Items {
				item := c.Item
				entry := models.CreativeItem{
					Name: data.ItemNetworkIDToName[item.ItemType.NetworkID],
					Meta: int16(item.ItemType.MetadataValue),
					NBT:  item.NBTData,
				}
				if entry.Meta == math.MaxInt16 {
					entry.Meta = 0
				}
				if item.BlockRuntimeID > 0 {
					if entry.Meta != 0 {
						panic(fmt.Errorf("block item %s has non-zero metadata %d", entry.Name, entry.Meta))
					}
					_, props, ok := chunk.RuntimeIDToState(uint32(item.BlockRuntimeID))
					if ok {
						entry.BlockProperties = props
					} else {
						panic(fmt.Errorf("failed to get block properties for item %s with runtime ID %d", entry.Name, item.BlockRuntimeID))
					}
				}
				entries = append(entries, entry)
			}
			writeNBT("output/dragonfly/server/item/creative/creative_items.nbt", entries)
		}
	}
}

func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal data for %s: %w", path, err))
	}
	writeRaw(path, b)
}

func writeNBT(path string, v any) {
	b, err := nbt.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("failed to marshal data for %s: %w", path, err))
	}
	writeRaw(path, b)
}

func writeRaw(path string, b []byte) {
	fmt.Println("Writing", path)
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	err := os.WriteFile(path, b, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to write data to %s: %w", path, err))
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
