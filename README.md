# 1. 说明

双数组检索树，固定词汇地高速地内存占用少地检索数据结构。

dat 可用于 httprule 的前缀匹配情景，或者其它固定词汇的检索情形。也可用于分词检索，其性能远优于二叉树。

# 2. 使用

```shell
go get gitee.com/ivfzhou/double-array-trie@latest
```

```golang
import dat "gitee.com/ivfzhou/double-array-trie"

// 需要被检索的数据
data := []string{
"/api/user/info",
"/api/user/register",
"/api/user/login"
}

// 构建双数组
d := dat.New(data)

// 判断是否存在与数据中
path := request.URL.Path
b := d.Matches(path)

// 判断数据中是否有前缀和 path 匹配
b := d.MatchPrefix(path)

// 获取 path 在字典中的索引
i := d.MatchesIndex(path)

// 返回所有匹配前缀 path 的数据
d.ObtainPrefixes(path)

// 解析 sentence，返回数据中出现在 sentence 中的数据和其位置
d.Analysis("sentence")

```

#### 状态转移方程：

    firstState = 1
    check[code + state -2] = state 
    base[code + state -2] = nextState

终节点：

    check[code + state -2] = state 
    base[code + state -2] = nextState
    base[nextState - 2] < 0 or check[nextState -2] = nextState

### 例子

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

    双数组值：
    check [1 1  4 1  3 3  8  9  10 11 7  7   14 8 8  10  18 19 20 12 12 16 13]
    base  [3 7 -1 16 4 8 -2 -3 -4 -5  12 19 -6  9 10 11 -7 -8 -9  13 18 20 14]
    code A=1 B=2 C=3 D=4 E=5 F=6 G=7 H=8 I=9 L=10 Z=11

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

# 3. 联系作者

电邮：ivfzhou@126.com
