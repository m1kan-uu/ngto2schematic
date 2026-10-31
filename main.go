//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

//go:embed res/icon.png
var iconPNG []byte

// 配色 (アイコンの設計図ブルーに合わせる)
var (
	colHeader  = walk.RGB(13, 71, 161)   // 濃い青
	colHeadSub = walk.RGB(187, 222, 251) // ヘッダーの説明文
	colBody    = walk.RGB(244, 246, 248) // 本体の背景
	colTitle   = walk.RGB(13, 51, 110)   // 見出し
	colAccent  = walk.RGB(46, 125, 50)   // 成功(緑)
	colWarn    = walk.RGB(230, 81, 0)    // 注意(オレンジ)
	colMuted   = walk.RGB(96, 112, 120)  // 補足文字
)

const uiFont = "Yu Gothic UI"

type inputModel struct {
	walk.ListModelBase
	items []string
}

func (m *inputModel) ItemCount() int { return len(m.items) }
func (m *inputModel) Value(i int) interface{} {
	p := m.items[i]
	if st, err := os.Stat(p); err == nil && st.IsDir() {
		return "[フォルダ]  " + p
	}
	return p
}
func (m *inputModel) add(paths ...string) {
	for _, p := range paths {
		dup := false
		for _, q := range m.items {
			if strings.EqualFold(p, q) {
				dup = true
				break
			}
		}
		if !dup {
			m.items = append(m.items, p)
		}
	}
	m.PublishItemsReset()
}

func loadIcon() (image.Image, error) { return png.Decode(bytes.NewReader(iconPNG)) }

func main() {
	var (
		mw         *walk.MainWindow
		list       *walk.ListBox
		countLbl   *walk.Label
		hintLbl    *walk.Label
		outEdit    *walk.LineEdit
		levelEdit  *walk.LineEdit
		levelLbl   *walk.Label
		target     *walk.ComboBox
		targetLbl  *walk.Label
		logEdit    *walk.TextEdit
		runBtn     *walk.PushButton
		progress   *walk.ProgressBar
		status     *walk.Label
		headerImg  *walk.ImageView
		rbForward  *walk.RadioButton
		rbBackward *walk.RadioButton
	)
	reverse := false // false: ngto→schematic / true: schematic→ngto
	model := &inputModel{}

	// exe にドロップされたファイルが schematic だけなら逆方向で起動
	if len(os.Args) > 1 {
		allSchem := true
		for _, a := range os.Args[1:] {
			if !strings.HasSuffix(strings.ToLower(a), ".schematic") {
				allSchem = false
			}
		}
		reverse = allSchem
		model.add(os.Args[1:]...)
	}

	updateCount := func() {
		if countLbl == nil {
			return
		}
		if n := len(model.items); n == 0 {
			countLbl.SetText("まだ何も追加されていません")
		} else {
			countLbl.SetText(fmt.Sprintf("%d 件追加済み", n))
		}
	}

	applyMode := func() {
		if reverse {
			mw.SetTitle("schematic → ngto 変換")
			hintLbl.SetText(".schematic ファイルやフォルダを、このウィンドウにドラッグ&&ドロップしてください (フォルダは中身をまとめて変換)")
			outEdit.SetCueBanner("空欄なら元のファイルの隣に「ngto」フォルダを作って保存します")
			targetLbl.SetEnabled(false)
			target.SetEnabled(false)
			levelLbl.SetText("元ワールドの level.dat")
			levelEdit.SetCueBanner("schematic を保存したワールド。Modブロックの名前を読み取ります (任意)")
		} else {
			mw.SetTitle("ngto → schematic 変換")
			hintLbl.SetText("ngto ファイルやフォルダを、このウィンドウにドラッグ&&ドロップしてください (フォルダは中身をまとめて変換)")
			outEdit.SetCueBanner("空欄なら元のファイルの隣に「schematic」フォルダを作って保存します")
			targetLbl.SetEnabled(true)
			target.SetEnabled(true)
			levelLbl.SetText("貼り付け先の level.dat")
			levelEdit.SetCueBanner("RTMの足場などModブロックのIDをワールドから読み取ります (任意・1.7.10)")
		}
		outEdit.Invalidate()
		levelEdit.Invalidate()
		rbForward.SetChecked(!reverse)
		rbBackward.SetChecked(reverse)
	}

	logf := func(s string) {
		mw.Synchronize(func() { logEdit.AppendText(s + "\r\n") })
	}
	setStatus := func(s string, c walk.Color) {
		mw.Synchronize(func() {
			status.SetTextColor(c)
			status.SetText(s)
		})
	}

	pickFolder := func(title string) string {
		dlg := new(walk.FileDialog)
		dlg.Title = title
		if ok, _ := dlg.ShowBrowseFolder(mw); ok {
			return dlg.FilePath
		}
		return ""
	}

	run := func() {
		if len(model.items) == 0 {
			walk.MsgBox(mw, "変換", "変換するファイルかフォルダを追加してください。\nウィンドウにドラッグ&ドロップでも追加できます。", walk.MsgBoxIconInformation)
			return
		}
		inputs := append([]string{}, model.items...)
		out := strings.Trim(strings.TrimSpace(outEdit.Text()), `"`)
		lvl := strings.Trim(strings.TrimSpace(levelEdit.Text()), `"`)
		tg := "1.7.10"
		if target.CurrentIndex() == 1 {
			tg = "1.12"
		}
		if lvl != "" { // ワールドフォルダを指定された場合
			if st, err := os.Stat(lvl); err == nil && st.IsDir() {
				lvl = filepath.Join(lvl, "level.dat")
			}
		}
		rev := reverse
		runBtn.SetEnabled(false)
		rbForward.SetEnabled(false)
		rbBackward.SetEnabled(false)
		runBtn.SetText("変換中…")
		logEdit.SetText("")
		progress.SetValue(0)
		setStatus("変換しています…", colMuted)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logf(fmt.Sprintf("エラー: 予期しない問題が起きました (%v)", r))
					setStatus("エラーで中断しました", colWarn)
				}
				mw.Synchronize(func() {
					runBtn.SetEnabled(true)
					rbForward.SetEnabled(true)
					rbBackward.SetEnabled(true)
					runBtn.SetText("変換する")
				})
			}()
			prog := func(done, total int) {
				mw.Synchronize(func() {
					progress.SetRange(0, total)
					progress.SetValue(done)
				})
			}
			var sum Summary
			if rev {
				sum = RunAllRev(inputs, out, lvl, logf, prog)
			} else {
				sum = RunAll(inputs, out, tg, lvl, "", logf, prog)
			}
			switch {
			case sum.Total == 0:
				setStatus("変換できるファイルが見つかりませんでした", colWarn)
			case sum.NG > 0:
				setStatus(fmt.Sprintf("完了: %d 件成功 / %d 件失敗 (ログを確認)", sum.OK, sum.NG), colWarn)
			case sum.Unknown > 0:
				setStatus(fmt.Sprintf("完了: %d 件 (一部のModブロックが空気になりました。ログを確認)", sum.OK), colWarn)
			default:
				setStatus(fmt.Sprintf("完了: %d 件すべて変換しました", sum.OK), colAccent)
			}
		}()
	}

	sectionTitle := func(text string) Widget {
		return Label{Text: text, Font: Font{Family: uiFont, PointSize: 11, Bold: true}, TextColor: colTitle}
	}

	err := MainWindow{
		AssignTo:   &mw,
		Title:      "ngto ⇄ schematic 変換",
		MinSize:    Size{Width: 660, Height: 700},
		Size:       Size{Width: 780, Height: 800},
		Font:       Font{Family: uiFont, PointSize: 10},
		Background: SolidColorBrush{Color: colBody},
		Layout:     VBox{MarginsZero: true, SpacingZero: true},
		OnDropFiles: func(files []string) {
			model.add(files...)
			updateCount()
		},
		Children: []Widget{
			// ---------- ヘッダー ----------
			Composite{
				Background: SolidColorBrush{Color: colHeader},
				Layout:     HBox{Margins: Margins{Left: 20, Top: 14, Right: 20, Bottom: 14}, Spacing: 14},
				Children: []Widget{
					ImageView{AssignTo: &headerImg, MinSize: Size{Width: 52, Height: 52}, MaxSize: Size{Width: 52, Height: 52}, Mode: ImageViewModeZoom},
					Composite{
						Layout: VBox{MarginsZero: true, Spacing: 2},
						Children: []Widget{
							Label{Text: "ngto ⇄ schematic 変換", Font: Font{Family: uiFont, PointSize: 16, Bold: true}, TextColor: walk.RGB(255, 255, 255)},
							Label{Text: "NGTLib / RTM の建物データ(.ngto)と WorldEdit・MCEdit の .schematic を相互に変換します", TextColor: colHeadSub},
						},
					},
					HSpacer{},
				},
			},
			// ---------- 本体 ----------
			Composite{
				Layout: VBox{Margins: Margins{Left: 20, Top: 14, Right: 20, Bottom: 16}, Spacing: 10},
				Children: []Widget{
					sectionTitle("変換の向き"),
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 24},
						Children: []Widget{
							RadioButton{AssignTo: &rbForward, Text: "ngto  →  schematic", Font: Font{Family: uiFont, PointSize: 11},
								OnClicked: func() { reverse = false; applyMode() }},
							RadioButton{AssignTo: &rbBackward, Text: "schematic  →  ngto", Font: Font{Family: uiFont, PointSize: 11},
								OnClicked: func() { reverse = true; applyMode() }},
							HSpacer{},
						},
					},

					VSpacer{Size: 2},
					sectionTitle("① 変換するデータ"),
					Label{AssignTo: &hintLbl, TextColor: colMuted},
					ListBox{AssignTo: &list, Model: model, MinSize: Size{Height: 120}},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 6},
						Children: []Widget{
							PushButton{Text: "＋ ファイル", MinSize: Size{Width: 96}, OnClicked: func() {
								dlg := new(walk.FileDialog)
								if reverse {
									dlg.Title = "schematic ファイルを選択"
									dlg.Filter = "schematic ファイル (*.schematic)|*.schematic|すべてのファイル (*.*)|*.*"
								} else {
									dlg.Title = "ngto ファイルを選択"
									dlg.Filter = "ngto ファイル (*.ngto)|*.ngto|すべてのファイル (*.*)|*.*"
								}
								if ok, _ := dlg.ShowOpenMultiple(mw); ok {
									model.add(dlg.FilePaths...)
									updateCount()
								}
							}},
							PushButton{Text: "＋ フォルダ", MinSize: Size{Width: 96}, OnClicked: func() {
								if p := pickFolder("変換するファイルが入ったフォルダを選択"); p != "" {
									model.add(p)
									updateCount()
								}
							}},
							PushButton{Text: "選択を外す", OnClicked: func() {
								i := list.CurrentIndex()
								if i >= 0 && i < len(model.items) {
									model.items = append(model.items[:i], model.items[i+1:]...)
									model.PublishItemsReset()
									updateCount()
								}
							}},
							PushButton{Text: "すべてクリア", OnClicked: func() {
								model.items = nil
								model.PublishItemsReset()
								updateCount()
							}},
							HSpacer{},
							Label{AssignTo: &countLbl, TextColor: colMuted},
						},
					},

					VSpacer{Size: 2},
					sectionTitle("② 設定"),
					Composite{
						Layout: Grid{Columns: 3, MarginsZero: true, Spacing: 8},
						Children: []Widget{
							Label{Text: "出力先フォルダ"},
							LineEdit{AssignTo: &outEdit},
							PushButton{Text: "参照…", OnClicked: func() {
								if p := pickFolder("出力先フォルダを選択"); p != "" {
									outEdit.SetText(p)
								}
							}},

							Label{AssignTo: &targetLbl, Text: "貼り付け先バージョン"},
							ComboBox{AssignTo: &target, Model: []string{"Minecraft 1.7.10", "Minecraft 1.10.2 / 1.12.2"}, CurrentIndex: 0},
							HSpacer{},

							Label{AssignTo: &levelLbl, Text: "level.dat"},
							LineEdit{AssignTo: &levelEdit},
							PushButton{Text: "参照…", OnClicked: func() {
								dlg := new(walk.FileDialog)
								dlg.Title = "ワールドの level.dat を選択"
								dlg.Filter = "level.dat|level.dat|すべてのファイル (*.*)|*.*"
								if ok, _ := dlg.ShowOpen(mw); ok {
									levelEdit.SetText(dlg.FilePath)
								}
							}},
						},
					},

					VSpacer{Size: 4},
					PushButton{AssignTo: &runBtn, Text: "変換する", MinSize: Size{Height: 44},
						Font: Font{Family: uiFont, PointSize: 12, Bold: true}, OnClicked: run},
					ProgressBar{AssignTo: &progress, MinSize: Size{Height: 10}, MaxSize: Size{Height: 10}},
					Label{AssignTo: &status, Text: "準備OK", TextColor: colMuted},

					sectionTitle("ログ"),
					TextEdit{AssignTo: &logEdit, ReadOnly: true, VScroll: true, MinSize: Size{Height: 130},
						Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
						Font:       Font{Family: "BIZ UDGothic", PointSize: 9}},
				},
			},
		},
	}.Create()
	if err != nil {
		walk.MsgBox(nil, "エラー", err.Error(), walk.MsgBoxIconError)
		return
	}

	if img, err := loadIcon(); err == nil {
		if ic, err := walk.NewIconFromImageForDPI(img, 96); err == nil {
			mw.SetIcon(ic)
		}
		if bmp, err := walk.NewBitmapFromImageForDPI(img, 96); err == nil {
			headerImg.SetImage(bmp)
		}
	}
	applyMode()
	updateCount()
	mw.Run()
}
