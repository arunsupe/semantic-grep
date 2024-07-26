# sgrep - Semantic Grep
sgrep is a command-line tool that performs semantic searches on text input using word embeddings. It's designed to find semantically similar content based on a given query, going beyond simple string matching.


## Features
- Semantic search using Word2Vec embeddings
- Configurable similarity threshold
- Adjustable window size for context
- Color-coded output based on similarity score
- Supports configuration via JSON file and command-line arguments


## Requirements
- A Word2Vec model in binary format. 
- download-model.sh is a simple helper script that will download the small word2vec model hosted by eyaler and save it in `models/googlenews-slim/` directory
- Alternatively, you can download and unzip the .bin file locally and update the config.json. 

    - Google's Word2Vec: from https://github.com/mmihaltz/word2vec-GoogleNews-vectors
    - A slim version: GoogleNews-vectors-negative300-SLIM.bin.gz model from https://github.com/eyaler/word2vec-slim/

Note: There are no external dependenceis; uses just the stdlib (and the model)


## Installation
- clone the repo
- run `go build -o bin/sgrep` 


## Usage
```bash
 curl -s 'https://gutenberg.ca/ebooks/hemingwaye-oldmanandthesea/hemingwaye-oldmanandthesea-00-t.txt' | bin/sgrep --similarity_threshold=0.50 --window=100 --query='promised fish' 
 ```
- run `bin/sgrep` to see commandline flags

## Configuration
sgrep can be configured using a JSON file. By default, it looks for `config.json` in the current directory. You can specify a different configuration file using the `-config` flag.

Example `config.json`:

```json
{
    "model_path": "models/googlenews-slim/GoogleNews-vectors-negative300-SLIM.bin",
    "similarity_threshold": 0.3,
    "window": 50
}
```


## Output
The output includes:

    The similarity score for each matching segment
    The matching text, color-coded based on similarity (red for lowest, magenta for highest)


## License
This project is distributed under the MIT License (refer to LICENSE file for details).


## Disclaimer
The provided model might not capture all semantic nuances and may require adjustments based on your specific use case. Consider exploring other models or training your own model for better accuracy.