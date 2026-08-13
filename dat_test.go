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
