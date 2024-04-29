package models

import (
	"fmt"
	"github.com/df-mc/datagen/data"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

const (
	CurrentBlockVersion = (1 << 24) | (20 << 16) | (80 << 8) | 3 // 18108419
)

// CraftingRecipes represents the structure for crafting_data.nbt that dragonfly uses.
type CraftingRecipes struct {
	Shaped    []ShapedRecipe    `nbt:"shaped"`
	Shapeless []ShapelessRecipe `nbt:"shapeless"`
}

// CreativeItem represents the structure of a creative item that dragonfly reads from creative_items.nbt.
type CreativeItem struct {
	Name            string         `nbt:"name"`
	Meta            int16          `nbt:"meta,omitempty"`
	NBT             map[string]any `nbt:"nbt,omitempty"`
	BlockProperties map[string]any `nbt:"block_properties,omitempty"`
}

// RecipeInputItem represents the structure of an input item in a recipe.
type RecipeInputItem struct {
	Name  string         `nbt:"name,omitempty"`
	Meta  int32          `nbt:"meta,omitempty"`
	Count int32          `nbt:"count"`
	State map[string]any `nbt:"block,omitempty"`
	Tag   string         `nbt:"tag,omitempty"`
}

// RecipeOutputItem represents the structure of an output item in a recipe.
type RecipeOutputItem struct {
	Name    string         `nbt:"name"`
	Meta    int32          `nbt:"meta,omitempty"`
	Count   int16          `nbt:"count"`
	State   map[string]any `nbt:"block,omitempty"`
	NBTData map[string]any `nbt:"data,omitempty"`
}

// ShapedRecipe represents the structure of a shaped recipe in dragonfly, used in crafting_data.nbt.
type ShapedRecipe struct {
	Input    []RecipeInputItem  `nbt:"input,omitempty"`
	Output   []RecipeOutputItem `nbt:"output,omitempty"`
	Block    string             `nbt:"block,omitempty"`
	Width    int32              `nbt:"width,omitempty"`
	Height   int32              `nbt:"height,omitempty"`
	Priority int32              `nbt:"priority,omitempty"`
}

// NewShapedRecipe creates a new ShapedRecipe from a protocol.ShapedRecipe. It converts the input and output
// items to the RecipeInputItem and RecipeOutputItem structures.
func NewShapedRecipe(recipe protocol.ShapedRecipe) ShapedRecipe {
	var input []RecipeInputItem
	for _, item := range recipe.Input {
		input = append(input, newInputItem(item, true))
	}
	var output []RecipeOutputItem
	for _, item := range recipe.Output {
		output = append(output, newOutputItem(item))
	}
	return ShapedRecipe{
		Input:    input,
		Output:   output,
		Block:    recipe.Block,
		Width:    recipe.Width,
		Height:   recipe.Height,
		Priority: recipe.Priority,
	}

}

// ShapelessRecipe represents the structure of a shapeless recipe in dragonfly, used in crafting_data.nbt but
// also in smithing_data.nbt and smithing_trim_data.nbt.
type ShapelessRecipe struct {
	Input    []RecipeInputItem  `nbt:"input,omitempty"`
	Output   []RecipeOutputItem `nbt:"output,omitempty"`
	Block    string             `nbt:"block,omitempty"`
	Priority int32              `nbt:"priority,omitempty"`
}

// NewShapelessRecipe creates a new ShapelessRecipe from a protocol.ShapelessRecipe. It converts the input and
// output items to the RecipeInputItem and RecipeOutputItem structures.
func NewShapelessRecipe(recipe protocol.ShapelessRecipe) ShapelessRecipe {
	var input []RecipeInputItem
	for _, item := range recipe.Input {
		input = append(input, newInputItem(item, false))
	}
	var output []RecipeOutputItem
	for _, item := range recipe.Output {
		output = append(output, newOutputItem(item))
	}
	return ShapelessRecipe{
		Input:    input,
		Output:   output,
		Block:    recipe.Block,
		Priority: recipe.Priority,
	}
}

// newInputItem returns a new RecipeInputItem from an ItemDescriptorCount. If includeAir is true, the item
// will return an air item if the descriptor is invalid. If includeAir is false, the function will panic if
// the descriptor is invalid.
func newInputItem(input protocol.ItemDescriptorCount, includeAir bool) RecipeInputItem {
	item := RecipeInputItem{Count: input.Count}
	switch it := input.Descriptor.(type) {
	case *protocol.InvalidItemDescriptor:
		if includeAir {
			return RecipeInputItem{Name: "minecraft:air"}
		}
		panic("invalid item descriptor")
	case *protocol.DefaultItemDescriptor:
		item.Name = data.ItemNetworkIDToName[int32(it.NetworkID)]
		item.Meta = int32(it.MetadataValue)
	case *protocol.MoLangItemDescriptor:
		panic("unsupported molang item descriptor")
	case *protocol.ItemTagItemDescriptor:
		item.Tag = it.Tag
	case *protocol.DeferredItemDescriptor:
		item.Name = it.Name
		item.Meta = int32(it.MetadataValue)
	case *protocol.ComplexAliasItemDescriptor:
		item.Name = it.Name
	default:
		panic(fmt.Errorf("unknown item descriptor %T", it))
	}
	if item.Meta == int32(math.MaxInt16) {
		return item
	}
	if itemMetas, ok := data.ItemMetaToBlockState[item.Name]; ok {
		if state, ok := itemMetas[item.Meta]; ok {
			item.Meta = 0
			item.State = state
		}
	}
	return item
}

// newOutputItem returns a new RecipeOutputItem from an ItemStack. It converts the ItemStack to a
// RecipeOutputItem, setting the name, meta, count and NBT data.
func newOutputItem(output protocol.ItemStack) RecipeOutputItem {
	item := RecipeOutputItem{
		Name:    data.ItemNetworkIDToName[output.NetworkID],
		Meta:    int32(output.MetadataValue),
		Count:   int16(output.Count),
		NBTData: output.NBTData,
	}
	name, props, ok := chunk.RuntimeIDToState(uint32(output.BlockRuntimeID))
	if ok {
		if itemMetas, ok := data.ItemMetaToBlockState[item.Name]; ok {
			if _, ok := itemMetas[item.Meta]; ok {
				item.Meta = 0
				item.State = map[string]any{
					"name":    name,
					"states":  props,
					"version": int32(CurrentBlockVersion),
				}
			}
		}
	}
	return item
}
