package main

// ngto -> schematic 変換ロジック (Python版 ngto2schematic.py と同じ内容)

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ---------------- ブロックID表 ----------------
var ids1710, ids112 map[string]int

var colors = strings.Fields("white orange magenta light_blue yellow lime pink gray silver cyan purple blue brown green red black")

func init() {
	v := strings.Fields(`air stone grass dirt cobblestone planks sapling bedrock flowing_water water
flowing_lava lava sand gravel gold_ore iron_ore coal_ore log leaves sponge glass lapis_ore
lapis_block dispenser sandstone noteblock bed golden_rail detector_rail sticky_piston web
tallgrass deadbush piston piston_head wool piston_extension yellow_flower red_flower
brown_mushroom red_mushroom gold_block iron_block double_stone_slab stone_slab brick_block
tnt bookshelf mossy_cobblestone obsidian torch fire mob_spawner oak_stairs chest
redstone_wire diamond_ore diamond_block crafting_table wheat farmland furnace lit_furnace
standing_sign wooden_door ladder rail stone_stairs wall_sign lever stone_pressure_plate
iron_door wooden_pressure_plate redstone_ore lit_redstone_ore unlit_redstone_torch
redstone_torch stone_button snow_layer ice snow cactus clay reeds jukebox fence pumpkin
netherrack soul_sand glowstone portal lit_pumpkin cake unpowered_repeater powered_repeater
stained_glass trapdoor monster_egg stonebrick brown_mushroom_block red_mushroom_block
iron_bars glass_pane melon_block pumpkin_stem melon_stem vine fence_gate brick_stairs
stone_brick_stairs mycelium waterlily nether_brick nether_brick_fence nether_brick_stairs
nether_wart enchanting_table brewing_stand cauldron end_portal end_portal_frame end_stone
dragon_egg redstone_lamp lit_redstone_lamp double_wooden_slab wooden_slab cocoa
sandstone_stairs emerald_ore ender_chest tripwire_hook tripwire emerald_block spruce_stairs
birch_stairs jungle_stairs command_block beacon cobblestone_wall flower_pot carrots potatoes
wooden_button skull anvil trapped_chest light_weighted_pressure_plate
heavy_weighted_pressure_plate unpowered_comparator powered_comparator daylight_detector
redstone_block quartz_ore hopper quartz_block quartz_stairs activator_rail dropper
stained_hardened_clay stained_glass_pane leaves2 log2 acacia_stairs dark_oak_stairs`)
	ids1710 = map[string]int{}
	for i, n := range v {
		ids1710["minecraft:"+n] = i
	}
	for n, i := range map[string]int{"hay_block": 170, "carpet": 171, "hardened_clay": 172,
		"coal_block": 173, "packed_ice": 174, "double_plant": 175} {
		ids1710["minecraft:"+n] = i
	}
	ids112 = map[string]int{}
	for k, x := range ids1710 {
		ids112[k] = x
	}
	for i, n := range strings.Fields("slime barrier iron_trapdoor prismarine sea_lantern") {
		ids112["minecraft:"+n] = 165 + i
	}
	for i, n := range strings.Fields(`standing_banner wall_banner daylight_detector_inverted red_sandstone red_sandstone_stairs
double_stone_slab2 stone_slab2 spruce_fence_gate birch_fence_gate jungle_fence_gate
dark_oak_fence_gate acacia_fence_gate spruce_fence birch_fence jungle_fence dark_oak_fence
acacia_fence spruce_door birch_door jungle_door acacia_door dark_oak_door end_rod chorus_plant
chorus_flower purpur_block purpur_pillar purpur_stairs purpur_double_slab purpur_slab end_bricks
beetroots grass_path end_gateway repeating_command_block chain_command_block frosted_ice magma
nether_wart_block red_nether_brick bone_block structure_void observer`) {
		ids112["minecraft:"+n] = 176 + i
	}
	for i, c := range colors {
		ids112["minecraft:"+c+"_shulker_box"] = 219 + i
		ids112["minecraft:"+c+"_glazed_terracotta"] = 235 + i
	}
	ids112["minecraft:concrete"] = 251
	ids112["minecraft:concrete_powder"] = 252
	ids112["minecraft:structure_block"] = 255

	// 木の種類違い・色違いの置き換えを追加
	for _, w := range []string{"spruce", "birch", "jungle", "dark_oak", "acacia"} {
		subst1710["minecraft:"+w+"_fence"] = sub{"minecraft:fence", zero}
		subst1710["minecraft:"+w+"_fence_gate"] = sub{"minecraft:fence_gate", keep}
		subst1710["minecraft:"+w+"_door"] = sub{"minecraft:wooden_door", keep}
	}
	for i, c := range colors {
		i := i
		subst1710["minecraft:"+c+"_shulker_box"] = sub{"minecraft:wool", func(int) int { return i }}
		subst1710["minecraft:"+c+"_glazed_terracotta"] = sub{"minecraft:stained_hardened_clay", func(int) int { return i }}
	}
}

// ---------------- 1.7.10 に無いブロックの置き換え ----------------
type sub struct {
	name string
	meta func(int) int
}

func keep(m int) int          { return m }
func zero(int) int            { return 0 }
func cst(v int) func(int) int { return func(int) int { return v } }
func pillar(m int) int {
	switch m {
	case 4:
		return 3
	case 8:
		return 4
	}
	return 2
}

var subst1710 = map[string]sub{
	"minecraft:iron_trapdoor":              {"minecraft:trapdoor", keep},
	"minecraft:prismarine":                 {"minecraft:stonebrick", zero},
	"minecraft:sea_lantern":                {"minecraft:glowstone", zero},
	"minecraft:slime":                      {"minecraft:wool", cst(5)},
	"minecraft:barrier":                    {"minecraft:air", zero},
	"minecraft:standing_banner":            {"minecraft:air", zero},
	"minecraft:wall_banner":                {"minecraft:air", zero},
	"minecraft:daylight_detector_inverted": {"minecraft:daylight_detector", keep},
	"minecraft:red_sandstone":              {"minecraft:sandstone", keep},
	"minecraft:red_sandstone_stairs":       {"minecraft:sandstone_stairs", keep},
	"minecraft:double_stone_slab2":         {"minecraft:double_stone_slab", cst(1)},
	"minecraft:stone_slab2":                {"minecraft:stone_slab", func(m int) int { return (m & 8) | 1 }},
	"minecraft:end_rod":                    {"minecraft:iron_bars", zero},
	"minecraft:chorus_plant":               {"minecraft:air", zero},
	"minecraft:chorus_flower":              {"minecraft:air", zero},
	"minecraft:purpur_block":               {"minecraft:quartz_block", zero},
	"minecraft:purpur_pillar":              {"minecraft:quartz_block", pillar},
	"minecraft:purpur_stairs":              {"minecraft:quartz_stairs", keep},
	"minecraft:purpur_double_slab":         {"minecraft:double_stone_slab", cst(7)},
	"minecraft:purpur_slab":                {"minecraft:stone_slab", func(m int) int { return (m & 8) | 7 }},
	"minecraft:end_bricks":                 {"minecraft:end_stone", zero},
	"minecraft:beetroots":                  {"minecraft:carrots", func(m int) int { return min(m*2, 7) }},
	"minecraft:grass_path":                 {"minecraft:grass", zero},
	"minecraft:end_gateway":                {"minecraft:air", zero},
	"minecraft:repeating_command_block":    {"minecraft:command_block", zero},
	"minecraft:chain_command_block":        {"minecraft:command_block", zero},
	"minecraft:frosted_ice":                {"minecraft:ice", zero},
	"minecraft:magma":                      {"minecraft:netherrack", zero},
	"minecraft:nether_wart_block":          {"minecraft:wool", cst(14)},
	"minecraft:red_nether_brick":           {"minecraft:nether_brick", zero},
	"minecraft:bone_block":                 {"minecraft:quartz_block", pillar},
	"minecraft:structure_void":             {"minecraft:air", zero},
	"minecraft:observer":                   {"minecraft:dispenser", keep},
	"minecraft:concrete":                   {"minecraft:stained_hardened_clay", keep},
	"minecraft:concrete_powder":            {"minecraft:stained_hardened_clay", keep},
	"minecraft:structure_block":            {"minecraft:air", zero},
	"doubleslabs:double_slab":              {"minecraft:double_stone_slab", zero},
}

// 1.8 で増えた石 (stone の meta 1〜6)
var stone1710 = map[int]struct {
	n string
	m int
}{
	1: {"minecraft:hardened_clay", 0}, 2: {"minecraft:hardened_clay", 0}, // 花崗岩
	3: {"minecraft:quartz_block", 0}, 4: {"minecraft:quartz_block", 0}, // 閃緑岩
	5: {"minecraft:stone", 0}, 6: {"minecraft:double_stone_slab", 8}, // 安山岩
}

var slabVariant = map[string]int{"stone": 0, "sandstone": 1, "wood_old": 2, "cobblestone": 3, "brick": 4,
	"stone_brick": 5, "nether_brick": 6, "quartz": 7}
var woodVariant = map[string]int{"oak": 0, "spruce": 1, "birch": 2, "jungle": 3, "acacia": 4, "dark_oak": 5}

var teNew2Old = map[string]string{"furnace": "Furnace", "chest": "Chest", "trapped_chest": "Chest",
	"ender_chest": "EnderChest", "jukebox": "RecordPlayer", "dispenser": "Trap", "dropper": "Dropper",
	"sign": "Sign", "mob_spawner": "MobSpawner", "noteblock": "Music", "piston": "Piston",
	"brewing_stand": "Cauldron", "enchanting_table": "EnchantTable", "end_portal": "Airportal",
	"command_block": "Control", "beacon": "Beacon", "skull": "Skull", "daylight_detector": "DLDetector",
	"hopper": "Hopper", "comparator": "Comparator", "flower_pot": "FlowerPot"}
var teOld2New = map[string]string{}
var enNew2Old = map[string]string{"minecraft:painting": "Painting", "minecraft:item_frame": "ItemFrame",
	"minecraft:sheep": "Sheep", "minecraft:chicken": "Chicken"}
var enOld2New = map[string]string{}
var skipEntities = map[string]bool{"Bat": true, "Item": true, "XPOrb": true,
	"minecraft:bat": true, "minecraft:item": true, "minecraft:xp_orb": true}

func init() {
	for k, v := range teNew2Old {
		if k != "trapped_chest" {
			teOld2New[v] = "minecraft:" + k
		}
	}
	for k, v := range enNew2Old {
		enOld2New[v] = k
	}
}

// ---------------- Mod ブロック ID ----------------

// 1.7.10 Forge の level.dat から {ブロック名: 数値ID}
func LoadLevelIDs(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	root, err := ReadNBT(data)
	if err != nil {
		return nil, err
	}
	fml := root.Comp("FML")
	if fml == nil || fml.List("ItemData") == nil {
		return nil, errors.New("level.dat に FML ItemData がありません (1.7.10 Forge のワールドではない?)")
	}
	ids := map[string]int{}
	for _, it := range fml.List("ItemData").Items {
		c, ok := it.V.(*Compound)
		if !ok {
			continue
		}
		k := c.Str("K")
		if len(k) > 1 && k[0] == '\x01' { // \x01 = ブロック, \x02 = アイテム
			ids[k[1:]] = c.Int("V", 0)
		}
	}
	return ids, nil
}

func LoadIDMap(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := map[string]int{}
	err = json.Unmarshal(data, &m)
	return m, err
}

func resolveMod(name string, table map[string]int) (int, bool) {
	if v, ok := table[name]; ok {
		return v, true
	}
	cands := []string{strings.ToLower(name)}
	if strings.Count(name, ":") >= 2 { // "RTM:rtm:insulator" -> "rtm:insulator"
		cands = append(cands, strings.ToLower(name[strings.Index(name, ":")+1:]))
	}
	for k, v := range table {
		lk := strings.ToLower(k)
		for _, c := range cands {
			if lk == c {
				return v, true
			}
		}
	}
	return 0, false
}

// ---------------- 変換 ----------------

type Report struct {
	Unknown map[string]map[string]bool // ブロック名 -> ファイル名
	Subst   map[string]bool
	DropTE  map[string]bool
	DropEn  map[string]bool
}

func NewReport() *Report {
	return &Report{map[string]map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}}
}

func IsNgto(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, 16)
	n, _ := f.Read(head)
	head = head[:n]
	return n >= 3 && head[0] == 0x0a && head[1] == 0 && head[2] == 0 && strings.Contains(string(head), "ByteData")
}

func jsonText(s string) string {
	if s == "" || !strings.ContainsAny(s[:1], "{[\"") {
		return s
	}
	var j interface{}
	if json.Unmarshal([]byte(s), &j) != nil {
		return s
	}
	var flat func(x interface{}) string
	flat = func(x interface{}) string {
		switch v := x.(type) {
		case string:
			return v
		case []interface{}:
			out := ""
			for _, i := range v {
				out += flat(i)
			}
			return out
		case map[string]interface{}:
			out, _ := v["text"].(string)
			if ex, ok := v["extra"].([]interface{}); ok {
				for _, i := range ex {
					out += flat(i)
				}
			}
			return out
		}
		return ""
	}
	return flat(j)
}

func tagInt(t *Tag) int {
	switch v := t.V.(type) {
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}

type bm struct{ id, meta int }

// Convert: 1ファイル変換。サイズを返す
func Convert(src, dst, target string, modids map[string]int, rep *Report) (int, int, int, error) {
	raw, err := os.ReadFile(src)
	if err != nil {
		return 0, 0, 0, err
	}
	outer, err := ReadNBT(raw)
	if err != nil {
		return 0, 0, 0, err
	}
	bd := outer.Get("ByteData")
	if bd == nil {
		return 0, 0, 0, errors.New("ngto ではありません (ByteData なし)")
	}
	inner, err := ReadNBT(bd.V.([]byte))
	if err != nil {
		return 0, 0, 0, err
	}
	X, Y, Z := inner.Int("SizeX", 0), inner.Int("SizeY", 0), inner.Int("SizeZ", 0)
	ox, oy, oz := inner.Int("OrigX", 0), inner.Int("OrigY", 0), inner.Int("OrigZ", 0)
	vol := X * Y * Z
	bdt := inner.Get("BData")
	if bdt == nil || vol <= 0 {
		return 0, 0, 0, errors.New("BData がありません")
	}
	bdata := bdt.V.([]byte)
	if len(bdata) != vol {
		return 0, 0, 0, fmt.Errorf("BDataの長さ(%d)がサイズ(%d)と合いません", len(bdata), vol)
	}
	if X > 32767 || Y > 32767 || Z > 32767 {
		return 0, 0, 0, errors.New("大きすぎます")
	}

	vanilla := ids112
	if target == "1.7.10" {
		vanilla = ids1710
	}
	used := map[string]int{}
	lookup := func(name string, meta int) (bm, bool) {
		if target == "1.7.10" {
			if name == "minecraft:stone" {
				if s, ok := stone1710[meta]; ok {
					name, meta = s.n, s.m
					rep.Subst["minecraft:stone(花崗岩等)"] = true
				}
			} else if s, ok := subst1710[name]; ok {
				if _, inV := vanilla[name]; !inV {
					rep.Subst[name] = true
					name, meta = s.name, s.meta(meta)
				}
			}
		}
		if id, ok := vanilla[name]; ok {
			used[name] = id
			return bm{id, meta & 0xF}, true
		}
		if id, ok := resolveMod(name, modids); ok {
			used[name] = id
			return bm{id, meta & 0xF}, true
		}
		return bm{}, false
	}

	palette := map[int]bm{}
	if l := inner.List("IdList"); l != nil {
		for _, e := range l.Items {
			c, ok := e.V.(*Compound)
			if !ok {
				continue
			}
			set := c.Comp("Set")
			if set == nil {
				continue
			}
			blk := set.Str("Block")
			meta := set.Int("Meta", 0) & 0xFF
			r, ok := lookup(blk, meta)
			if !ok {
				if rep.Unknown[blk] == nil {
					rep.Unknown[blk] = map[string]bool{}
				}
				rep.Unknown[blk][filepath.Base(src)] = true
			}
			palette[c.Int("Id", 0)] = r
		}
	}

	blocks := make([]byte, vol)
	data := make([]byte, vol)
	add := make([]byte, (vol+1)/2)
	useAdd := false
	sidx := func(x, y, z int) int { return (y*Z+z)*X + x } // schematic の並び
	put := func(i, id, meta int) {
		blocks[i] = byte(id)
		data[i] = byte(meta & 0xF)
		hi := byte((id >> 8) & 0xF)
		if hi != 0 {
			useAdd = true
			if i&1 == 1 {
				add[i>>1] = add[i>>1]&0x0F | hi<<4
			} else {
				add[i>>1] = add[i>>1]&0xF0 | hi
			}
		}
	}
	for x := 0; x < X; x++ {
		for y := 0; y < Y; y++ {
			base := (x*Y + y) * Z
			for z := 0; z < Z; z++ {
				b := palette[int(bdata[base+z]+128)] // 値+128 = IdList の Id
				if b.id != 0 {
					put(sidx(x, y, z), b.id, b.meta)
				}
			}
		}
	}

	// ---- タイルエンティティ ----
	var tiles []*Tag
	if nb := inner.Comp("NBTs"); nb != nil {
		keys := append([]string{}, nb.Keys...)
		sort.Slice(keys, func(a, b int) bool { ka, _ := strconv.Atoi(keys[a]); kb, _ := strconv.Atoi(keys[b]); return ka < kb })
		for _, key := range keys {
			te, ok := nb.M[key].V.(*Compound)
			if !ok {
				continue
			}
			k, err := strconv.Atoi(key) // キー = BData の index (座標より確実)
			if err != nil || k < 0 || k >= vol {
				continue
			}
			x, y, z := k/(Y*Z), (k/Z)%Y, k%Z
			tid := te.Str("id")
			d := te.Clone()
			if tid == "doubleslabs:dynamic_slab" { // 重ねハーフ -> 二段ハーフ
				st := te.Comp("negativeBlock")
				if st == nil {
					st = te.Comp("positiveBlock")
				}
				nm, vr := "", "stone"
				if st != nil && st.Comp("state") != nil {
					s := st.Comp("state")
					nm = s.Str("Name")
					if p := s.Comp("Properties"); p != nil && p.Str("variant") != "" {
						vr = p.Str("variant")
					}
				}
				if nm == "minecraft:wooden_slab" {
					put(sidx(x, y, z), vanilla["minecraft:double_wooden_slab"], woodVariant[vr])
				} else {
					put(sidx(x, y, z), vanilla["minecraft:double_stone_slab"], slabVariant[vr])
				}
				if _, ok := modids["doubleslabs:double_slab"]; target == "1.7.10" || !ok {
					continue
				}
			}
			if strings.HasPrefix(tid, "minecraft:banner") && target == "1.7.10" {
				continue
			}
			if target == "1.7.10" && strings.HasPrefix(tid, "minecraft:") {
				n, ok := teNew2Old[tid[10:]]
				if !ok {
					rep.DropTE[tid] = true
					continue
				}
				d.Set("id", TStr(n))
			} else if target != "1.7.10" {
				if n, ok := teOld2New[tid]; ok {
					d.Set("id", TStr(n))
				}
			}
			if target == "1.7.10" {
				for _, t := range []string{"Text1", "Text2", "Text3", "Text4"} {
					if s := d.Get(t); s != nil && s.T == TagString {
						d.Set(t, TStr(jsonText(s.V.(string))))
					}
				}
				if it := d.Get("Item"); it != nil && it.T == TagString { // 植木鉢の中身
					d.Set("Item", TInt(vanilla[it.V.(string)]))
				}
				d.Del("Lock")
			}
			d.Set("x", TInt(x))
			d.Set("y", TInt(y))
			d.Set("z", TInt(z))
			tiles = append(tiles, TComp(d))
		}
	}

	// ---- エンティティ (座標を建物の角からの相対に) ----
	var ents []*Tag
	if el := inner.List("Entities"); el != nil {
		for _, it := range el.Items {
			e, ok := it.V.(*Compound)
			if !ok {
				continue
			}
			eid := e.Str("id")
			if skipEntities[eid] {
				continue
			}
			d := e.Clone()
			if n, ok := enNew2Old[eid]; ok && target == "1.7.10" {
				d.Set("id", TStr(n))
				if f := d.Get("Facing"); f != nil { // 額縁・絵の向き
					fv := tagInt(f)
					d.Del("Facing")
					d.Set("Direction", TByte(fv))
					dir := map[int]int{0: 2, 1: 1, 2: 0, 3: 3}[fv]
					d.Set("Dir", TByte(dir))
				}
			} else if strings.HasPrefix(eid, "minecraft:") && target == "1.7.10" {
				rep.DropEn[eid] = true
				continue
			} else if n, ok := enOld2New[eid]; ok && target != "1.7.10" {
				d.Set("id", TStr(n))
				if d.Get("Facing") == nil && d.Get("Direction") != nil {
					d.Set("Facing", TByte(tagInt(d.Get("Direction"))))
				}
			}
			if p := d.List("Pos"); p != nil && len(p.Items) == 3 {
				o := []float64{float64(ox), float64(oy), float64(oz)}
				var np []*Tag
				for i, t := range p.Items {
					v, _ := t.V.(float64)
					np = append(np, TDouble(v-o[i]))
				}
				d.Set("Pos", TList(TagDouble, np))
			}
			for k, o := range map[string]int{"TileX": ox, "TileY": oy, "TileZ": oz} {
				if t := d.Get(k); t != nil {
					d.Set(k, TInt(tagInt(t)-o))
				}
			}
			for _, k := range []string{"UUIDMost", "UUIDLeast", "Dimension", "ForgeCaps"} {
				d.Del(k)
			}
			ents = append(ents, TComp(d))
		}
	}

	out := NewCompound()
	out.Set("Width", TShort(X))
	out.Set("Height", TShort(Y))
	out.Set("Length", TShort(Z))
	out.Set("Materials", TStr("Alpha"))
	out.Set("Blocks", TBytes(blocks))
	out.Set("Data", TBytes(data))
	out.Set("Entities", TList(TagCompound, ents))
	out.Set("TileEntities", TList(TagCompound, tiles))
	// 原点0・オフセット0 = 貼り付けると自分の足元が建物の最小角(西・下・北)
	for _, k := range []string{"WEOriginX", "WEOriginY", "WEOriginZ", "WEOffsetX", "WEOffsetY", "WEOffsetZ"} {
		out.Set(k, TInt(0))
	}
	mp := NewCompound()
	names := make([]string, 0, len(used))
	for n := range used {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		mp.Set(n, TShort(used[n]))
	}
	out.Set("SchematicaMapping", TComp(mp))
	if useAdd {
		out.Set("AddBlocks", TBytes(add))
	}
	bytesOut, err := WriteNBT("Schematic", out)
	if err != nil {
		return 0, 0, 0, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, 0, 0, err
	}
	return X, Y, Z, os.WriteFile(dst, bytesOut, 0o644)
}

// ---------------- 入力の展開 ----------------

type Job struct{ Src, Dst string }

func outName(name string) string {
	if strings.HasSuffix(strings.ToLower(name), ".ngto") {
		name = name[:len(name)-5]
	}
	return name + ".schematic"
}

// 出力先を省略したときに作るサブフォルダ名
// (同梱の .schematic を上書きしないよう、入力の隣の「schematic」フォルダに出す)
const DefaultOutSub = "schematic"

// 入力 (ファイル/フォルダ) から変換ジョブ一覧を作る。outDir が空なら入力の隣の schematic フォルダ
func BuildJobs(inputs []string, outDir string) []Job {
	var jobs []Job
	for _, in := range inputs {
		st, err := os.Stat(in)
		if err != nil {
			continue
		}
		if st.IsDir() {
			base := outDir
			if base == "" {
				base = filepath.Join(in, DefaultOutSub)
			}
			absBase, _ := filepath.Abs(base)
			filepath.Walk(in, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					if ap, _ := filepath.Abs(p); ap == absBase { // 出力先フォルダの中は見ない
						return filepath.SkipDir
					}
					return nil
				}
				if !IsNgto(p) {
					return nil
				}
				rel, _ := filepath.Rel(in, p)
				jobs = append(jobs, Job{p, filepath.Join(base, filepath.Dir(rel), outName(filepath.Base(rel)))})
				return nil
			})
		} else {
			base := outDir
			if base == "" {
				base = filepath.Join(filepath.Dir(in), DefaultOutSub)
			}
			jobs = append(jobs, Job{in, filepath.Join(base, outName(filepath.Base(in)))})
		}
	}
	return jobs
}

// 全部変換してログを返す
func RunAll(inputs []string, outDir, target, levelPath, idmapPath string, logf func(string), prog func(done, total int)) Summary {
	var sum Summary
	modids := map[string]int{}
	if levelPath != "" {
		m, err := LoadLevelIDs(levelPath)
		if err != nil {
			logf("level.dat 読み込み失敗: " + err.Error())
		} else {
			for k, v := range m {
				modids[k] = v
			}
			logf(fmt.Sprintf("level.dat から %d 個のブロックIDを読み込みました", len(m)))
		}
	}
	if idmapPath != "" {
		m, err := LoadIDMap(idmapPath)
		if err != nil {
			logf("ID表JSON 読み込み失敗: " + err.Error())
		} else {
			for k, v := range m {
				modids[k] = v
			}
		}
	}
	jobs := BuildJobs(inputs, outDir)
	if len(jobs) == 0 {
		logf("ngto ファイルが見つかりませんでした")
		return sum
	}
	sum.Total = len(jobs)
	rep := NewReport()
	ok, ng := 0, 0
	for n, j := range jobs {
		if prog != nil {
			prog(n, len(jobs))
		}
		x, y, z, err := Convert(j.Src, j.Dst, target, modids, rep)
		if err != nil {
			ng++
			logf("NG  " + filepath.Base(j.Src) + " : " + err.Error())
		} else {
			ok++
			logf(fmt.Sprintf("OK  %s  (%dx%dx%d)", filepath.Base(j.Dst), x, y, z))
		}
	}
	if prog != nil {
		prog(len(jobs), len(jobs))
	}
	sum.OK, sum.NG, sum.Unknown = ok, ng, len(rep.Unknown)
	logf("")
	logf(fmt.Sprintf("%d 件変換 / %d 件失敗  (対象 %s)", ok, ng, target))
	if len(rep.Subst) > 0 {
		logf("")
		logf("[置き換え] " + target + " に無いブロックを近いものにしました: " + joinKeys(rep.Subst))
	}
	if len(rep.DropTE) > 0 {
		logf("[削除] 未対応タイルエンティティ: " + joinKeys(rep.DropTE))
	}
	if len(rep.DropEn) > 0 {
		logf("[削除] 未対応エンティティ: " + joinKeys(rep.DropEn))
	}
	if len(rep.Unknown) > 0 {
		logf("")
		logf("[空気になったブロック] 数値IDが不明です。level.dat を指定してください:")
		names := make([]string, 0, len(rep.Unknown))
		for n := range rep.Unknown {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fs := make([]string, 0)
			for f := range rep.Unknown[n] {
				fs = append(fs, f)
			}
			sort.Strings(fs)
			logf(fmt.Sprintf("  %s  … %d ファイル (例: %s)", n, len(fs), fs[0]))
		}
	}
	return sum
}

type Summary struct{ Total, OK, NG, Unknown int }

func joinKeys(m map[string]bool) string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return strings.Join(ks, ", ")
}
