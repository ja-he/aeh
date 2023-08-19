# äh...

Just a simple front for calling out to LLM APIs, right now only the OpenAI-API.

## Scope

Just for my personal use.
I like [`mods`](https://github.com/charmbracelet/mods) for some things but I found it a touch intransparent in which model it actually uses and wanted something super-simple that basically just stripped the JSON-cruft around the actual answer in a not-completely-shabby way.

## Building / Installing

It's Go, so you need `go`.

```
go install github.com/ja-he/aeh
```

This will unfortunately (or fortunately) not give you the full non-ASCII binary name experience, so you may also want to:

```
git clone https://github.com/ja-he/aeh äh
cd äh
go install äh.go
```

## Usage

For example:
```
$ äh "A Haiku on linear programming"
remaining requests: 199
remaining tokens: 9975
Optimizing lines,
Boundaries clearly defined,
Solutions we find.
model: gpt-4-0613
total tokens used: 28
```

You can get rid of those extra lines by simply redirecting STDERR:
```
$ äh "A limerick about Langton's ant" 2> /dev/null
Langton's ant, quite a curious sight
Moves in a grid, left or right
Black to white, white to black
Never does it look back
In its algorithmic, infinite flight.
```

For more, see `äh -h`.

## Why not call it 'hmm' so it could be typed on more Keyboards?

äh...
