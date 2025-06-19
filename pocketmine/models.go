package pocketmine

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/df-mc/datagen/data"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CurrentBlockVersion = (1 << 24) | (21 << 16) | (20 << 8) | 6 // 18158598
)

type AvailableActorIdentifiers struct {
	IDList []ActorIdentifier `nbt:"idlist"`
}

type ActorIdentifier struct {
	BID         string `nbt:"bid"`
	HasSpawnEgg bool   `nbt:"hasspawnegg"`
	ID          string `nbt:"id"`
	RuntimeID   int32  `nbt:"rid"`
	Summonable  bool   `nbt:"summonable"`
}

type CreativeItems struct {
	Groups []CreativeGroup `json:"groups"`
	Items  []CreativeItem  `json:"items"`
}

type CreativeGroup struct {
	CategoryID   int32         `json:"category_id"`
	CategoryName string        `json:"category_name"`
	Icon         ItemStackData `json:"icon"`
}

type CreativeItem struct {
	GroupID uint32        `json:"group_id"`
	Item    ItemStackData `json:"item"`
}

type RequiredItemEntry struct {
	RuntimeID      int16 `json:"runtime_id"`
	ComponentBased bool  `json:"component_based"`
}

type ItemStackData struct {
	BlockStates string `json:"block_states,omitempty"`
	Meta        int16  `json:"meta,omitempty"`
	Name        string `json:"name,omitempty"`
	NBT         []byte `json:"nbt,omitempty"`
}

func itemStackData(s protocol.ItemStack) ItemStackData {
	stack := ItemStackData{
		Name: data.ItemNetworkIDToName[s.NetworkID],
		Meta: int16(s.MetadataValue),
	}
	if stack.Meta == math.MaxInt16 {
		stack.Meta = 0
	}
	if len(s.NBTData) > 0 {
		b, err := nbt.MarshalEncoding(s.NBTData, nbt.LittleEndian)
		if err != nil {
			panic(fmt.Errorf("failed to marshal NBT data for item %s: %w", stack.Name, err))
		}
		stack.NBT = b
	}
	if s.BlockRuntimeID > 0 {
		if stack.Meta != 0 {
			panic(fmt.Errorf("block item %s has non-zero metadata %d", stack.Name, stack.Meta))
		}
		_, props, ok := chunk.RuntimeIDToState(uint32(s.BlockRuntimeID))
		if ok {
			b, err := nbt.MarshalEncoding(props, nbt.LittleEndian)
			if err != nil {
				panic(fmt.Errorf("failed to marshal block properties for item %s: %w", stack.Name, err))
			}
			stack.BlockStates = base64.StdEncoding.EncodeToString(b)
		} else {
			panic(fmt.Errorf("failed to get block properties for item %s with runtime ID %d", stack.Name, s.BlockRuntimeID))
		}
	}
	return stack
}

type RecipeIngredientData struct {
	BlockStates      []byte `json:"block_states,omitempty"`
	Count            int32  `json:"count,omitempty"`
	Meta             int16  `json:"meta,omitempty"`
	MolangExpression string `json:"molang_expression,omitempty"`
	MolangVersion    byte   `json:"molang_version,omitempty"`
	Name             string `json:"name,omitempty"`
	Tag              string `json:"tag,omitempty"`
}

func recipeIngredientData(c protocol.ItemDescriptorCount) RecipeIngredientData {
	var ingredient RecipeIngredientData
	switch d := c.Descriptor.(type) {
	case *protocol.InvalidItemDescriptor:
		panic("invalid item descriptor")
	case *protocol.DefaultItemDescriptor:
		ingredient.Name = data.ItemNetworkIDToName[int32(d.NetworkID)]
		if d.MetadataValue == 32767 {
			_, props, ok := chunk.RuntimeIDToState(uint32(d.MetadataValue))
			if ok {
				b, err := nbt.MarshalEncoding(props, nbt.LittleEndian)
				if err != nil {
					panic(fmt.Errorf("failed to marshal block properties for item %s: %w", ingredient.Name, err))
				}
				ingredient.BlockStates = b
			} else {
				ingredient.Meta = d.MetadataValue
			}
		} else {
			ingredient.Meta = d.MetadataValue
		}
	case *protocol.MoLangItemDescriptor:
		ingredient.MolangExpression = d.Expression
		ingredient.MolangVersion = d.Version
	case *protocol.ItemTagItemDescriptor:
		ingredient.Tag = d.Tag
	case *protocol.DeferredItemDescriptor:
		ingredient.Name = d.Name
		if d.MetadataValue == 32767 {
			_, props, ok := chunk.RuntimeIDToState(uint32(d.MetadataValue))
			if ok {
				b, err := nbt.MarshalEncoding(props, nbt.LittleEndian)
				if err != nil {
					panic(fmt.Errorf("failed to marshal block properties for item %s: %w", ingredient.Name, err))
				}
				ingredient.BlockStates = b
			} else {
				ingredient.Meta = d.MetadataValue
			}
		} else {
			ingredient.Meta = d.MetadataValue
		}
	case *protocol.ComplexAliasItemDescriptor:
		ingredient.Name = d.Name
	}
	if c.Count != 1 {
		ingredient.Count = c.Count
	}
	return ingredient
}

type FurnaceRecipeData struct {
	Block  string               `json:"block"`
	Input  RecipeIngredientData `json:"input"`
	Output ItemStackData        `json:"output"`
}

func furnaceRecipeData(r *protocol.FurnaceRecipe) FurnaceRecipeData {
	return FurnaceRecipeData{
		Input: recipeIngredientData(protocol.ItemDescriptorCount{
			Count: 1,
			Descriptor: &protocol.DefaultItemDescriptor{
				NetworkID:     int16(r.InputType.NetworkID),
				MetadataValue: int16(r.InputType.MetadataValue),
			},
		}),
		Output: itemStackData(r.Output),
		Block:  r.Block,
	}
}

type PotionTypeRecipeData struct {
	Ingredient RecipeIngredientData `json:"ingredient"`
	Input      RecipeIngredientData `json:"input"`
	Output     ItemStackData        `json:"output"`
}

func potionTypeRecipeData(r protocol.PotionRecipe) PotionTypeRecipeData {
	return PotionTypeRecipeData{
		Input: recipeIngredientData(protocol.ItemDescriptorCount{
			Count: 1,
			Descriptor: &protocol.DefaultItemDescriptor{
				NetworkID:     int16(r.InputPotionID),
				MetadataValue: int16(r.InputPotionMetadata),
			},
		}),
		Ingredient: recipeIngredientData(protocol.ItemDescriptorCount{
			Count: 1,
			Descriptor: &protocol.DefaultItemDescriptor{
				NetworkID:     int16(r.ReagentItemID),
				MetadataValue: int16(r.ReagentItemMetadata),
			},
		}),
		Output: itemStackData(protocol.ItemStack{
			ItemType: protocol.ItemType{
				NetworkID:     r.OutputPotionID,
				MetadataValue: uint32(r.OutputPotionMetadata),
			},
		}),
	}
}

type PotionContainerChangeRecipeData struct {
	Ingredient     RecipeIngredientData `json:"ingredient"`
	InputItemName  string               `json:"input_item_name"`
	OutputItemName string               `json:"output_item_name"`
}

func potionContainerChangeRecipeData(r protocol.PotionContainerChangeRecipe) PotionContainerChangeRecipeData {
	return PotionContainerChangeRecipeData{
		InputItemName: data.ItemNetworkIDToName[r.InputItemID],
		Ingredient: recipeIngredientData(protocol.ItemDescriptorCount{
			Count: 1,
			Descriptor: &protocol.DefaultItemDescriptor{
				NetworkID: int16(r.ReagentItemID),
			},
		}),
		OutputItemName: data.ItemNetworkIDToName[r.OutputItemID],
	}
}

type ShapedRecipeData struct {
	Block                string                 `json:"block"`
	Input                []RecipeIngredientData `json:"input"`
	Output               []ItemStackData        `json:"output"`
	Priority             int32                  `json:"priority"`
	Shape                []string               `json:"shape"`
	UnlockingIngredients []RecipeIngredientData `json:"unlockingIngredients,omitempty"`
}

func shapedRecipeData(r *protocol.ShapedRecipe) ShapedRecipeData {
	var inputs []protocol.ItemDescriptorCount
	shape := make([][]string, r.Width)
	keys := make(map[string]string)
	char := 'A'

	for x := 0; x < int(r.Width); x++ {
		shape[x] = make([]string, r.Height)
		for y := 0; y < int(r.Height); y++ {
			ingredient := r.Input[y*int(r.Width)+x]
			if _, ok := ingredient.Descriptor.(*protocol.InvalidItemDescriptor); ok {
				shape[x][y] = " "
				continue
			}
			hash, _ := json.Marshal(recipeIngredientData(ingredient))
			if k, ok := keys[string(hash)]; ok {
				shape[x][y] = k
				continue
			}
			k := string(char)
			shape[x][y] = k
			keys[string(hash)] = k
			inputs = append(inputs, ingredient)
			char++
		}
	}

	return ShapedRecipeData{
		Shape: mapSlice(shape, func(s []string) string {
			return strings.Join(s, "")
		}),
		Input:                mapSlice(inputs, recipeIngredientData),
		Output:               mapSlice(r.Output, itemStackData),
		Block:                r.Block,
		Priority:             r.Priority,
		UnlockingIngredients: mapSlice(r.UnlockRequirement.Ingredients, recipeIngredientData),
	}
}

type ShapelessRecipeData struct {
	Block                string                 `json:"block"`
	Input                []RecipeIngredientData `json:"input"`
	Output               []ItemStackData        `json:"output"`
	Priority             int32                  `json:"priority"`
	UnlockingIngredients []RecipeIngredientData `json:"unlockingIngredients,omitempty"`
}

func shapelessRecipeData(r *protocol.ShapelessRecipe) ShapelessRecipeData {
	return ShapelessRecipeData{
		Input:                mapSlice(r.Input, recipeIngredientData),
		Output:               mapSlice(r.Output, itemStackData),
		Block:                r.Block,
		Priority:             r.Priority,
		UnlockingIngredients: mapSlice(r.UnlockRequirement.Ingredients, recipeIngredientData),
	}
}

type SmithingTransformRecipeData struct {
	Addition RecipeIngredientData `json:"addition"`
	Block    string               `json:"block"`
	Input    RecipeIngredientData `json:"input"`
	Output   ItemStackData        `json:"output"`
	Template RecipeIngredientData `json:"template"`
}

func smithingTransformRecipeData(r *protocol.SmithingTransformRecipe) SmithingTransformRecipeData {
	return SmithingTransformRecipeData{
		Template: recipeIngredientData(r.Template),
		Input:    recipeIngredientData(r.Base),
		Addition: recipeIngredientData(r.Addition),
		Output:   itemStackData(r.Result),
		Block:    r.Block,
	}
}

type SmithingTrimRecipeData struct {
	Addition RecipeIngredientData `json:"addition"`
	Block    string               `json:"block"`
	Input    RecipeIngredientData `json:"input"`
	Template RecipeIngredientData `json:"template"`
}

func smithingTrimRecipeData(r *protocol.SmithingTrimRecipe) SmithingTrimRecipeData {
	return SmithingTrimRecipeData{
		Template: recipeIngredientData(r.Template),
		Input:    recipeIngredientData(r.Base),
		Addition: recipeIngredientData(r.Addition),
		Block:    r.Block,
	}
}

type Colour struct {
	A uint8 `json:"a"`
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

func int32ToRGBA(x int32) Colour {
	return Colour{
		A: uint8((x >> 24) & 0xFF),
		R: uint8((x >> 16) & 0xFF),
		G: uint8((x >> 8) & 0xFF),
		B: uint8(x & 0xFF),
	}
}

type BiomeDefinition struct {
	BiomeID uint16 `json:"id,omitempty"`

	Temperature      float32 `json:"temperature"`
	Downfall         float32 `json:"downfall"`
	RedSporeDensity  float32 `json:"redSporeDensity"`
	BlueSporeDensity float32 `json:"blueSporeDensity"`
	AshDensity       float32 `json:"ashDensity"`
	WhiteAshDensity  float32 `json:"whiteAshDensity"`

	Depth          float32 `json:"depth"`
	Scale          float32 `json:"scale"`
	MapWaterColour Colour  `json:"mapWaterColour"`

	Rain bool     `json:"rain"`
	Tags []string `json:"tags"`
}

func newBiomeDefinition(definition protocol.BiomeDefinition, list []string) BiomeDefinition {
	var biomeID uint16
	if v, ok := definition.BiomeID.Value(); ok {
		biomeID = v
	}
	var tags []string
	if v, ok := definition.Tags.Value(); ok {
		tags = lo.Map(v, func(i uint16, _ int) string {
			return list[i]
		})
	}
	return BiomeDefinition{
		BiomeID:          biomeID,
		Temperature:      definition.Temperature,
		Downfall:         definition.Downfall,
		RedSporeDensity:  definition.RedSporeDensity,
		BlueSporeDensity: definition.BlueSporeDensity,
		AshDensity:       definition.AshDensity,
		WhiteAshDensity:  definition.WhiteAshDensity,
		Depth:            definition.Depth,
		Scale:            definition.Scale,
		MapWaterColour:   int32ToRGBA(definition.MapWaterColour),
		Rain:             definition.Rain,
		Tags:             tags,
	}
}
