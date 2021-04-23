# aho-corasick
Efficient string matching in Golang via the aho-corasick algorithm.

x20 faster than https://github.com/cloudflare/ahocorasick and x3 faster than https://github.com/anknown/ahocorasick

This library is heavily inspired by https://github.com/BurntSushi/aho-corasick

## Usage

  ```
    builder := NewAhoCorasickBuilder(Opts{
	    AsciiCaseInsensitive: true,
        MatchOnlyWholeWords:  true,
        MatchKind:            LeftMostLongestMatch,
	})

	ac := builder.Build([]string{"bear", "masha"})
	haystack := "The Bear and Masha"
	matches := ac.FindAll(haystack)

	for _, match := range matches {
		println(haystack[match.Start():match.End()])
	}
```
