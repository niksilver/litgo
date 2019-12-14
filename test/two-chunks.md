# A document with two chunks

Here is chapter one...

``` chapter-one.go
func main() {
    @{Later inclusion}
}
```

# Later inclusion

``` Later inclusion
fmt.Println("Something to be included multiple times")
```

# Chapter two

This is a bit like chapter one...

``` chapter-two.go
func two() {
    @{Later inclusion}
}
```
