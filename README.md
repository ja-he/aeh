# äh...

Just a simple front for calling out to Large Language Model (LLM) APIs, currently supporting only the OpenAI API.

## Scope

Just for my personal use.
I like [`mods`](https://github.com/charmbracelet/mods) for some things but I found it a touch intransparent in which model it actually uses and wanted something super-simple that basically just stripped the JSON-cruft around the actual answer in a not-completely-shabby way.

## Features

- Query OpenAI's GPT models via the command line.
- Supports setting configurations via a YAML file.
- Customizable model and temperature settings.
- Records prompt and response history.
- Graceful handling of `SIGINT` and `SIGTERM` signals.

## Building / Installing

It's Go, so you need `go`.

### Quick Install

```sh
go install github.com/ja-he/aeh
```

This will unfortunately (or fortunately) not give you the full non-ASCII binary name experience, so you may also want to:

```sh
git clone https://github.com/ja-he/aeh äh
cd äh
go install äh.go
```

### Manual Build

1. Ensure you have Go installed on your system.
2. Clone the repository or download the code files.
3. Navigate to the directory containing the code.
4. Run the following command to build the utility:

    ```sh
    go build
    ```

## Configuration

The utility uses a configuration file in YAML format.
By default, the configuration file is created in `~/.config/äh/config.yaml` if it does not exist.
The configuration file allows you to set default values for the model and temperature.

Example default configuration:

```yaml
defaults:
  model: gpt-3.5-turbo
  temp: 0.7
```

## Usage

To run the tool, use the following syntax:

```sh
./äh [flags] <prompt>
```

### Flags

- `-m` : Specify the model to use (e.g., `gpt-3.5-turbo`, `gpt-4`).
- `-t` : Specify the temperature to use for the model (a value between 0.0 and 1.0).

### Examples

1. Here is how you do a simple prompt:

    ```sh
    ./äh "A Haiku on linear programming"
    ```

   That will give you something like this:

    ```
    remaining requests: 199
    remaining tokens: 9975
    model: gpt-4-0613
    total tokens used: 28
    Optimizing lines,
    Boundaries clearly defined,
    Solutions we find.
    ```

2. As we all know, LLMs can do some really impressive things that people struggle with, and a key skill is knowing _how_ to use them right.
   Accordingly, you can of course configure the temperature and model you use, for example:

    ```sh
    äh \
        -m 'gpt-4o' \
        -t 1.8 \
        'generate the best UUID. Just do it, no nonsense. I do not want one of those random, common UUIDs, I want a really good one. Answer with just the UUID.'
    ```

    ```
    remaining requests: 4999
    remaining tokens: 599945
    model: gpt-4o-2024-05-13
    total tokens used: 58
    4e495ardenl3anspw394utitkte-Uniseterminate
    ```

    Marvelous. 🤯

3. You can get rid of those extra lines (which are mostly informational cruft from the response headers) by simply redirecting STDERR.
   If something goes wrong with the request though, that might leave you blind to what it was.

    ```sh
    ./äh "A limerick about Langton's ant" 2> /dev/null
    ```

    ```
    Langton's ant, quite a curious sight
    Moves in a grid, left or right
    Black to white, white to black
    Never does it look back
    In its algorithmic, infinite flight.
    ```

For more, see `./äh -h`.

### Environment Variables

- `OPENAI_API_KEY` : Set this environment variable to your OpenAI API key.
- `AEH_CONFIG_DIR` : (Optional) Custom directory for storing configuration and history files.
- `AEH_ERR` : (Optional) Set to `"false"` to disable error messages.
- `AEH_SPIN` : (Optional) Set to `"false"` to disable the spinner animation.

### Error Handling

If the `OPENAI_API_KEY` environment variable is not set, the utility will terminate with an error message.
Ensure to set this environment variable before running the tool.

### History

The tool appends each prompt and response to a history file located in the configuration directory (`~/.config/äh/history.json`).

## License

This project is licensed under the MIT License, see [`LICENSE`](./LICENSE).

## Contributing

Feel free to open issues or submit pull requests, but don't expect too much from me here;
I am quite happy keeping this tool fairly basic (or perhaps rather, I am averse to turning this thing I just made to scratch a small itch I have from time to time into a complicated thing that diverges from my narrow expectations and needs).

Feel free to fork as well, of course, if you have a cool idea that I prefer not to have here 🙂

## Contact

For any questions or feedback, please feel free to open an issue or discussion.

## Why not call it 'hmm' so it could be typed on more keyboards?

äh...
