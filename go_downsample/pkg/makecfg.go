package dsamp

import (
	"fmt"
	"encoding/json"
	"os"
	"regexp"
)

// Generate config file for downsampling
func MakeCfg() {
	paths := GetPaths(os.Stdin)
	cfg := DownsampleArgs{}
	cfg.Countspath = "counts.txt"
	re := regexp.MustCompile(`\.[^.]*$`)
	for i, path := range paths {
		cfg.IoSets = append(cfg.IoSets, IoSet{
			Inpath: path,
			Outpath: re.ReplaceAllString(path, "_downsampled$0"),
			Seed: int64(i),
		})
	}
	out, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
