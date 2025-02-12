package pocketmine

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/datagen/data"
	"github.com/df-mc/datagen/write"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"slices"
	"strings"
)

func HandleGameData(gameData minecraft.GameData) {
	requiredItemList := make(map[string]RequiredItemEntry)
	for _, item := range gameData.Items {
		data.ItemNameToNetworkID[item.Name] = int32(item.RuntimeID)
		data.ItemNetworkIDToName[int32(item.RuntimeID)] = item.Name
		requiredItemList[item.Name] = RequiredItemEntry{
			RuntimeID:      item.RuntimeID,
			ComponentBased: item.ComponentBased,
		}
	}
	write.JSON("output/pocketmine/required_item_list.json", requiredItemList)
}

func HandleAvailableActorIdentifiers(pk *packet.AvailableActorIdentifiers) {
	var identifiers AvailableActorIdentifiers
	err := nbt.Unmarshal(pk.SerialisedEntityIdentifiers, &identifiers)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal entity identifiers: %w", err))
	}
	list := identifiers.IDList
	slices.SortFunc(list, func(a, b ActorIdentifier) int {
		return int(a.RuntimeID - b.RuntimeID)
	})
	var lines []string
	for _, id := range list {
		lines = append(lines, fmt.Sprintf("\t\"%s\": %d", id.ID, id.RuntimeID))
	}
	b := []byte(fmt.Sprintf("{\n%s\n}", strings.Join(lines, ",\n")))
	write.Raw("output/pocketmine/entity_id_map.json", b)
	write.Raw("output/pocketmine/entity_identifiers.nbt", pk.SerialisedEntityIdentifiers)
}

func HandleBiomeDefinitionList(pk *packet.BiomeDefinitionList) {
	var biomes map[string]any
	err := nbt.Unmarshal(pk.SerialisedBiomeDefinitions, &biomes)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal biome definitions: %w", err))
	}
	for _, v := range biomes {
		m := v.(map[string]any)
		keys := []string{
			"minecraft:capped_surface",
			"minecraft:consolidated_features",
			"minecraft:frozen_ocean_surface",
			"minecraft:legacy_world_generation_rules",
			"minecraft:mesa_surface",
			"minecraft:mountain_parameters",
			"minecraft:multinoise_generation_rules",
			"minecraft:overworld_generation_rules",
			"minecraft:surface_material_adjustments",
			"minecraft:surface_parameters",
			"minecraft:swamp_surface",
		}
		for _, key := range keys {
			delete(m, key)
		}
	}
	write.NBT("output/pocketmine/biome_definitions.nbt", biomes)
	write.Raw("output/pocketmine/biome_definitions_full.nbt", pk.SerialisedBiomeDefinitions)
}

func HandleCraftingData(pk *packet.CraftingData) {
	recipes := make(map[string][]any)
	for _, recipe := range pk.Recipes {
		var key string
		var value any
		switch r := recipe.(type) {
		case *protocol.ShapelessRecipe:
			key = "shapeless_crafting"
			value = shapelessRecipeData(r)
		case *protocol.ShapedRecipe:
			key = "shaped_crafting"
			if !r.AssumeSymmetry {
				key += "_asymmetric"
			}
			value = shapedRecipeData(r)
		case *protocol.FurnaceRecipe:
			key = "smelting"
			value = furnaceRecipeData(r)
		case *protocol.FurnaceDataRecipe:
			key = "smelting"
			value = furnaceRecipeData(&r.FurnaceRecipe)
		case *protocol.MultiRecipe:
			key = "special_hardcoded"
			value = r.UUID.String()
		case *protocol.ShulkerBoxRecipe:
			key = "shapeless_shulker_box"
			value = shapelessRecipeData(&r.ShapelessRecipe)
		case *protocol.ShapelessChemistryRecipe:
			key = "shapeless_chemistry"
			value = shapelessRecipeData(&r.ShapelessRecipe)
		case *protocol.ShapedChemistryRecipe:
			key = "shaped_chemistry"
			if !r.AssumeSymmetry {
				key += "_asymmetric"
			}
			value = shapedRecipeData(&r.ShapedRecipe)
		case *protocol.SmithingTransformRecipe:
			key = "smithing"
			value = smithingTransformRecipeData(r)
		case *protocol.SmithingTrimRecipe:
			key = "smithing_trim"
			value = smithingTrimRecipeData(r)
		default:
			panic(fmt.Errorf("unknown recipe type %T", r))
		}
		recipes[key] = append(recipes[key], value)
	}
	for _, r := range pk.PotionRecipes {
		recipes["potion_type"] = append(recipes["potion_type"], potionTypeRecipeData(r))
	}
	for _, r := range pk.PotionContainerChangeRecipes {
		recipes["potion_container_change"] = append(recipes["potion_container_change"], potionContainerChangeRecipeData(r))
	}

	type keyValue struct {
		k string
		v any
	}
	for name, entries := range recipes {
		var sorted []keyValue
		seen := make(map[string]int)
		for _, entry := range entries {
			data, _ := json.Marshal(entry)
			key := string(data)
			dupe, _ := seen[key]
			seen[key] = dupe + 1
			suffix := string('a' + rune(dupe))
			sorted = append(sorted, keyValue{key + suffix, entry})
		}
		slices.SortFunc(sorted, func(a, b keyValue) int {
			return strings.Compare(a.k, b.k)
		})
		recipes[name] = mapSlice(sorted, func(kv keyValue) any {
			return kv.v
		})
		for key, count := range seen {
			if count > 1 {
				fmt.Printf("warning: %s recipe %s was seen %d times\n", name, key, count)
			}
		}
	}
	for k, v := range recipes {
		write.JSON(fmt.Sprintf("output/pocketmine/recipes/%s.json", k), v)
	}
}

func HandleCreativeContent(pk *packet.CreativeContent) {
	var content CreativeItems
	for _, group := range pk.Groups {
		content.Groups = append(content.Groups, CreativeGroup{
			CategoryID:   group.Category,
			CategoryName: group.Name,
			Icon:         itemStackData(group.Icon),
		})
	}
	for _, item := range pk.Items {
		content.Items = append(content.Items, CreativeItem{
			GroupID: item.GroupIndex,
			Item:    itemStackData(item.Item),
		})
	}
	write.JSON("output/pocketmine/creativeitems.json", content)
}

func mapSlice[A, B any](s []A, f func(A) B) []B {
	var r []B
	for _, v := range s {
		r = append(r, f(v))
	}
	return r
}
