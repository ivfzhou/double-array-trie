/*
 * Copyright (c) 2023 ivfzhou
 * double-array-trie is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package double_array_trie

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	// MinExpansiveFactor 适配 state 时，数组最小扩容系数。如果同层子节点 code 相差大，则该值该调大。
	MinExpansiveFactor = 1.2
	// InitArrayFactor 初始扩容系数。
	InitArrayFactor = 2.5
)

type dat struct {
	// 数据结构状态数组。
	base, check []int
	// 字符与其 code 的映射。
	dictionary map[rune]int
	// 词汇。
	keys []string
	// 数组未使用个数。
	hollowNum int
}

// New 构建双数组树。
func New(keys []string) *dat {
	if len(keys) == 0 {
		return nil
	}

	// 复制一份词汇。
	copiedKeys := make([]string, len(keys))
	_ = copy(copiedKeys, keys)

	// 词汇排序。
	sort.Slice(copiedKeys, func(i, j int) bool {
		return copiedKeys[i] < copiedKeys[j]
	})

	// 初始化实例。
	instance := &dat{
		dictionary: make(map[rune]int, len(keys)*2),
		keys:       copiedKeys,
		hollowNum:  -1,
	}

	// 生成字典。
	charMap := make(map[rune]struct{})
	charArr := make([]rune, 0)
	for _, key := range instance.keys {
		for _, char := range key {
			if _, ok := charMap[char]; !ok {
				charArr = append(charArr, char)
				charMap[char] = struct{}{}
			}

		}
	}
	charMap = nil
	sort.Slice(charArr, func(i, j int) bool {
		return charArr[i] < charArr[j]
	})
	for i, char := range charArr {
		instance.dictionary[char] = i + 1
	}
	charArr = nil

	// node 节点对象。code 为字符状态值，left 到 right 为其子节点范围，depth 为节点层级。
	type node struct {
		code, left, right, depth int
	}

	var (
		// resize 扩容双数组。
		resize func(newSize int)
		// fetch 拉取子节点。
		fetch func(*node) []*node
		// search 找出符合父子转移方程的 state 值。
		search func([]*node) int
		// build 构建树。
		build func([]*node) int
		// used 标记已使用了的位置。
		used = make(map[int]bool)
		// previousState 上次 state 值。
		previousState = 0
		// size 记录数组已使用了的长度。
		size = 0
		// allocatedSize 数组长度。
		allocatedSize = 0
	)

	// 扩容双数组。
	resize = func(size int) {
		newBase := make([]int, size)
		newCheck := make([]int, size)
		if allocatedSize > 0 {
			copy(newBase, instance.base)
			copy(newCheck, instance.check)
		}
		allocatedSize = size
		instance.base = newBase
		instance.check = newCheck
	}

	// 拉取子节点。
	fetch = func(parent *node) (children []*node) {
		// 上一个节点的 code。
		prevCode := 0

		for code, i := 0, parent.left; i < parent.right; i++ {
			// 代表树的一个链路。
			word := []rune(instance.keys[i])

			// 链路已设置完。
			if parent.depth > len(word) {
				continue
			}

			// 找出节点 code。
			if parent.depth != len(word) {
				code = instance.dictionary[word[parent.depth]]
			}

			// 相同节点不添加。
			// 终节点添加一次，即 code = 0 and prevCode = 0。
			if code != prevCode || len(children) == 0 {
				newNode := &node{code, i, 0, parent.depth + 1}
				if len(children) != 0 {
					// 设置上一个节点的 right。
					children[len(children)-1].right = i
				}
				children = append(children, newNode)
			}

			prevCode = code
		}

		// 设置上一个节点 right。
		if len(children) != 0 {
			children[len(children)-1].right = parent.right
		}

		return
	}

	// 找出符合父子节点转移方程的 state 值。
	search = func(children []*node) (state int) {
	Loop:
		for pos := children[0].code + previousState + 1; ; pos++ {

			state = pos - children[0].code
			// state 须还未使用。
			if used[state] {
				continue
			}

			// 数组下标过大需扩容。
			if maxIndex := state + children[len(children)-1].code - 2; maxIndex >= allocatedSize {
				resize(int(float64(maxIndex+1) * math.Max(MinExpansiveFactor, 1)))
			}

			// check 位不能是已占用。
			if instance.check[pos-2] != 0 {
				continue
			}

			// 检查 state 符合所有子节点。
			for i := 1; i < len(children); i++ {
				if instance.check[children[i].code+state-2] != 0 {
					continue Loop
				}
			}

			break
		}
		// 标记 state 已使用。
		used[state] = true
		// 记录实际使用有效位数。
		if usedSize := state + children[len(children)-1].code - 1; size < usedSize {
			size = usedSize
		}

		return state
	}

	// 构建树。
	build = func(children []*node) int {
		state := search(children)

		// 设置 check。
		for _, child := range children {
			instance.check[child.code+state-2] = state
		}

		// 继续拉取子节点。
		for _, child := range children {
			// 拉取孙子节点。
			grandchildren := fetch(child)

			// child 是终止点。
			if len(grandchildren) == 0 {
				// 设置 base，值代表树链路序号。
				instance.base[child.code+state-2] = -child.left - 1
			} else {
				childState := build(grandchildren)
				// 设置完子节点后再设置父节点 base。
				instance.base[child.code+state-2] = childState
			}
		}

		// 返回子节点 state 给父节点设置。
		return state
	}

	// 初始化数组长度。
	resize(int(float64(len(instance.keys)) * math.Max(InitArrayFactor, 0.1)))
	// 构建数。
	build(fetch(&node{0, 0, len(instance.keys), 0}))
	// 缩减数组长度。
	resize(size)

	return instance
}

// ReadFromFile 从备份文件中恢复对象。
func ReadFromFile(filePath string) (*dat, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	reader, err := gzip.NewReader(bytes.NewReader(file))
	if err != nil {
		return nil, err
	}

	file, err = io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	d := &dat{}
	lines := strings.Split(string(file), "\n")
	bases := strings.Split(lines[0], ",")
	checks := strings.Split(lines[1], ",")
	dicts := strings.Split(lines[2], ";")

	d.base = make([]int, len(bases))
	for i, v := range bases {
		d.base[i], _ = strconv.Atoi(v)
	}
	bases = nil

	d.check = make([]int, len(checks))
	for i, v := range checks {
		d.check[i], _ = strconv.Atoi(v)
	}
	checks = nil

	d.dictionary = make(map[rune]int, len(dicts))
	for _, v := range dicts[:len(dicts)-1] {
		m := strings.Split(v, "=")
		char, _ := strconv.Atoi(m[0])
		code, _ := strconv.Atoi(m[1])
		d.dictionary[rune(char)] = code
	}
	dicts = nil

	lines = lines[3:]
	d.keys = make([]string, len(lines))
	for i, v := range lines {
		d.keys[i] = v
	}

	return d, nil
}

// Matches 判断 word 是否存在于词库。
func (d *dat) Matches(word string) bool {
	if len(word) == 0 {
		return false
	}

	state := 1

	for _, r := range word {
		code, ok := d.dictionary[r]
		// 字典无该字符说明不匹配。
		if !ok {
			return false
		}

		// 不存在的匹配。
		if code+state-2 >= len(d.check) {
			return false
		}

		// check 不过说明字符不匹配。
		if d.check[code+state-2] != state {
			return false
		} else {
			// 更新下次的状态值。
			state = d.base[code+state-2]
		}
	}

	// 检查最后一位是不是空节点。
	if d.base[state-2] < 0 {
		return true
	}

	return false
}

// MatchPrefix 判断 word 是否匹配词库中任何一个词的前缀。
func (d *dat) MatchPrefix(word string) bool {
	if len(word) == 0 {
		return false
	}

	state := 1

	for _, r := range word {
		code, ok := d.dictionary[r]
		// 字典无该字符说明不匹配。
		if !ok {
			return false
		}

		// 不存在的匹配。
		if code+state-2 >= len(d.check) {
			return false
		}

		// check 不过说明字符不匹配。
		if d.check[code+state-2] != state {
			return false
		} else {
			// 更新下次的状态值。
			state = d.base[code+state-2]
		}
	}

	return true
}

// ObtainPrefixes 返回所有匹配 word 前缀的词。
func (d *dat) ObtainPrefixes(word string) (res []string) {
	if len(word) == 0 {
		return nil
	}

	state := 1

	for _, r := range word {
		code, ok := d.dictionary[r]
		// 字典无该字符说明不匹配。
		if !ok {
			return
		}

		// 不存在的匹配。
		if code+state-2 >= len(d.check) {
			return
		}

		// check 不过，说明字符不匹配。
		if d.check[code+state-2] != state {
			return
		} else {
			// 更新下次的状态值。
			state = d.base[code+state-2]

			// 检查是否为终止点。
			if base := d.base[state-2]; base < 0 {
				res = append(res, d.keys[-base-1])
			}

		}
	}

	return
}

// Analysis 返回词库中词匹配句子 sentence 中词的词和对应起始位置。
func (d *dat) Analysis(sentence string) (keys []string, indexes []int) {
	runes := []rune(sentence)
	words := make(map[string]bool, len(runes))

Loop:
	for i := range runes {

		state := 1
		for j := i; j < len(runes); j++ {
			code, ok := d.dictionary[runes[j]]
			// 字典无该字符说明不匹配。
			if !ok {
				continue Loop
			}

			// 不存在的匹配。
			if code+state-2 >= len(d.check) {
				continue Loop
			}

			// check 不过，说明字符不匹配。
			if d.check[code+state-2] != state {
				continue Loop
			} else {
				// 更新下次的状态值。
				state = d.base[code+state-2]

				// 检查是否为终止点。
				if base := d.base[state-2]; base < 0 {
					key := d.keys[-base-1]
					if !words[key] {
						indexes = append(indexes, len([]byte(string(runes[:i]))))
						keys = append(keys, key)
						words[key] = true
					}
				}

			}
		}

	}

	return
}

// Size 数组长度。
func (d *dat) Size() int {
	return len(d.base)
}

// KeySize 词汇个数。
func (d *dat) KeySize() int {
	return len(d.keys)
}

// Hollow 返回数组未使用的下标个数。
func (d *dat) Hollow() int {
	if d.hollowNum < 0 {
		count := 0
		for _, v := range d.check {
			if v == 0 {
				count++
			}
		}
		d.hollowNum = count
	}
	return d.hollowNum
}

// DumpToFile 备份树。
func (d *dat) DumpToFile(filePath string) error {
	var (
		buf   = &bytes.Buffer{}
		base  = &bytes.Buffer{}
		check = &bytes.Buffer{}
	)
	for i := 0; i < len(d.check)-1; i++ {
		_, _ = base.WriteString(fmt.Sprintf("%d,", d.base[i]))
		_, _ = check.WriteString(fmt.Sprintf("%d,", d.check[i]))
	}
	_, _ = base.WriteString(fmt.Sprintf("%d", d.base[len(d.check)-1]))
	_, _ = check.WriteString(fmt.Sprintf("%d", d.check[len(d.check)-1]))
	_, _ = buf.WriteString(base.String() + "\n" + check.String() + "\n")
	base, check = nil, nil

	for char, code := range d.dictionary {
		_, _ = buf.WriteString(fmt.Sprintf("%d=%d;", char, code))
	}
	_, _ = buf.WriteString("\n")

	for _, v := range d.keys[:len(d.keys)-1] {
		_, _ = buf.WriteString(fmt.Sprintf("%s\n", v))
	}
	_, _ = buf.WriteString(d.keys[len(d.keys)-1])

	gz := bytes.NewBuffer(nil)
	writer := gzip.NewWriter(gz)
	_, err := writer.Write(buf.Bytes())
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, gz.Bytes(), 0b_111100100)
	if err != nil {
		return err
	}

	return nil
}
