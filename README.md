# Data Gen

This repository is a simple tool which connects
to [BDS](https://www.minecraft.net/en-us/download/server/bedrock)
using [gophertunnel](https://github.com/Sandertv/gophertunnel) and generates necessary data
for [Dragonfly](https://github.com/df-mc/dragonfly) updates.

## Usage

1. Download the [latest version of BDS](https://www.minecraft.net/en-us/download/server/bedrock) and run the
   server. You will also need to generate a vanilla world with education features and any other appropriate
   experiments enabled.
2. Make sure `data/block_state_meta_map.json` and `data/canonical_block_states.nbt` are up-to-date
   from [BedrockData](https://github.com/pmmp/BedrockData) (or newly generated
   from [bds-mod-mapping](https://github.com/pmmp/bds-mod-mapping))
3. Run `go run main.go` and authenticate with Xbox if it is your first time running the tool
4. Once the data is generated, copy the required folders from `output` into the desired location

> [!NOTE]
> All `.nbt` files use the network-encoding variant of NBT.

> [!TIP]
> The tool will generate the data in a structured format, allowing you to easily copy the data into the
> respective repositories.

## Dragonfly data (output/dragonfly)

| File                                                                                                                                  | Description                                                                         |
|---------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------|
| [server/item/creative/creative_items.nbt](https://github.com/df-mc/dragonfly/blob/master/server/item/creative/creative_items.nbt)     | This file contains the creative groups and items in the vanilla order               |
| [server/item/recipe/crafting_data.nbt](https://github.com/df-mc/dragonfly/blob/master/server/item/recipe/crafting_data.nbt)           | This file contains a list of shaped and shapeless crafting recipes                  |
| [server/item/recipe/furnace_data.nbt](https://github.com/df-mc/dragonfly/blob/master/server/item/recipe/furnace_data.nbt)             | This file contains a list of furnace recipes                                        |
| [server/item/recipe/potion_data.nbt](https://github.com/df-mc/dragonfly/blob/master/server/item/recipe/potion_data.nbt)               | This file contains a list of brewing stand recipes                                  |
| [server/item/recipe/smithing_data.nbt](https://github.com/df-mc/dragonfly/blob/master/server/item/recipe/smithing_data.nbt)           | This file contains a list of recipes for the smithing table, excluding armour trims |
| [server/item/recipe/smithing_trim_data.nbt](https://github.com/df-mc/dragonfly/blob/master/server/item/recipe/smithing_trim_data.nbt) | This file contains a list of recipes for armour trims in the smithing table         |
| [server/vanilla_items.nbt](https://github.com/df-mc/dragonfly/blob/master/server/world/item_runtime_ids.nbt)                          | This file contains a list of all vanilla items with their runtime ID and version    |

## PMMP Data (output/pocketmine)

> [!NOTE]
> The ordering of recipes does not currently match the
> [existing ordering](https://github.com/pmmp/PocketMine-MP/blob/stable/tools/generate-bedrock-data-from-packets.php#L455-L475)
> for BedrockData, creating unreliable diffs if used.

| File                                                                                                                                 | Description                                                                                                 |
|--------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------|
| [biome_definitions.nbt](https://github.com/pmmp/BedrockData/blob/master/biome_definitions.nbt)                                       | This file contains a filtered version of the biome mappings obtained from the BiomeDefinitionList packet    |
| [biome_definitions_full.nbt](https://github.com/pmmp/BedrockData/blob/master/biome_definitions_full.nbt)                             | This file contains the raw biome mappings obtained from the BiomeDefinitionList packet                      |
| [creativeitems.json](https://github.com/pmmp/BedrockData/blob/master/creativeitems.json)                                             | The file contains the creative groups and items obtained from the CreativeContent packet                    |
| [entity_id_map.json](https://github.com/pmmp/BedrockData/blob/master/entity_id_map.json)                                             | This file contains a mapping of entity identifiers to their legacy, numerical IDs                           |
| [entity_identifiers.nbt](https://github.com/pmmp/BedrockData/blob/master/entity_identifiers.nbt)                                     | This file contains entity identifier mappings obtained from the AvailableActorIdentifiers packet            |
| [required_item_list.json](https://github.com/pmmp/BedrockData/blob/master/required_item_list.json)                                   | This file contains a list of items with their runtime ID and version, obtained from the ItemRegistry packet |
| [recipes/potion_container_change.json](https://github.com/pmmp/BedrockData/blob/master/recipes/potion_container_change.json)         | This file contains the brewing recipes that affect the bottle of the potion                                 |
| [recipes/potion_type.json](https://github.com/pmmp/BedrockData/blob/master/recipes/potion_type.json)                                 | This file contains the brewing recipes, excluding the container changes                                     |
| [recipes/shaped_chemistry_asymmetric.json](https://github.com/pmmp/BedrockData/blob/master/recipes/shaped_chemistry_asymmetric.json) | This file contains the shaped chemistry recipes                                                             |
| [recipes/shaped_crafting.json](https://github.com/pmmp/BedrockData/blob/master/recipes/shaped_crafting.json)                         | This file contains the shaped crafting recipes                                                              |
| [recipes/shapeless_chemistry.json](https://github.com/pmmp/BedrockData/blob/master/recipes/shapeless_chemistry.json)                 | This file contains the shapeless chemistry recipes                                                          |
| [recipes/shapeless_crafting.json](https://github.com/pmmp/BedrockData/blob/master/recipes/shapeless_crafting.json)                   | This file contains the shapeless crafting recipes                                                           |
| [recipes/shapeless_shulker_box.json](https://github.com/pmmp/BedrockData/blob/master/recipes/shapeless_shulker_box.json)             | This file contains the recipes for coloured containers                                                      |
| [recipes/smelting.json](https://github.com/pmmp/BedrockData/blob/master/recipes/smelting.json)                                       | This file contains the furnace recipes                                                                      |
| [recipes/smithing.json](https://github.com/pmmp/BedrockData/blob/master/recipes/smithing.json)                                       | This file contains the smithing table recipes, excluding armour trims                                       |
| [recipes/smithing_trim.json](https://github.com/pmmp/BedrockData/blob/master/recipes/smithing_trim.json)                             | This file contains the armour trim recipes for the smithing table                                           |
| [recipes/special_hardcoded.json](https://github.com/pmmp/BedrockData/blob/master/recipes/special_hardcoded.json)                     | This file contains the UUIDs for recipes that are hardcoded on the client                                   |