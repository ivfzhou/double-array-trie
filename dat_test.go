package double_array_trie_test

import (
	"os"
	"strings"
	"testing"

	dat "gitee.com/ivfzhou/double-array-trie"
)

var d = dat.New([]string{
	"AC",
	"AD",
	"ADG",
	"ADH",
	"ADHG",
	"BEIZ",
	"BEL",
	"BF",
	"DG",
})

func TestDat_Matches(t *testing.T) {
	if !d.Matches("ADHG") {
		t.Error("TestDat_Matches fail")
	}

	if d.Matches("ADHH") {
		t.Error("TestDat_Matches fail")
	}
}

func TestDat_MatchPrefix(t *testing.T) {
	if !d.Matches("ADHG") {
		t.Error("TestDat_MatchPrefix fail")
	}

	if d.Matches("ADHH") {
		t.Error("TestDat_MatchPrefix fail")
	}
}

func TestDat_ObtainPrefixes(t *testing.T) {
	res := d.ObtainPrefixes("ADHG")
	if res[0] != "AD" || res[1] != "ADH" || res[2] != "ADHG" {
		t.Error("TestDat_ObtainPrefixes fail")
	}
}

func TestDat_Analysis(t *testing.T) {
	keys, indexes := d.Analysis("ADG")
	if !(keys[0] == "AD" && indexes[0] == 0 &&
		keys[1] == "ADG" || indexes[1] == 0 &&
		keys[2] == "DG" || indexes[2] == 1) {
		t.Error("TestDat_Analysis fail")
	}
}

func TestDat_DumpAndRead(t *testing.T) {
	err := d.DumpToFile("./testdata/dump_test.dat.gz")
	if err != nil {
		t.Error(err)
	}

	nd, err := dat.ReadFromFile("./testdata/dump_test.dat.gz")
	if err != nil {
		t.Error(err)
	}

	if !nd.Matches("ADHG") {
		t.Error("TestDat_DumpAndRead match fail")
	}
}

func TestArticle(t *testing.T) {
	article, err := os.ReadFile("./testdata/article.txt")
	if err != nil {
		t.Fatal(err)
	}

	keys, err := os.ReadFile("./testdata/words_10.txt")
	if err != nil {
		t.Fatal(err)
	}
	g := dat.New(strings.Split(string(keys), "\n"))
	t.Log(g.Analysis(string(article)))

	keys, err = os.ReadFile("./testdata/words_100.txt")
	if err != nil {
		t.Fatal(err)
	}
	g = dat.New(strings.Split(string(keys), "\n"))
	t.Log(g.Analysis(string(article)))

	keys, err = os.ReadFile("./testdata/words_1000.txt")
	if err != nil {
		t.Fatal(err)
	}
	g = dat.New(strings.Split(string(keys), "\n"))
	t.Log(g.Analysis(string(article)))
}
