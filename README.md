# aho-corasick
efficient string matching in Golang via the aho-corasick algorithm.

This library is heavily inspired by https://github.com/BurntSushi/aho-corasick

## Usage

  ```
    builder := NewAhoCorasickBuilder(Opts{
		asciiCaseInsensitive: true,
		matchOnlyWholeWords:  true,
		matchKind:            LeftMostLongestMatch,
	})

	ac := builder.Build([]string{"bear", "masha"})
	haystack := "The Bear and Masha"
	matches := ac.FindAll(haystack)

	for _, match := range matches {
		println(haystack[match.Start():match.End()])
	}
```
