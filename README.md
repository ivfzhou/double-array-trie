# 一、说明

双数组检索树。

[![codecov](https://codecov.io/gh/ivfzhou/double-array-trie/graph/badge.svg?token=JO8RFXW1SP)](https://codecov.io/gh/ivfzhou/double-array-trie)
[![Go Reference](https://pkg.go.dev/badge/gitee.com/ivfzhou/double-array-trie.svg)](https://pkg.go.dev/gitee.com/ivfzhou/double-array-trie)

状态转移方程：
```txt
firstState = 1

父节点：
base[parentCode + parentState - 2] = childState
check[parentCode + parentState - 2] = parentState

子节点：
base[childCode + childState - 2] = grandchildState
check[childCode + childState - 2] = childState

......

终节点：
base[code + state - 2] < 0

code: 字符编码值。对应词汇中的一个字符。
state: 状态值，从 1 开始。
```

例子：

    排好了序的词汇：
    []string{
    0     AC
    1     AD
    2     ADG
    3     ADH
    4     ADHG
    5     BEIZ
    6     BEL
    7     BF
    8     DG
    }
    code  A=1 B=2 C=3 D=4 E=5 F=6 G=7 H=8 I=9 L=10 Z=11

    双数组值，负值表示为该节点对应词汇的下标：
    base  [3 7 -1 16 4 8 -2 -3 -4 -5 12 19 -6 9 10 11 -7 -8 -9 13 18 20 14]
    check [1 1 4  1  3 3 8  9  10 11 7  7  14 8 8  10 18 19 20 12 12 16 13]

    树结构：
    depth
      0                            Root
                                   ⁰ ⁹
                         /                         \           \
      1                 A                          B           D    
                       ⁰ ⁵                         ⁵ ⁷         ⁸ ⁹ 
             /     /      \                     /        \      \
      2     C     D        D                  E           F      G 
           ⁰ ¹            ¹ ⁵                ⁵ ⁷         ⁷ ⁷    ⁸ ⁹
                /     /      \              /    \
      3        G     H        H            I      L
              ² ³            ³ ⁵          ⁵ ⁶    ⁶ ⁷                  
                               \         /
      4                         G       Z
                               ⁴ ⁵     ⁵ ⁶

# 二、使用

```shell
go get gitee.com/ivfzhou/double-array-trie@latest
```

```golang
import dat "gitee.com/ivfzhou/double-array-trie"

// 需要被检索的数据
data := []string{
	"/api/user/info",
	"/api/user/register",
	"/api/user/login",
}

// 构建双数组
d := dat.New(data)

// 获取有 /api 前缀的词。
array := d.HitKeys("/api")

// 判断是否存在与数据中
ok := d.Matches(word)

// 判断数据中是否有前缀和 path 匹配
ok := d.MatchPrefix(path)

// 获取 path 在字典中的索引
index := d.MatchesIndex(path)
key := d.GetKey(index)

// 返回所有匹配前缀 path 的数据
keys := d.ObtainPrefixes(path)

// 解析 sentence，返回数据中出现在 sentence 中的数据和其位置
keys, indexes := d.Analysis("sentence")

// 若匹配词汇，就返回对应 value。
d := dat.NewWithValue(map[string]any{
	"/api/user/info":     func() {},
	"/api/user/register": func() {},
	"/api/user/login":    func() {},
})
value, ok := d.MatchGet(word)
if ok {
	value()
}
```
