package write

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func JSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(fmt.Errorf("failed to marshal data for %s: %w", path, err))
	}
	Raw(path, b)
}

func NBT(path string, v any) {
	b, err := nbt.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("failed to marshal data for %s: %w", path, err))
	}
	Raw(path, b)
}

func Raw(path string, b []byte) {
	fmt.Println("Writing", path)
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	err := os.WriteFile(path, b, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to write data to %s: %w", path, err))
	}
}
