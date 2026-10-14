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

# 三、快速开始

```shell
go get gitee.com/ivfzhou/double-array-trie@latest
```

## 3.1 基本用法

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

## 3.2 键值存储

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

## 3.3 持久化

```go
// 序列化到文件
err := d.DumpToFile("./dat.dat.gz")

// 从文件恢复
d, err := dat.ReadFromFile("./dat.dat.gz")
```

# 四、应用场景

`Dat` 的核心能力可以概括为四类：**全词匹配**、**前缀匹配**、**句内分词（`Analysis`）** 以及 **键值存储**。下面结合这几个能力列举典型业务场景。

## 4.1 路由匹配 / API 网关

将 URL 路径作为词库，并绑定对应的处理器，实现 \(O(m)\) 的路由分发；也支持按前缀批量取出某个模块下的所有路由。

```go
d := dat.NewWithValues(map[string]any{
    "/api/user/info":    handleUserInfo,
    "/api/user/login":   handleUserLogin,
    "/api/order/create": handleOrderCreate,
})

// 精确匹配路由并分发
if h, ok := d.MatchGet("/api/user/info"); ok {
    h.(func())() // 调用对应处理器
}

// 前缀匹配：一次性取出 /api/user 下的所有路由（如批量导出、权限统计）
routes := d.HitKeys("/api/user") // ["/api/user/info", "/api/user/login"]
```

## 4.2 敏感词过滤 / 内容审核

词库为敏感词集合，用 `Analysis` 一次性扫描出正文中所有命中词及其位置，再统一替换或拦截，无需逐词尝试匹配。

```go
d := dat.New([]string{"违禁词A", "违禁词B", "诈骗"})

text := "这是一段包含诈骗与违禁词A的文本"
keys, indexes := d.Analysis(text)
for i, k := range keys {
    // k 为命中词，indexes[i] 为该词在原文中的字节位置
    fmt.Println(k, indexes[i])
}
```

## 4.3 中文分词 / 词典切分

把词典中的词作为词库，对句子做全量词典匹配，返回所有命中的词及其位置，可在此基础上实现最大匹配、最小切分等分词策略。

```go
d := dat.New([]string{"自然语言", "语言", "处理", "自然"})
keys, _ := d.Analysis("自然语言处理") // ["自然", "自然语言", "语言", "处理"]
```

## 4.4 搜索联想 / 自动补全

词库为候选词集合，输入前缀实时返回所有以该前缀开头的词，用于搜索框联想、命令补全、代码补全等场景。

```go
d := dat.New([]string{"javascript", "java", "golang", "python"})
suggestions := d.HitKeys("ja") // ["javascript", "java"]
```

## 4.5 黑白名单 / 名单匹配

对 IP、域名、手机号等做精确命中或前缀命中（例如域名后缀、网段判断）。

```go
d := dat.New([]string{"192.168.1.1", "10.0.0.1", "example.com"})

d.Matches("192.168.1.1")    // true
d.MatchPrefix("192.168")    // true（"192.168" 是 "192.168.1.1" 的前缀）
```

## 4.6 词典 / 键值映射

把词作为 key，绑定任意类型的 value，实现极速的「词 → 值」映射，适合 ID 映射、拼音→汉字、词语→词性等查询。

```go
d := dat.NewWithValues(map[string]string{
    "apple":  "苹果",
    "banana": "香蕉",
})
value, _ := d.MatchGet("apple") // "苹果"
```

## 4.7 前缀词组提取

`ObtainPrefixes` 返回「输入串的所有词典前缀」，常用于判断一个字符串由哪些前缀词组构成，或提取其全部前缀词组。

```go
d := dat.New([]string{"ab", "abc", "abg", "jkl"})
d.ObtainPrefixes("abc") // ["ab", "abc"]
```

## 4.8 持久化词库

词库构建成本相对较高，可在离线/启动阶段构建一次后 `DumpToFile` 落盘，服务运行时 `ReadFromFile` 直接加载，避免重复构建。

```go
d, err := dat.ReadFromFile("./dict.dat.gz")
```

# 五、算法原理

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
