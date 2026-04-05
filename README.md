# 一、说明

[![codecov](https://codecov.io/gh/ivfzhou/double-array-trie/graph/badge.svg?token=JO8RFXW1SP)](https://codecov.io/gh/ivfzhou/double-array-trie)
[![Go Reference](https://pkg.go.dev/badge/gitee.com/ivfzhou/double-array-trie.svg)](https://pkg.go.dev/gitee.com/ivfzhou/double-array-trie)

基于 **双数组 Trie（Double-Array Trie）** 算法实现的 Go 高性能字符串检索库，支持前缀匹配、全词匹配、分词分析及键值对存储。

# 二、特性

- **高效检索** — 基于 `base` / `check` 双数组实现，查询时间复杂度仅 \(O(m)\)（\(m\) 为输入串长度），与词库规模无关
- **泛型支持** — 通过 `NewWithValues[T]` 支持任意类型的值绑定（`Map[string]T`）
- **持久化** — 支持 `DumpToFile` / `ReadFromFile` 将检索树序列化到文件并恢复（gzip 压缩）
- **Unicode 安全** — 基于 `rune` 操作，原生支持中文等多字节字符集
- **零依赖** — 仅使用 Go 标准库

# 三、安装

```shell
go get gitee.com/ivfzhou/double-array-trie@latest
```

> 要求 Go >= 1.21（使用泛型）

# 四、快速开始

## 4.1 基本用法

```go
import dat "gitee.com/ivfzhou/double-array-trie"

data := []string{
    "/api/user/info",
    "/api/user/register",
    "/api/user/login",
}

// 构建双数组 Trie
d := dat.New(data)

// 前缀搜索：获取所有以 "/api" 开头的词汇
keys := d.HitKeys("/api")  // ["/api/user/info", "/api/user/login", ...]

// 全词匹配
ok := d.Matches("/api/user/login")  // true
ok := d.Matches("/api")              // false

// 前缀判断：判断输入是否为某词汇的前缀
ok := d.MatchPrefix("/api/user")     // true

// 获取索引与原始词汇
index := d.MatchIndex("/api/user/login")
key := d.GetKey(index)               // "/api/user/login"

// 获取 sentence 中所有命中的词汇及其位置
keys, indexes := d.Analysis("some /api/user/login text")

// 获取所有是 word 前缀的词汇
prefixes := d.ObtainPrefixes("/api/user/login")
```

## 4.2 键值存储

```go
d := dat.NewWithValues(map[string]any{
    "/api/user/info":     handlerInfo,
    "/api/user/register": handlerRegister,
    "/api/user/login":    handlerLogin,
})

value, ok := d.MatchGet("/api/user/info")
if ok {
    value.(func())()
}
```

## 4.3 持久化

```go
// 序列化到文件
err := d.DumpToFile("./dat.dat.gz")

// 从文件恢复
d, err := dat.ReadFromFile("./dat.dat.gz")
```

# 五、API 参考

| 方法 | 说明 |
|------|------|
| `New(keys []string)` | 构建双数组 Trie，自动去重、忽略空串 |
| `NewWithValues[T](map[string]T)` | 构建带值的 Trie，可通过 `MatchGet` 取值（不支持持久化） |
| `ReadFromFile(path) (*Dat, error)` | 从 gzip 文件恢复 Trie |
| `Matches(word string) bool` | 判断 word 是否**精确匹配**词库中的某个词 |
| `MatchIndex(word string) int` | 返回 word 在排序后词库中的索引，未找到返回 `-1` |
| `MatchGet(word string) (value, ok)` | 匹配时返回对应的 value（需通过 `NewWithValues` 构建） |
| `MatchPrefix(word string) bool` | 判断 word 是否为词库中某词的**前缀** |
| `ObtainPrefixes(word string) []string` | 返回所有是 word 前缀的词（按字典序） |
| `HitKeys(prefix string) []string` | 返回所有以 prefix 为前缀的词（含 prefix 本身） |
| `Analysis(sentence string) (keys []string, indexes []int)` | 分词分析，返回 sentence 中出现的词及其起始位置 |
| `GetKey(index int) string` | 根据索引获取原始词汇 |
| `DumpToFile(filePath string) error` | 将 Trie 序列化为 gzip 压缩文件 |
| `Size() int` | 内部数组长度 |
| `KeySize() int` | 词库中有效词汇数 |
| `Hollow() int` | 数组中空闲位置数（空间利用率参考） |
| `Depth() int` | Trie 树的最大深度 |

# 六、算法原理

双数组 Trie 用两个一维数组 `base[]` 和 `check[]` 表示 Trie 树结构：

```
状态转移方程：

  firstState = 1

  base[parentCode + parentState - 2] = childState
  check[parentCode + parentState - 2] = parentState

  终止节点特征：
  base[code + state - 2] < 0    （负值为对应词汇的下标）

  其中：
    code  → 字符编码值（字符在字典中的序号）
    state → 状态值（从 1 开始）
```

# 七、示例

给定已排序的词汇表和字符编码映射：

```
词汇（已排序）：            字符编码：
  0  AC                       A=1  B=2  C=3  D=4  E=5
  1  AD                       F=6  G=7  H=8  I=9  L=10
  2  ADG                      Z=11
  3  ADH
  4  ADHG
  5  BEIZ
  6  BEL
  7  BF
  8  DG
```

构建后的双数组（负值表示终止节点对应的词汇下标）：

```
base  [3 7 -1 16 4 8 -2 -3 -4 -5 12 19 -6 9 10 11 -7 -8 -9 13 18 20 14]
check [1 1 4  1  3 3 8  9  10 11 7  7  14 8 8  10 18 19 20 12 12 16 13]
```

对应的树结构：

```
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
```

# 八、配置项

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MinExpansiveFactor` | `1.2` | 数组扩容最小系数。同层子节点编码跨度大时应调大 |
| `InitArrayFactor` | `2.5` | 初始数组长度相对于词汇数的系数 |
