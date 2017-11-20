# sgf2gif
Generate animated gifs from sgf files.

# Installation
```
$ go get github.com/alcortesm/sgf2gif
```

# Usage Example
Given the file `/tmp/foo.sgf`
(AlphaGo vs. Lee Sedol 2016-03-15 match)
you can generate the corresponding gif file as follows:

```
$ sgf2gif /tmp/foo.sgf /tmp/foo.gif
```

The resulting gif is shown below.

![AlpahGo vs. Lee Sedol 2016-03-15](https://user-images.githubusercontent.com/9169414/33006598-3c0b2106-cdcb-11e7-94d0-d6db14675d71.gif)

# Limitations
Dead stones are not removed from the board.
