# aho-corasick
Efficient string matching in Golang via the aho-corasick algorithm.

x20 faster than https://github.com/cloudflare/ahocorasick and x3 faster than https://github.com/anknown/ahocorasick

This library is heavily inspired by https://github.com/BurntSushi/aho-corasick

## Usage

  ```go
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

It's plenty fast but if you want to use it in parallel, that is also possible.

Memory consumption won't increase because the read-only automaton is not actually copied, only the counters are.

The magic line is `ac := ac`

```go
    builder := NewAhoCorasickBuilder(Opts{
		AsciiCaseInsensitive: true,
		MatchOnlyWholeWords:  true,
		MatchKind:            LeftMostLongestMatch,
	})

	ac := builder.Build(t2.patterns)
	var w sync.WaitGroup

	w.Add(50)
	for i := 0; i < 50; i++ {
		go func() {
			ac := ac
			matches := ac.FindAll(t2.haystack)
			if len(matches) != len(t2.matches) {
				t.Errorf("test %v expected %v matches got %v", 0, len(matches), len(t2.matches))
			}
			for i, m := range matches {
				if m != t2.matches[i] {
					t.Errorf("test %v expected %v matche got %v", i, m, t2.matches[i])
				}
			}
			w.Done()
		}()
	}
	w.Wait()
```
