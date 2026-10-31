package main

// schematic -> ngto 変換 (MCEdit / WorldEdit / Schematica 形式の .schematic を読む)

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 数値ID -> ブロック名 (バニラ。1.12 までの全ID)
var vanillaByID = func() map[int]string {
	m := map[int]string{}
	for n, i := range ids112 {
		m[i] = n
	}
	return m
}

func IsSchematic(path string) bool {
	if !strings.HasSuffix(strings.ToLower(path), ".schematic") {
		return false
	}
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// ConvertToNgto: 1ファイル変換。srcIDs は元ワールドの level.dat 由来の {名前: ID}
func ConvertToNgto(src, dst string, srcIDs map[string]int, rep *Report) (int, int, int, error) {
	raw, err := os.ReadFile(src)
	if err != nil {
		return 0, 0, 0, err
	}
	s, err := ReadNBT(raw)
	if err != nil {
		return 0, 0, 0, err
	}
	if s.Get("Palette") != nil || s.Get("BlockData") != nil {
		return 0, 0, 0, errors.New("1.13 以降の新形式 schematic (.schem) は未対応です")
	}
	if s.Get("Blocks") == nil {
		return 0, 0, 0, errors.New("Blocks がありません (schematic ではない?)")
	}
	X, Y, Z := s.Int("Width", 0), s.Int("Height", 0), s.Int("Length", 0)
	vol := X * Y * Z
	blocks, _ := s.Get("Blocks").V.([]byte)
	var data []byte
	if t := s.Get("Data"); t != nil {
		data, _ = t.V.([]byte)
	}
	if vol <= 0 || len(blocks) != vol || (data != nil && len(data) != vol) {
		return 0, 0, 0, fmt.Errorf("サイズ(%dx%dx%d)とブロック配列の長さが合いません", X, Y, Z)
	}
	var addNib, addFull []byte
	if t := s.Get("AddBlocks"); t != nil {
		addNib, _ = t.V.([]byte)
	}
	if t := s.Get("Add"); t != nil {
		addFull, _ = t.V.([]byte)
	}
	blockID := func(i int) int {
		id := int(blocks[i])
		if addNib != nil && i>>1 < len(addNib) {
			if i&1 == 1 {
				id |= int(addNib[i>>1]>>4&0xF) << 8
			} else {
				id |= int(addNib[i>>1]&0xF) << 8
			}
		} else if addFull != nil && i < len(addFull) {
			id |= int(addFull[i]) << 8
		}
		return id
	}

	// ---- 数値ID -> 名前 の表 (優先度: schematic 内の対応表 > level.dat > バニラ) ----
	names := vanillaByID()
	for n, i := range srcIDs {
		names[i] = n
	}
	if mp := s.Comp("SchematicaMapping"); mp != nil { // 名前 -> ID
		for _, n := range mp.Keys {
			names[tagInt(mp.M[n])] = n
		}
	}
	if mp := s.Comp("BlockIDs"); mp != nil { // MCEdit-Unified: "ID" -> 名前
		for _, k := range mp.Keys {
			if i, err := strconv.Atoi(k); err == nil {
				if n, ok := mp.M[k].V.(string); ok {
					names[i] = n
				}
			}
		}
	}

	// ---- パレット作成 (ngto は 1バイトなので最大256種) ----
	type key struct {
		name string
		meta int
	}
	palIndex := map[key]int{{"minecraft:air", 0}: 0}
	palOrder := []key{{"minecraft:air", 0}}
	bdata := make([]byte, vol)
	tooMany := false
	unknownHere := map[int]bool{}
	for x := 0; x < X && !tooMany; x++ {
		for y := 0; y < Y && !tooMany; y++ {
			for z := 0; z < Z; z++ {
				i := (y*Z+z)*X + x // schematic の並び
				id := blockID(i)
				meta := 0
				if data != nil {
					meta = int(data[i] & 0xF)
				}
				k := key{"minecraft:air", 0}
				if id != 0 {
					n, ok := names[id]
					if !ok {
						unknownHere[id] = true
					} else {
						k = key{n, meta}
					}
				}
				p, ok := palIndex[k]
				if !ok {
					if len(palOrder) >= 256 {
						tooMany = true
						break
					}
					p = len(palOrder)
					palIndex[k] = p
					palOrder = append(palOrder, k)
				}
				bdata[(x*Y+y)*Z+z] = byte(p - 128) // ngto は Id-128 で保存
			}
		}
	}
	if tooMany {
		return 0, 0, 0, errors.New("ブロックの種類(メタ値違い含む)が256を超えるため ngto にできません")
	}
	for id := range unknownHere {
		n := fmt.Sprintf("ID %d", id)
		if rep.Unknown[n] == nil {
			rep.Unknown[n] = map[string]bool{}
		}
		rep.Unknown[n][filepath.Base(src)] = true
	}

	// ---- 原点 (ngto の OrigX/Y/Z。WorldEdit の保存時座標があれば使う) ----
	ox, oy, oz := s.Int("WEOriginX", 0), s.Int("WEOriginY", 0), s.Int("WEOriginZ", 0)

	idList := make([]*Tag, 0, len(palOrder))
	for i, k := range palOrder {
		set := NewCompound()
		set.Set("Meta", TByte(k.meta))
		set.Set("Block", TStr(k.name))
		e := NewCompound()
		e.Set("Set", TComp(set))
		e.Set("Id", TInt(i))
		idList = append(idList, TComp(e))
	}

	// ---- タイルエンティティ ----
	nbts := NewCompound()
	if tl := s.List("TileEntities"); tl != nil {
		type kv struct {
			k int
			c *Compound
		}
		var arr []kv
		for _, it := range tl.Items {
			te, ok := it.V.(*Compound)
			if !ok {
				continue
			}
			x, y, z := te.Int("x", -1), te.Int("y", -1), te.Int("z", -1)
			if x < 0 || y < 0 || z < 0 || x >= X || y >= Y || z >= Z {
				continue
			}
			d := te.Clone()
			d.Set("x", TInt(ox+x))
			d.Set("y", TInt(oy+y))
			d.Set("z", TInt(oz+z))
			arr = append(arr, kv{(x*Y+y)*Z + z, d})
		}
		sort.Slice(arr, func(a, b int) bool { return arr[a].k < arr[b].k })
		for _, e := range arr {
			nbts.Set(strconv.Itoa(e.k), TComp(e.c))
		}
	}

	// ---- エンティティ (ngto はワールド座標で持つ) ----
	var ents []*Tag
	if el := s.List("Entities"); el != nil {
		for _, it := range el.Items {
			e, ok := it.V.(*Compound)
			if !ok || skipEntities[e.Str("id")] {
				continue
			}
			d := e.Clone()
			// MCEdit は相対座標、WorldEdit は保存時のワールド座標で入っている
			rel := true
			if p := d.List("Pos"); p != nil && len(p.Items) == 3 {
				v := make([]float64, 3)
				for i, t := range p.Items {
					v[i], _ = t.V.(float64)
				}
				size := []float64{float64(X), float64(Y), float64(Z)}
				for i := range v {
					if v[i] < -1 || v[i] > size[i]+1 {
						rel = false
					}
				}
				if rel {
					d.Set("Pos", TList(TagDouble, []*Tag{TDouble(v[0] + float64(ox)), TDouble(v[1] + float64(oy)), TDouble(v[2] + float64(oz))}))
				}
			}
			if rel {
				for k, o := range map[string]int{"TileX": ox, "TileY": oy, "TileZ": oz} {
					if t := d.Get(k); t != nil {
						d.Set(k, TInt(tagInt(t)+o))
					}
				}
			}
			ents = append(ents, TComp(d))
		}
	}

	inner := NewCompound()
	inner.Set("SizeX", TInt(X))
	inner.Set("ObjId", &Tag{TagLong, time.Now().UnixMilli()})
	inner.Set("OrigY", TInt(oy))
	inner.Set("NBTs", TComp(nbts))
	inner.Set("OrigZ", TInt(oz))
	inner.Set("OrigX", TInt(ox))
	inner.Set("SizeZ", TInt(Z))
	inner.Set("SizeY", TInt(Y))
	inner.Set("BData", TBytes(bdata))
	inner.Set("IdList", TList(TagCompound, idList))
	inner.Set("Entities", TList(TagCompound, ents))
	innerGz, err := WriteNBT("", inner)
	if err != nil {
		return 0, 0, 0, err
	}
	outer := NewCompound()
	outer.Set("ByteData", TBytes(innerGz))
	outRaw, err := WriteNBTRaw("", outer) // 外側は非圧縮
	if err != nil {
		return 0, 0, 0, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, 0, 0, err
	}
	return X, Y, Z, os.WriteFile(dst, outRaw, 0o644)
}

// 元ワールドの level.dat から {名前: ID}
func loadSrcIDs(levelPath string, logf func(string)) map[string]int {
	ids := map[string]int{}
	if levelPath == "" {
		return ids
	}
	m, err := LoadLevelIDs(levelPath)
	if err != nil {
		logf("level.dat 読み込み失敗: " + err.Error())
		return ids
	}
	logf(fmt.Sprintf("level.dat から %d 個のブロックIDを読み込みました", len(m)))
	return m
}

func buildRevJobs(inputs []string, outDir string) []Job {
	var jobs []Job
	name := func(n string) string { return strings.TrimSuffix(n, filepath.Ext(n)) + ".ngto" }
	for _, in := range inputs {
		st, err := os.Stat(in)
		if err != nil {
			continue
		}
		if st.IsDir() {
			base := outDir
			if base == "" {
				base = filepath.Join(in, "ngto")
			}
			absBase, _ := filepath.Abs(base)
			filepath.Walk(in, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					if ap, _ := filepath.Abs(p); ap == absBase {
						return filepath.SkipDir
					}
					return nil
				}
				if !IsSchematic(p) {
					return nil
				}
				rel, _ := filepath.Rel(in, p)
				jobs = append(jobs, Job{p, filepath.Join(base, filepath.Dir(rel), name(filepath.Base(rel)))})
				return nil
			})
		} else {
			base := outDir
			if base == "" {
				base = filepath.Join(filepath.Dir(in), "ngto")
			}
			jobs = append(jobs, Job{in, filepath.Join(base, name(filepath.Base(in)))})
		}
	}
	return jobs
}

// schematic -> ngto をまとめて実行
func RunAllRev(inputs []string, outDir, levelPath string, logf func(string), prog func(done, total int)) Summary {
	var sum Summary
	srcIDs := loadSrcIDs(levelPath, logf)
	jobs := buildRevJobs(inputs, outDir)
	if len(jobs) == 0 {
		logf(".schematic ファイルが見つかりませんでした")
		return sum
	}
	sum.Total = len(jobs)
	rep := NewReport()
	for n, j := range jobs {
		if prog != nil {
			prog(n, len(jobs))
		}
		x, y, z, err := ConvertToNgto(j.Src, j.Dst, srcIDs, rep)
		if err != nil {
			sum.NG++
			logf("NG  " + filepath.Base(j.Src) + " : " + err.Error())
		} else {
			sum.OK++
			logf(fmt.Sprintf("OK  %s  (%dx%dx%d)", filepath.Base(j.Dst), x, y, z))
		}
	}
	if prog != nil {
		prog(len(jobs), len(jobs))
	}
	sum.Unknown = len(rep.Unknown)
	logf("")
	logf(fmt.Sprintf("%d 件変換 / %d 件失敗  (schematic → ngto)", sum.OK, sum.NG))
	if len(rep.Unknown) > 0 {
		logf("")
		logf("[空気になったブロック] 名前が分からない数値IDがありました。元ワールドの level.dat を指定してください:")
		ks := make([]string, 0, len(rep.Unknown))
		for k := range rep.Unknown {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		for _, k := range ks {
			fs := make([]string, 0)
			for f := range rep.Unknown[k] {
				fs = append(fs, f)
			}
			sort.Strings(fs)
			logf(fmt.Sprintf("  %s  … %d ファイル (例: %s)", k, len(fs), fs[0]))
		}
	}
	return sum
}

// 非圧縮で書き出す (ngto の外側用)
func WriteNBTRaw(rootName string, c *Compound) ([]byte, error) {
	w := &writer{}
	w.u8(TagCompound)
	w.str(rootName)
	w.payload(&Tag{TagCompound, c})
	return w.b.Bytes(), nil
}
