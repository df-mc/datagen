package data

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

var (
	//go:embed canonical_block_states.nbt
	blockPaletteData []byte
	//go:embed block_state_meta_map.json
	metaMapData []byte
)

var (
	ItemNameToNetworkID = make(map[string]int32)
	ItemNetworkIDToName = make(map[int32]string)

	ItemMetaToBlockState = make(map[string]map[int32]map[string]any)
)

func init() {
	var metaMap []int32
	err := json.Unmarshal(metaMapData, &metaMap)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal block_state_meta_map.json: %w", err))
	}

	buf := bytes.NewBuffer(blockPaletteData)
	decoder := nbt.NewDecoder(buf)
	var i int
	for buf.Len() > 0 {
		var state map[string]any
		err = decoder.Decode(&state)
		if err != nil {
			panic(fmt.Errorf("failed to unmarshal canonical_block_states.nbt: %w", err))
		} else if i >= len(metaMap) {
			panic(fmt.Errorf("meta map does not contain meta value for state: %v", state))
		}
		name := state["name"].(string)
		meta := metaMap[i]
		if m, ok := ItemMetaToBlockState[name]; ok {
			m[meta] = state
		} else {
			ItemMetaToBlockState[name] = map[int32]map[string]any{meta: state}
		}
		i++
	}
}
