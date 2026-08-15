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

package double_array_trie_test

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	dat "gitee.com/ivfzhou/double-array-trie"
)

const (
	maximumKeySize        = 4096
	maximumKeyLength      = 16
	maximumSentenceLength = 10240
	testTimes             = 10
)

func TestSimple(t *testing.T) {
	keys := []string{
		"AC",
		"AD",
		"ADG",
		"ADH",
		"ADHG",
		"BEIZ",
		"BEL",
		"BF",
		"DG",
	}
	slices.Sort(keys)
	d := dat.New(keys)
	for i, key := range keys {
		if !d.Matches(key) {
			t.Errorf("unmatch key: %s", key)
		}

		if d.MatchIndex(key) != i {
			t.Errorf("unmatch key index: %s", key)
		}
	}

	word := "AD"
	if !d.MatchPrefix(word) {
		t.Errorf("unmatch key prefix: %s", word)
	}

	for _, v := range d.ObtainPrefixes(word) {
		if !strings.HasPrefix(v, word) {
			t.Errorf("unmatch key prefix with %s: %s", word, v)
		}
	}

	sentence := "ABCD EFG HIJK BFEL"
	hits, indexes := d.Analysis(sentence)
	for i, hit := range hits {
		if sentence[indexes[i]:indexes[i]+len(hit)] != hit {
			t.Errorf("sentence not match: %v %v", hit, indexes[i])
		}
	}

	for _, v := range d.HitKeys(word) {
		if !strings.HasPrefix(v, word) {
			t.Errorf("unmatch key prefix: %s", v)
		}
	}

	fmt.Println(d)
}

func TestDat_HitKeys(t *testing.T) {
	for range testTimes {
		keys := randomKeys()
		d := dat.New(keys)
		for _, key := range keys {
			runes := []rune(key)
			key = string(runes[:rand.Intn(len(runes))])
			if len(key) <= 0 {
				continue
			}
			for _, v := range d.HitKeys(key) {
				if !strings.HasPrefix(v, key) {
					t.Errorf("does not match key prefix")
				}
			}
		}
	}
}

func TestDat_Analysis(t *testing.T) {
	for range testTimes {
		sentence := randomSentence()
		keys := randomKeys()
		d := dat.New(keys)
		hitKeys, indexes := d.Analysis(sentence)
		if len(hitKeys) != len(indexes) {
			t.Errorf("wrong number of hit keys")
			continue
		}
		for _, hitKey := range hitKeys {
			if !slices.Contains(keys, hitKey) {
				t.Errorf("key not exists")
			}
		}
		for i, index := range indexes {
			hitKey := hitKeys[i]
			word := sentence[index : index+len(hitKey)]
			if word != hitKey {
				t.Errorf("word mismatch")
			}
		}
	}
}

func TestDat_MatchPrefix(t *testing.T) {
	for range testTimes {
		keys := randomKeys()
		otherKeys := randomKeys()
		d := dat.New(keys)
		for _, key := range keys {
			runes := []rune(key)
			key = string(runes[:rand.Intn(len(runes))])
			if len(key) <= 0 {
				continue
			}
			if !d.MatchPrefix(key) {
				t.Errorf("does not match key prefix")
			}
		}
		for _, otherKey := range otherKeys {
			has := false
			for _, key := range keys {
				if strings.HasPrefix(key, otherKey) {
					has = true
					break
				}
			}
			if d.MatchPrefix(otherKey) != has {
				t.Errorf("unexpected matching key")
			}
		}
	}
}

func TestDat_ObtainPrefixes(t *testing.T) {
	for range testTimes {
		keys := randomKeys()
		otherKeys := randomKeys()
		d := dat.New(keys)
		for _, key := range keys {
			if !slices.Contains(d.ObtainPrefixes(key+"~"), key) {
				t.Errorf("does not obtain key")
			}
		}
		for _, otherKey := range otherKeys {
			hitKey := ""
			for _, key := range keys {
				if strings.HasPrefix(otherKey, key) {
					hitKey = key
					break
				}
			}
			if len(hitKey) > 0 {
				if !slices.Contains(d.ObtainPrefixes(otherKey), hitKey) {
					t.Errorf("does not obtain key")
				}
			} else {
				if len(d.ObtainPrefixes(otherKey)) != 0 {
					t.Errorf("ca not obtain key")
				}
			}
		}
	}
}

func TestDat_MatchGet(t *testing.T) {
	for range testTimes {
		keys := randomKeys()
		keyToValue := make(map[string]any, len(keys))
		for _, key := range keys {
			keyToValue[key] = rand.Intn(math.MaxInt)
		}
		d := dat.NewWithValues(keyToValue)
		for _, key := range keys {
			value, ok := d.MatchGet(key)
			if !ok {
				t.Errorf("key doesn't exist")
				continue
			}
			if keyToValue[key] != value {
				t.Errorf("value mismatch")
			}
		}
	}
}

func TestDat_Matches(t *testing.T) {
	for range testTimes {
		keys := randomKeys()
		otherKeys := randomKeys()
		d := dat.New(keys)
		for _, key := range keys {
			if ok := d.Matches(key); !ok {
				t.Errorf("does not match key")
			}
		}
		for _, otherKey := range otherKeys {
			has := slices.Contains(keys, otherKey)
			if d.Matches(otherKey) != has {
				t.Errorf("unexpected matching key")
			}
		}
	}
}

func TestDat_MatchIndex(t *testing.T) {
	for range testTimes {
		keys := randomKeys()
		otherKeys := randomKeys()
		d := dat.New(keys)
		for _, key := range keys {
			index := d.MatchIndex(key)
			if index < 0 {
				t.Errorf("does not match key")
				continue
			}
			if d.GetKey(index) != key {
				t.Errorf("key index mismatch")
			}
		}
		for _, otherKey := range otherKeys {
			has := slices.Contains(keys, otherKey)
			index := d.MatchIndex(otherKey)
			if has && index < 0 {
				t.Errorf("unexpected matching key")
			}
			if !has && index >= 0 {
				t.Errorf("unexpected matching key")
			}
			if index >= 0 && d.GetKey(index) != otherKey {
				t.Errorf("key index mismatch")
			}
		}
	}
}

func TestDat_DumpAndRead(t *testing.T) {
	for range testTimes {
		keys := randomKeys()
		d := dat.New(keys)
		for _, key := range keys {
			if !d.Matches(key) {
				t.Errorf("does not match key")
			}
			index := d.MatchIndex(key)
			if index < 0 {
				t.Errorf("index mismatch")
				continue
			}
			if d.GetKey(index) != key {
				t.Errorf("key index mismatch")
			}
			if !d.MatchPrefix(key) {
				t.Errorf("prefix mismatch")
			}
		}

		temporaryFilePath := filepath.Join(os.TempDir(), randomReadableString(32)+".dat.gz")
		err := d.DumpToFile(temporaryFilePath)
		if err != nil {
			t.Errorf("failed to dump dat %v", err)
			continue
		}
		newDat, err := dat.ReadFromFile(temporaryFilePath)
		_ = os.Remove(temporaryFilePath)
		if err != nil {
			t.Errorf("failed to read dat %v", err)
			continue
		}

		for _, key := range keys {
			if !newDat.Matches(key) {
				t.Errorf("does not match key")
			}
			index := newDat.MatchIndex(key)
			if index < 0 {
				t.Errorf("index mismatch")
				continue
			}
			if newDat.GetKey(index) != key {
				t.Errorf("key index mismatch")
			}
			if !newDat.MatchPrefix(key) {
				t.Errorf("prefix mismatch")
			}
		}
	}
}

func TestNew_EmptyAndFiltered(t *testing.T) {
	// 空输入。
	if d := dat.New(nil); d.KeySize() != 0 || d.Size() != 0 || d.Depth() != 0 {
		t.Errorf("empty input should produce empty trie")
	}
	// 全部为空串，不应 panic。
	if d := dat.New([]string{"", "", ""}); d.KeySize() != 0 || d.Size() != 0 {
		t.Errorf("all-empty keys should produce empty trie")
	}
	// 去重。
	d := dat.New([]string{"abc", "abc", "abc"})
	if d.KeySize() != 1 {
		t.Errorf("duplicate keys should be removed, got %d", d.KeySize())
	}
	if !d.Matches("abc") {
		t.Errorf("abc should match")
	}
	// 过滤空串并保留有效 key。
	d = dat.New([]string{"", "abc", "", "def", "abc"})
	if d.KeySize() != 2 {
		t.Errorf("expected 2 keys, got %d", d.KeySize())
	}
	if !d.Matches("abc") || !d.Matches("def") {
		t.Errorf("valid keys should match")
	}
	if d.Matches("") {
		t.Errorf("empty string should not match")
	}
}

func TestDat_EmptyTrie(t *testing.T) {
	d := dat.New(nil)
	if d.Size() != 0 || d.KeySize() != 0 || d.Depth() != 0 || d.Hollow() != 0 {
		t.Errorf("empty trie metadata mismatch")
	}
	if d.Matches("a") || d.Matches("") {
		t.Errorf("empty trie should not match")
	}
	if d.MatchPrefix("a") {
		t.Errorf("empty trie should not match prefix")
	}
	if d.MatchIndex("a") != -1 {
		t.Errorf("empty trie should not match index")
	}
	if hits := d.HitKeys("a"); len(hits) != 0 {
		t.Errorf("empty trie should have no hits")
	}
	if got := d.ObtainPrefixes("a"); len(got) != 0 {
		t.Errorf("empty trie should have no prefixes")
	}
	keys, indexes := d.Analysis("abc")
	if len(keys) != 0 || len(indexes) != 0 {
		t.Errorf("empty trie analysis should be empty")
	}
}

func TestDat_Unicode(t *testing.T) {
	keys := []string{"中", "中国", "中国人", "中文", "英文"}
	slices.Sort(keys)
	d := dat.New(keys)

	for i, key := range keys {
		if !d.Matches(key) {
			t.Errorf("unmatch key: %s", key)
		}
		if index := d.MatchIndex(key); index != i {
			t.Errorf("unmatch key index: %s", key)
		}
	}

	if hits := d.HitKeys("中"); !slices.Equal(hits, []string{"中", "中国", "中国人", "中文"}) {
		t.Errorf("HitKeys(中) = %v", hits)
	}
	if got := d.ObtainPrefixes("中国人"); !slices.Equal(got, []string{"中", "中国", "中国人"}) {
		t.Errorf("ObtainPrefixes(中国人) = %v", got)
	}
	if got := d.ObtainPrefixes("英文"); !slices.Equal(got, []string{"英文"}) {
		t.Errorf("ObtainPrefixes(英文) = %v", got)
	}
	if !d.MatchPrefix("中") || !d.MatchPrefix("中国") || !d.MatchPrefix("英文") {
		t.Errorf("should match prefix")
	}
	if d.MatchPrefix("中文的") {
		t.Errorf("should not match prefix")
	}

	hitKeys, indexes := d.Analysis("我爱中国人和中文")
	if len(hitKeys) != len(indexes) {
		t.Fatalf("hit keys and indexes length mismatch")
	}
	for i, k := range hitKeys {
		if word := "我爱中国人和中文"[indexes[i] : indexes[i]+len(k)]; word != k {
			t.Errorf("word mismatch: %s at %d", k, indexes[i])
		}
	}
}

func TestDat_SingleCharAndDedup(t *testing.T) {
	d := dat.New([]string{"a", "b", "c", "a"})
	if d.KeySize() != 3 {
		t.Errorf("expected 3 keys, got %d", d.KeySize())
	}
	for i, key := range []string{"a", "b", "c"} {
		if !d.Matches(key) {
			t.Errorf("unmatch key: %s", key)
		}
		if index := d.MatchIndex(key); index != i {
			t.Errorf("unmatch key index: %s", key)
		}
	}
	if got := d.HitKeys("a"); !slices.Equal(got, []string{"a"}) {
		t.Errorf("HitKeys(a) = %v", got)
	}
}

func TestDat_AnalysisOverlapping(t *testing.T) {
	d := dat.New([]string{"ab", "abc", "bc"})
	keys, indexes := d.Analysis("abc")
	wantKeys := []string{"ab", "abc", "bc"}
	wantIndexes := []int{0, 0, 1}
	if !slices.Equal(keys, wantKeys) {
		t.Errorf("Analysis keys = %v, want %v", keys, wantKeys)
	}
	if !slices.Equal(indexes, wantIndexes) {
		t.Errorf("Analysis indexes = %v, want %v", indexes, wantIndexes)
	}
}

func TestDat_HitKeys_Edge(t *testing.T) {
	d := dat.New([]string{"abc", "def", "jkl"})
	if hits := d.HitKeys(""); len(hits) != 0 {
		t.Errorf("empty prefix should have no hits, got %v", hits)
	}
	if hits := d.HitKeys("abc"); !slices.Equal(hits, []string{"abc"}) {
		t.Errorf("HitKeys(abc) = %v", hits)
	}
	if hits := d.HitKeys("xyz"); len(hits) != 0 {
		t.Errorf("unmatched prefix should have no hits, got %v", hits)
	}

	d2 := dat.New([]string{"a", "ab", "abc"})
	if hits := d2.HitKeys("a"); !slices.Equal(hits, []string{"a", "ab", "abc"}) {
		t.Errorf("HitKeys(a) = %v", hits)
	}
}

func TestDat_ObtainPrefixes_NoMatch(t *testing.T) {
	d := dat.New([]string{"abc"})
	if got := d.ObtainPrefixes("xyz"); len(got) != 0 {
		t.Errorf("should have no prefixes, got %v", got)
	}
}

func TestDat_MatchGet_NoValues(t *testing.T) {
	d := dat.New([]string{"abc"})
	if value, ok := d.MatchGet("abc"); ok || value != nil {
		t.Errorf("value-less trie should not return value, got %v %v", value, ok)
	}
	if value, ok := d.MatchGet("none"); ok || value != nil {
		t.Errorf("unmatched word should not return value, got %v %v", value, ok)
	}
}

func TestDat_NewWithValues_EmptyMap(t *testing.T) {
	d := dat.NewWithValues(map[string]int{})
	if d.KeySize() != 0 {
		t.Errorf("empty map should produce empty trie, got %d keys", d.KeySize())
	}
	if value, ok := d.MatchGet("a"); ok || value != nil {
		t.Errorf("empty trie should not return value, got %v %v", value, ok)
	}
}

func TestDat_DumpRead_RestoresMetadata(t *testing.T) {
	keys := []string{"abc", "def", "abdi", "jkl"}
	d := dat.New(keys)

	temporaryFilePath := filepath.Join(os.TempDir(), randomReadableString(32)+".dat.gz")
	if err := d.DumpToFile(temporaryFilePath); err != nil {
		t.Fatalf("failed to dump dat: %v", err)
	}
	defer os.Remove(temporaryFilePath)

	r, err := dat.ReadFromFile(temporaryFilePath)
	if err != nil {
		t.Fatalf("failed to read dat: %v", err)
	}
	if r.Depth() != d.Depth() {
		t.Errorf("depth mismatch: got %d want %d", r.Depth(), d.Depth())
	}
	if r.Size() != d.Size() {
		t.Errorf("size mismatch: got %d want %d", r.Size(), d.Size())
	}
	if r.KeySize() != d.KeySize() {
		t.Errorf("key size mismatch: got %d want %d", r.KeySize(), d.KeySize())
	}
	if hits := r.HitKeys("ab"); !slices.Equal(hits, []string{"abc", "abdi"}) {
		t.Errorf("HitKeys(ab) after read = %v", hits)
	}
	for _, key := range keys {
		if !r.Matches(key) || !r.MatchPrefix(key) {
			t.Errorf("unmatch key after read: %s", key)
		}
		if r.MatchIndex(key) < 0 {
			t.Errorf("index mismatch after read: %s", key)
		}
	}
}

func TestDat_DumpRead_Empty(t *testing.T) {
	d := dat.New(nil)
	temporaryFilePath := filepath.Join(os.TempDir(), randomReadableString(32)+".dat.gz")
	if err := d.DumpToFile(temporaryFilePath); err != nil {
		t.Fatalf("failed to dump empty dat: %v", err)
	}
	defer os.Remove(temporaryFilePath)

	r, err := dat.ReadFromFile(temporaryFilePath)
	if err != nil {
		t.Fatalf("failed to read empty dat: %v", err)
	}
	if r.KeySize() != 0 || r.Size() != 0 || r.Depth() != 0 {
		t.Errorf("empty trie metadata mismatch: keys=%d size=%d depth=%d", r.KeySize(), r.Size(), r.Depth())
	}
	if hits := r.HitKeys("a"); len(hits) != 0 {
		t.Errorf("empty trie should have no hits, got %v", hits)
	}
}

// BenchmarkDat_MatchGet 基准测试 Dat.MatchGet 的读性能。
func BenchmarkDat_MatchGet(b *testing.B) {
	keys := randomKeys()
	keyToValue := make(map[string]any, len(keys))
	for _, key := range keys {
		keyToValue[key] = rand.Intn(math.MaxInt)
	}
	d := dat.NewWithValues(keyToValue)

	for i := 0; b.Loop(); i++ {
		key := keys[i%len(keys)]
		value, ok := d.MatchGet(key)
		if !ok {
			b.Fatalf("key not found: %s", key)
		}
		_ = value
	}
}

// BenchmarkMap_Read 基准测试原生 map 的读性能，作为对照。
func BenchmarkMap_Read(b *testing.B) {
	keys := randomKeys()
	m := make(map[string]any, len(keys))
	for _, key := range keys {
		m[key] = rand.Intn(math.MaxInt)
	}

	for i := 0; b.Loop(); i++ {
		key := keys[i%len(keys)]
		value, ok := m[key]
		if !ok {
			b.Fatalf("key not found: %s", key)
		}
		_ = value
	}
}

// BenchmarkDat_MatchPrefix 基准测试 Dat.MatchPrefix 的匹配性能。
func BenchmarkDat_MatchPrefix(b *testing.B) {
	keys := randomKeys()
	d := dat.New(keys)

	// 生成查询词：取每个 key 的非空前缀，属于前缀匹配场景。
	words := make([]string, len(keys))
	for i, key := range keys {
		runes := []rune(key)
		words[i] = string(runes[:rand.Intn(len(runes))+1])
	}

	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		word := words[i%len(words)]
		_ = d.MatchPrefix(word)
	}
}

// BenchmarkStrings_HasPrefix 基准测试 strings.HasPrefix 的匹配性能，作为对照。
// MatchPrefix(word) 等价于「词库中存在以 word 为前缀的词」，
// 因此这里用 strings.HasPrefix 遍历词库实现相同语义，保证对比公平。
func BenchmarkStrings_HasPrefix(b *testing.B) {
	keys := randomKeys()

	words := make([]string, len(keys))
	for i, key := range keys {
		runes := []rune(key)
		words[i] = string(runes[:rand.Intn(len(runes))+1])
	}

	for i := 0; b.Loop(); i++ {
		word := words[i%len(words)]
		_ = slices.ContainsFunc(keys, func(key string) bool {
			return strings.HasPrefix(key, word)
		})
	}
}

func randomKeys() []string {
	length := rand.Intn(maximumKeySize) + 1
	keys := make([]string, length)
	for i := range keys {
		keys[i] = randomKey()
	}
	return keys
}

func randomKey() string {
	length := rand.Intn(maximumKeyLength) + 1
	runes := make([]rune, length)
	for i := range length {
		switch rand.Intn(4) {
		case 0:
			runes[i] = rune(rand.Int63n(0x7f + 1))
		case 1:
			runes[i] = rune(rand.Int63n(0x7ff-0x80+1) + 0x80)
		case 2:
			runes[i] = rune(rand.Int63n(0xffff-0x800+1) + 0x800)
		case 3:
			runes[i] = rune(rand.Int63n(0x1fffff-0x10000+1) + 0x10000)
		}
	}
	return string(runes)
}

func randomReadableString(length int) string {
	str := make([]rune, length)
	for i := range str {
	Loop:
		code := rand.Intn(1 << 7)
		switch {
		case code >= '0' && code <= '9':
			fallthrough
		case code >= 'a' && code <= 'z':
			fallthrough
		case code >= 'A' && code <= 'Z':
			str[i] = rune(code)
		default:
			goto Loop
		}
	}
	return string(str)
}

func randomSentence() string {
	length := rand.Intn(maximumSentenceLength) + 1
	str := make([]rune, length)
	for i := range length {
		str[i] = rune(rand.Intn(math.MaxInt32 + 1))
	}
	return string(str)
}
