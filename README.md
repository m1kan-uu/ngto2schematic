# ngto ⇄ schematic 変換ツール

NGTLib / RealTrainMod (RTM) の建物データ (`.ngto`) と、WorldEdit・MCEdit・Schematica で使う `.schematic` を相互に変換するツールです。

## 動作環境

- Windows 10 / 11 (64bit)
- インストール不要。`ngto2schematic.exe` をダブルクリックで起動します。

## 使い方

1. 「変換の向き」で **ngto → schematic** か **schematic → ngto** を選ぶ
2. 変換したいファイルやフォルダをウィンドウにドラッグ＆ドロップ
   (「＋ファイル」「＋フォルダ」ボタンでも追加できます)
3. 必要なら設定を変える (下の「[設定](#設定)」参照)
4. 「変換する」を押す

- フォルダを入れると、中のファイルをまとめて変換します。
- 出力先を空欄にすると、元のファイルの隣に `schematic` または `ngto` フォルダを自動で作って保存します。
- exe のアイコンにファイルを直接ドロップしても起動できます。

## 設定

### 出力先フォルダ

変換したファイルの保存先。空欄なら上記の自動フォルダです。

### 貼り付け先バージョン (ngto → schematic のみ)

Minecraft 1.7.10 か 1.10.2 / 1.12.2 を選びます。
1.7.10 を選ぶと、1.7.10 に無いブロックを近いブロックに置き換えます。

| 元のブロック | 置き換え先 |
|---|---|
| コンクリート | 色付き粘土 |
| シーランタン | グロウストーン |
| 閃緑岩 | クォーツブロック |
| 木の種類違いのフェンス | オークのフェンス |

旗 (バナー) は 1.7.10 に無いため消えます。

### level.dat (任意)

RTM の足場など、Mod のブロックを正しく変換するために使います。
Mod のブロックはワールドごとに番号 (ID) が違うため、ワールドフォルダの中の `level.dat` を指定してください。

| 変換の向き | 指定する level.dat |
|---|---|
| ngto → schematic | 貼り付け先のワールドの level.dat |
| schematic → ngto | schematic を保存したワールドの level.dat |

指定しない場合、Mod のブロックは空気になります (ログに一覧が出ます)。

> ※ level.dat から ID を読めるのは 1.7.10 (Forge) のワールドだけです。

## 変換したファイルの使い方

- 各自調べてください。

## できないこと・注意

- 1.13 以降の WorldEdit が保存する新形式 (`.schem`) には対応していません。
- ngto は 1マス1バイトの形式なので、ブロックの種類 (メタ値違いも別カウント) が 256 種類を超える建物は schematic → ngto できません。
- チェストの中身などのアイテムは、バージョンを越えると正しく引き継がれないことがあります。
- 初回起動時に「Windows によって PC が保護されました」と出た場合は、「詳細情報」→「実行」で起動できます。
  (exe を右クリック →「プロパティ」→「許可する」にチェックしても消えます)
- 変換した建物データの著作権は、元のデータの作者に帰属します。各データの利用規約に従ってください。

## ライセンス

MITライセンスです。煮るなり焼くなりお好きにどうぞ。

## 使わせてもらったもの

[Go](https://go.dev/) / [walk](https://github.com/lxn/walk) / [win](https://github.com/lxn/win) / [x/sys](https://github.com/golang/sys) / [govaluate](https://github.com/Knetic/govaluate) — ありがとうございます。

<details>
<summary>ライセンス文 (長いので折りたたみ)</summary>

### Go (標準ライブラリ)

```
Copyright 2009 The Go Authors.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google LLC nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

### golang.org/x/sys

```
Copyright 2009 The Go Authors.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google LLC nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

### lxn/walk

```
Copyright (c) 2010 The Walk Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:
1. Redistributions of source code must retain the above copyright
   notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright
   notice, this list of conditions and the following disclaimer in the
   documentation and/or other materials provided with the distribution.
3. The names of the authors may not be used to endorse or promote products
   derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE AUTHORS ``AS IS'' AND ANY EXPRESS OR
IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

### lxn/win

```
Copyright (c) 2010 The win Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:
1. Redistributions of source code must retain the above copyright
   notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright
   notice, this list of conditions and the following disclaimer in the
   documentation and/or other materials provided with the distribution.
3. The names of the authors may not be used to endorse or promote products
   derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE AUTHORS ``AS IS'' AND ANY EXPRESS OR
IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

### Knetic/govaluate

```
The MIT License (MIT)

Copyright (c) 2014-2016 George Lester

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

</details>

---

バージョン：1.0 (2026/10/31) / 作者：みうかん
