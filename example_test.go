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

	dat "gitee.com/ivfzhou/double-array-trie"
)

func ExampleDat_HitKeys() {
	d := dat.New([]string{"abc", "def", "abdi", "jkl"})
	fmt.Println(d.HitKeys("ab"))

	// Output:
	// [abc abdi]
}

func ExampleDat_Analysis() {
	d := dat.New([]string{"abc", "ab", "abg", "jkl"})
	sentence := "abc ab"
	hitKeys, indexes := d.Analysis(sentence)
	for i, index := range indexes {
		word := sentence[index : index+len(hitKeys[i])]
		fmt.Println(word == hitKeys[i])
	}

	// Output:
	// true
	// true
}

func ExampleDat_MatchPrefix() {
	d := dat.New([]string{"abc", "def", "ghi", "jkl"})
	fmt.Println(d.MatchPrefix("abc"))
	fmt.Println(d.MatchPrefix("ab"))
	fmt.Println(d.MatchPrefix("cab"))

	// Output:
	// true
	// true
	// false
}

func ExampleDat_ObtainPrefixes() {
	d := dat.New([]string{"abc", "ab", "abg", "jkl"})
	fmt.Println(d.ObtainPrefixes("abc"))

	// Output:
	// [ab abc]
}

func ExampleDat_MatchGet() {
	keyToValue := map[string]any{
		"abc": 1,
		"afr": 2,
		"kja": 3,
	}
	d := dat.NewWithValues(keyToValue)
	fmt.Println(d.MatchGet("abc"))
	fmt.Println(d.MatchGet("none"))

	// Output:
	// 1 true
	// <nil> false
}

func ExampleDat_Matches() {
	d := dat.New([]string{"abc", "def", "ghi", "jkl"})
	fmt.Println(d.Matches("abc"))
	fmt.Println(d.Matches("abcd"))
	fmt.Println(d.Matches("ab"))
	fmt.Println(d.Matches("jkl"))

	// Output:
	// true
	// false
	// false
	// true
}

func ExampleDat_MatchIndex() {
	d := dat.New([]string{"abc", "def", "ghi", "jkl"})
	index := d.MatchIndex("abc")
	fmt.Println(d.GetKey(index) == "abc")

	// Output:
	// true
}
