package dragonfly

import (
	"fmt"
	"math"

	"github.com/df-mc/datagen/data"
	"github.com/df-mc/datagen/write"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func HandleGameData(gameData minecraft.GameData) {
	vanillaItems := make(map[string]VanillaItemEntry)
	for _, item := range gameData.Items {
		data.ItemNameToNetworkID[item.Name] = int32(item.RuntimeID)
		data.ItemNetworkIDToName[int32(item.RuntimeID)] = item.Name
		vanillaItems[item.Name] = VanillaItemEntry{
			RuntimeID:      int32(item.RuntimeID),
			ComponentBased: item.ComponentBased,
			Version:        item.Version,
			Data:           item.Data,
		}
	}
	write.NBT("output/dragonfly/server/world/vanilla_items.nbt", vanillaItems)
}

func HandleCraftingData(pk *packet.CraftingData) {
	var (
		furnace []FurnaceRecipe
		shaped  []ShapedRecipe
		potions []PotionRecipe

		shapeless, smithing, smithingTrim []ShapelessRecipe
		potionContainerChanges            []PotionContainerChangeRecipe
	)
	for _, recipe := range pk.Recipes {
		switch recipe := recipe.(type) {
		case *protocol.FurnaceRecipe:
			furnace = append(furnace, NewFurnaceRecipe(*recipe))
		case *protocol.FurnaceDataRecipe:
			furnace = append(furnace, NewFurnaceRecipe(recipe.FurnaceRecipe))
		case *protocol.ShapelessRecipe:
			shapeless = append(shapeless, NewShapelessRecipe(*recipe))
		case *protocol.ShapedRecipe:
			shaped = append(shaped, NewShapedRecipe(*recipe))
		case *protocol.SmithingTransformRecipe:
			smithing = append(smithing, NewShapelessRecipe(protocol.ShapelessRecipe{
				Input:  []protocol.ItemDescriptorCount{recipe.Base, recipe.Addition, recipe.Template},
				Output: []protocol.ItemStack{recipe.Result},
				Block:  recipe.Block,
			}))
		case *protocol.SmithingTrimRecipe:
			smithingTrim = append(smithingTrim, NewShapelessRecipe(protocol.ShapelessRecipe{
				Input: []protocol.ItemDescriptorCount{recipe.Base, recipe.Addition, recipe.Template},
				Block: recipe.Block,
			}))
		}
	}
	for _, recipe := range pk.PotionRecipes {
		potions = append(potions, NewPotionRecipe(recipe))
	}
	for _, recipe := range pk.PotionContainerChangeRecipes {
		potionContainerChanges = append(potionContainerChanges, NewPotionContainerChangeRecipe(recipe))
	}
	write.NBT("output/dragonfly/server/item/recipe/furnace_data.nbt", furnace)
	write.NBT("output/dragonfly/server/item/recipe/crafting_data.nbt", CraftingRecipes{Shaped: shaped, Shapeless: shapeless})
	write.NBT("output/dragonfly/server/item/recipe/smithing_data.nbt", smithing)
	write.NBT("output/dragonfly/server/item/recipe/smithing_trim_data.nbt", smithingTrim)
	write.NBT("output/dragonfly/server/item/recipe/potion_data.nbt", PotionRecipes{Potions: potions, ContainerChanges: potionContainerChanges})
}

func HandleCreativeContent(pk *packet.CreativeContent) {
	var groups []CreativeGroup
	var items []CreativeItem
	for _, group := range pk.Groups {
		groups = append(groups, CreativeGroup{
			Category: group.Category,
			Name:     group.Name,
			Icon:     creativeItemFromStack(group.Icon),
		})
	}
	for _, entry := range pk.Items {
		ci := creativeItemFromStack(entry.Item)
		ci.GroupIndex = int32(entry.GroupIndex)
		items = append(items, ci)
	}
	write.NBT("output/dragonfly/server/item/creative/creative_items.nbt", CreativeContent{groups, items})
}

func creativeItemFromStack(s protocol.ItemStack) CreativeItem {
	ci := CreativeItem{
		Name: data.ItemNetworkIDToName[s.ItemType.NetworkID],
		Meta: int16(s.ItemType.MetadataValue),
		NBT:  s.NBTData,
	}
	if ci.Meta == math.MaxInt16 {
		ci.Meta = 0
	}
	if s.BlockRuntimeID > 0 {
		if ci.Meta != 0 {
			panic(fmt.Errorf("block item %s has non-zero metadata %d", ci.Name, ci.Meta))
		}
		_, props, ok := chunk.RuntimeIDToState(uint32(s.BlockRuntimeID))
		if ok {
			ci.BlockProperties = props
		} else {
			panic(fmt.Errorf("failed to get block properties for item %s with runtime ID %d", ci.Name, s.BlockRuntimeID))
		}
	}
	return ci
}
