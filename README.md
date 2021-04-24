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

Replacing of matches in the haystack.

`replaceWith` needs to be the same length as the `patterns`
```go
replaced := ac.ReplaceAll(haystack, replaceWith)
```

`ReplaceAllFunc` is useful, for example, if you want to use the original text cassing but you are matching
case insensitively.
```go
replaced := ac.ReplaceAllFunc(haystack, func(match Match) string {
    return `<a>` + haystack[match.Start():match.End()] + `<\a>`
})
```

Search for matches one at a time via the iterator

```go
iter := ac.Iter(haystack)

for next := iter.Next(); next != nil; next = iter.Next() {
    ...
}
```

It's plenty fast but if you want to use it in parallel, that is also possible.

Memory consumption won't increase because the read-only automaton is not actually copied, only the counters are.

The magic line is `ac := ac`

```go
var w sync.WaitGroup

w.Add(50)
for i := 0; i < 50; i++ {
    go func() {
        ac := ac
        matches := ac.FindAll(haystack)
        println(len(matches))
        w.Done()
    }()
}
w.Wait()
```
