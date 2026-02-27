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
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

var (
	// MinExpansiveFactor 适配 state 时，数组最小扩容系数。如果同层子节点 code 相差大，则该值该调大。
	MinExpansiveFactor = 1.2
	// InitArrayFactor 初始扩容系数。
	InitArrayFactor = 2.5
)

type Dat[T any] struct {
	// 数据结构状态数组。
	base, check []int
	// 字符与其 code 的映射。
	dictionary map[rune]int
	// 最大的 code 值。
	maximumCodeValue int
	// 深度。
	depth int
	// 词汇。
	keys   []string
	values []T
}

// New 构建双数组树。不处理空串，去除重复。
func New(keys []string) *Dat[any] {
	return newInstance[any](keys)
}

// NewWithValues 构建双数组树。不处理空串。可使用 Dat.MatchGet 根据 key 获取 value。不支持备份到文件。
func NewWithValues[T any](keyToValue map[string]T) *Dat[T] {
	instance := newInstance[T](slices.Collect(maps.Keys(keyToValue)))
	instance.values = make([]T, len(instance.keys))
	for i, key := range instance.keys {
		instance.values[i] = keyToValue[key]
	}
	return instance
}

// ReadFromFile 从文件中恢复双数组树。
func ReadFromFile(filePath string) (*Dat[any], error) {
	// 解压。读取数据。
	fileStream, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if fileStream != nil {
			_ = fileStream.Close()
		}
	}()
	reader, err := gzip.NewReader(fileStream)
	if err != nil {
		return nil, err
	}
	defer func() {
		if reader != nil {
			_ = reader.Close()
		}
	}()
	fileData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if err = reader.Close(); err != nil {
		return nil, err
	}

	// 生成对象。
	d := &Dat[any]{}
	lines := strings.Split(string(fileData), "\n")
	if len(lines) != 4 {
		return nil, errors.New("invalid file format")
	}
	bases := strings.Split(lines[0], ",")
	checks := strings.Split(lines[1], ",")
	dicts := strings.Split(lines[2], ",")
	keys := strings.Split(lines[3], ";")

	// 处理 base。
	d.base = make([]int, len(bases))
	for i, v := range bases {
		d.base[i], err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
	}
	bases = nil

	// 处理 check。
	d.check = make([]int, len(checks))
	for i, v := range checks {
		d.check[i], err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
	}
	checks = nil

	// 处理字典。
	d.dictionary = make(map[rune]int, len(dicts))
	for _, v := range dicts {
		pair := strings.Split(v, "=")
		if len(pair) != 2 {
			return nil, errors.New("invalid dictionary format")
		}
		char, err2 := strconv.Atoi(pair[0])
		if err2 != nil {
			return nil, err2
		}
		code, err2 := strconv.Atoi(pair[1])
		if err2 != nil {
			return nil, err2
		}
		d.dictionary[rune(char)] = code
	}
	dicts = nil

	// 处理词汇。
	d.keys = make([]string, len(keys))
	for i, v := range keys {
		runesString := strings.Split(v, ",")
		runes := make([]rune, len(runesString))
		for j, char := range runesString {
			r, err2 := strconv.Atoi(char)
			if err2 != nil {
				return nil, err2
			}
			runes[j] = rune(r)
		}
		d.keys[i] = string(runes)
	}

	return d, nil
}

// HitKeys 获取词库中前缀部分是 prefix 的词。包含 prefix 与词完全一样的情形。
func (d *Dat[T]) HitKeys(prefix string) (hits []string) {
	runes := []rune(prefix)
	state := 1

	// 从一个节点开始查找。
	var seekNodes func(state, depth int)
	seekNodes = func(state, depth int) {
		if depth > d.depth {
			return
		}

		for code := 1; code <= d.maximumCodeValue; code++ {
			// 不存在的匹配。
			if code+state-2 >= len(d.check) {
				continue
			}

			// check 不过，说明字符不匹配。
			if d.check[code+state-2] != state {
				continue
			}

			// 更新下次的状态值。
			newState := d.base[code+state-2]

			// 检查是否为终止点。
			if base := d.base[newState-2]; base < 0 {
				key := d.keys[-base-1]
				hits = append(hits, key)
			}

			seekNodes(newState, depth+1)
		}
	}

	depth := 1
	for i := 0; i < len(runes); i++ {
		code, ok := d.dictionary[runes[i]]

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
		}

		// 更新下次的状态值。
		state = d.base[code+state-2]

		depth++

		// 遍历子节点。
		if i == len(runes)-1 {
			// 检查是否为终止点。
			if base := d.base[state-2]; base < 0 {
				key := d.keys[-base-1]
				hits = append(hits, key)
			}

			seekNodes(state, depth)
		}
	}
	return
}

// Analysis 分析 sentence，获取 sentence 中包含词库的词，和该词在 sentence 中的位置。
func (d *Dat[T]) Analysis(sentence string) (keys []string, indexes []int) {
	runes := []rune(sentence)
	hits := make(map[string]struct{}, len(d.keys))

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
			}

			// 更新下次的状态值。
			state = d.base[code+state-2]

			// 检查是否为终止点。
			if base := d.base[state-2]; base < 0 {
				key := d.keys[-base-1]
				if _, ok = hits[key]; !ok {
					indexes = append(indexes, len(string(runes[:i])))
					keys = append(keys, key)
					hits[key] = struct{}{}
				}
			}
		}

	}

	return
}

// MatchPrefix 判断 word 是否是词库中任何一个词的前缀。包含 word 与词完全一样的情形。
func (d *Dat[T]) MatchPrefix(word string) bool {
	state := 1

	for _, char := range word {
		code, ok := d.dictionary[char]

		// 字典无该字符说明不匹配。
		if !ok {
			return false
		}

		// 不存在的匹配。
		if code+state-2 >= len(d.check) {
			return false
		}

		// check 不过，说明字符不匹配。
		if d.check[code+state-2] != state {
			return false
		}

		// 更新下次的状态值。
		state = d.base[code+state-2]
	}

	return true
}

// ObtainPrefixes 返回所有匹配 word 前缀的词。
func (d *Dat[T]) ObtainPrefixes(word string) (result []string) {
	state := 1

	for _, char := range word {
		code, ok := d.dictionary[char]

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
		}

		// 更新下次的状态值。
		state = d.base[code+state-2]

		// 检查是否为终止点。
		if base := d.base[state-2]; base < 0 {
			result = append(result, d.keys[-base-1])
		}
	}

	return
}

// MatchGet 判断 word 是否在词库中，返回词对应的 value。
func (d *Dat[T]) MatchGet(word string) (value any, ok bool) {
	index := d.MatchIndex(word)
	if index >= 0 {
		return d.values[index], true
	}
	return
}

// Matches 判断 word 是否存在于词库中。
func (d *Dat[T]) Matches(word string) bool {
	state := 1

	for _, char := range word {
		code, ok := d.dictionary[char]

		// 字典无该字符说明不匹配。
		if !ok {
			return false
		}

		// 不存在的匹配。
		if code+state-2 >= len(d.check) {
			return false
		}

		// check 不过，说明字符不匹配。
		if d.check[code+state-2] != state {
			return false
		}

		// 更新下次的状态值。
		state = d.base[code+state-2]
	}

	// 检查最后一位是不是空节点。
	if state >= 2 && d.base[state-2] < 0 {
		return true
	}

	return false
}

// Deprecated: 使用 MatchIndex。
func (d *Dat[T]) MatchesIndex(word string) int {
	return d.MatchIndex(word)
}

// MatchIndex 检索 word 是否在词库中，并返回词位置，不存在返回 -1。
func (d *Dat[T]) MatchIndex(word string) int {
	state := 1
	chars := []rune(word)
	length := len(chars)

	for i, char := range chars {
		code, ok := d.dictionary[char]

		// 字典无该字符说明不匹配。
		if !ok {
			return -1
		}

		// 不存在的匹配。
		if code+state-2 >= len(d.check) {
			return -1
		}

		// check 不过说明字符不匹配。
		if d.check[code+state-2] != state {
			return -1
		}

		// 更新下次的状态值。
		state = d.base[code+state-2]

		// 是最后一个字符，且是终止节点。
		if i == length-1 {
			if base := d.base[state-2]; base < 0 {
				return -base - 1
			}
		}
	}

	return -1
}

// Size 数组长度。
func (d *Dat[T]) Size() int {
	return len(d.base)
}

// KeySize 词库中词的个数。
func (d *Dat[T]) KeySize() int {
	return len(d.keys)
}

// Hollow 返回数组未使用的下标个数。
func (d *Dat[T]) Hollow() int {
	count := 0
	for _, v := range d.check {
		if v == 0 {
			count++
		}
	}
	return count
}

// GetKey 获取词。
func (d *Dat[T]) GetKey(index int) string {
	return d.keys[index]
}

// DumpToFile 持久化检索树。
func (d *Dat[T]) DumpToFile(filePath string) error {
	// 处理数组。
	var (
		buf   = &bytes.Buffer{}
		base  = &bytes.Buffer{}
		check = &bytes.Buffer{}
	)
	for i := 0; i < len(d.check)-1; i++ {
		_, _ = fmt.Fprintf(base, "%d,", d.base[i])
		_, _ = fmt.Fprintf(check, "%d,", d.check[i])
	}
	_, _ = fmt.Fprintf(base, "%d", d.base[len(d.check)-1])
	_, _ = fmt.Fprintf(check, "%d", d.check[len(d.check)-1])
	_, _ = fmt.Fprint(buf, base.String()+"\n"+check.String()+"\n")
	base, check = nil, nil

	// 处理字典。
	for char, code := range d.dictionary {
		_, _ = fmt.Fprintf(buf, "%d=%d,", char, code)
	}
	if len(d.dictionary) > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	_, _ = fmt.Fprintf(buf, "\n")

	// 处理词汇。
	for i := 0; i < len(d.keys)-1; i++ {
		runes := []rune(d.keys[i])
		for j := 0; j < len(runes)-1; j++ {
			_, _ = fmt.Fprintf(buf, "%d,", runes[j])
		}
		if len(runes) > 0 {
			_, _ = fmt.Fprintf(buf, "%d", runes[len(runes)-1])
		}
		_, _ = fmt.Fprintf(buf, ";")
	}
	if len(d.keys) > 0 {
		runes := []rune(d.keys[len(d.keys)-1])
		for i := 0; i < len(runes)-1; i++ {
			_, _ = fmt.Fprintf(buf, "%d,", runes[i])
		}
		if len(runes) > 0 {
			_, _ = fmt.Fprintf(buf, "%d", runes[len(runes)-1])
		}
	}

	// 压缩。
	gzData := &bytes.Buffer{}
	writer := gzip.NewWriter(gzData)
	_, err := io.Copy(writer, buf)
	if err != nil {
		_ = writer.Close()
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}

	// 写入文件。
	if err = os.MkdirAll(filepath.Dir(filePath), 0744); err != nil {
		return err
	}
	if err = os.WriteFile(filePath, gzData.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

// String 输出字符串。
func (d *Dat[T]) String() string {
	return fmt.Sprintf("base:  %v\ncheck: %v\n", d.base, d.check)
}

// Depth 检索树深度。
func (d *Dat[T]) Depth() int {
	return d.depth
}

func newInstance[T any](keys []string) *Dat[T] {
	// 处理空串，去除重复。
	copiedKeys := make([]string, 0, len(keys))
	keyMap := make(map[string]struct{}, len(keys))
	for _, v := range keys {
		if len(v) <= 0 {
			continue
		}
		if _, ok := keyMap[v]; ok {
			continue
		}
		keyMap[v] = struct{}{}
		copiedKeys = append(copiedKeys, v)
	}
	keyMap = nil
	if len(keys) == 0 {
		return &Dat[T]{}
	}

	// 裁剪数组。
	newKeys := make([]string, len(copiedKeys))
	copy(newKeys, copiedKeys)
	copiedKeys = nil

	// 词汇排序。
	slices.Sort(newKeys)

	// 初始化实例。
	instance := &Dat[T]{keys: newKeys}

	// 生成字典。
	charMap := make(map[rune]struct{}, len(instance.keys))
	charArr := make([]rune, 0, len(instance.keys)*2)
	for _, key := range instance.keys {
		for _, char := range key {
			if _, ok := charMap[char]; ok {
				continue
			}
			charMap[char] = struct{}{}
			charArr = append(charArr, char)
		}
	}
	charMap = nil
	slices.Sort(charArr)
	instance.dictionary = make(map[rune]int, len(charArr))
	for i, char := range charArr {
		instance.dictionary[char] = i + 1
	}
	instance.maximumCodeValue = len(charArr)
	charArr = nil

	// node 节点对象。code 为字符状态值，left 到 right 为其子节点范围，depth 为节点层级。
	type node struct {
		code, left, right, depth int
	}

	var (
		// 扩容双数组。
		resizeArray func(newSize int)
		// 拉取子节点。
		fetchChildren func(*node) []*node
		// 找出符合父子转移方程的 state 值。
		findState func([]*node) int
		// 构建树。
		buildTrie func([]*node) int
		// 标记已使用了的位置。
		usedState = make(map[int]struct{})
		// 记录数组已使用了的长度。
		size = 0
		// 数组长度。
		allocatedSize = 0
	)

	// 扩容双数组。
	resizeArray = func(size int) {
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

	// 获取子节点。instance.keys 需要是有序的。
	fetchChildren = func(parent *node) (children []*node) {
		// 上一个节点的 code。
		previousCode := 0

		for i := parent.left; i < parent.right; i++ {
			// 代表树的一个分支。
			word := []rune(instance.keys[i])

			if parent.depth > len(word) { // 链路已设置完。
				continue
			} else if parent.depth < len(word) { // 找出节点 code。
				code, ok := instance.dictionary[word[parent.depth]]
				if !ok || code == 0 {
					panic("dictionary is corrupt")
				}

				// 相同节点不添加。
				if code != previousCode {
					newNode := &node{code, i, 0, parent.depth + 1}
					if len(children) != 0 {
						// 设置左相邻节点的 right。
						children[len(children)-1].right = i
					}
					children = append(children, newNode)
				}

				previousCode = code
			} else { // 终止节点，添加它。
				children = append(children, &node{0, i, -1, parent.depth + 1})
			}
		}

		// 设置左相邻节点 right。
		if len(children) != 0 {
			children[len(children)-1].right = parent.right
		}

		return
	}

	// 找出所有子节点都能使用的 state 值。并自动扩容数组。发生冲突时，线性向后加一位查找空闲。
	findState = func(children []*node) (state int) {
		maximumIndex := children[len(children)-1].code - 2
		firstCode := children[0].code

		// 获取 check 第一个为 0 的位置。
		startState := 2
		if firstCode != 0 {
			startState = 1
		}
		for ; ; startState++ {
			index := firstCode + startState - 2
			if len(instance.check) <= index {
				break
			}
			if instance.check[index] == 0 {
				break
			}
		}

	Loop:
		for position := firstCode + startState; ; position++ {
			state = position - firstCode

			// state 须还未使用。
			if _, ok := usedState[state]; ok {
				continue
			}

			// 数组下标过大需扩容。
			if maxIndex := state + maximumIndex; maxIndex >= allocatedSize {
				resizeArray(int(float64(maxIndex+1) * max(MinExpansiveFactor, 1)))
			}

			// check 位不能是已占用。
			if instance.check[position-2] != 0 {
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
		usedState[state] = struct{}{}

		// 记录实际使用有效位数。
		if usedSize := state + children[len(children)-1].code - 1; size < usedSize {
			size = usedSize
		}

		return state
	}

	// 构建树。
	buildTrie = func(nodes []*node) int {
		if nodes[0].depth > instance.depth {
			instance.depth = nodes[0].depth
		}

		state := findState(nodes)

		// 设置 check。
		for _, v := range nodes {
			instance.check[v.code+state-2] = state
		}

		for _, v := range nodes {
			// 拉取子节点。
			children := fetchChildren(v)

			// v 是终止节点。
			if len(children) == 0 {
				// 设置 base，值代表树链路序号。
				instance.base[v.code+state-2] = -v.left - 1
			} else {
				childState := buildTrie(children)
				// 设置完子节点后再设置父节点 base。
				instance.base[v.code+state-2] = childState
			}
		}

		// 返回子节点 state 给父节点设置。
		return state
	}

	// 初始化数组长度。
	resizeArray(int(float64(len(instance.keys)) * max(InitArrayFactor, 0.1)))

	// 构建树。
	buildTrie(fetchChildren(&node{0, 0, len(instance.keys), 0}))
	instance.depth--

	// 缩减数组长度。
	resizeArray(size)

	return instance
}
