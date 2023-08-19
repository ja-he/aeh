# äh...

Just a simple front for calling out to LLM APIs, right now only the OpenAI-API.

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

See `go run äh.go -h`.
