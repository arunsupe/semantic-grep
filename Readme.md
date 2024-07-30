# sgrep - Semantic Grep

sgrep is a command-line tool that performs semantic searches on text input using word embeddings. It's designed to find semantically similar matches to the query, going beyond simple string matching. Supports word2vec and glove models. The experience is designed to be similar to grep. 

## Example Usage

Search for words similar to "death" in Hemingway's "The Old Man and the Sea" with context and line numbers:

```bash
curl -s 'https://gutenberg.ca/ebooks/hemingwaye-oldmanandthesea/hemingwaye-oldmanandthesea-00-t.txt' \
    | sgrep -C 2 -n -threshold 0.55 death
```

Output:
![alt text](demo/image.png)

This command:

    - Fetches the text of "The Old Man and the Sea" from Project Gutenberg Canada
    - Pipes the text to sgrep
    - Searches for words semantically similar to "death"
    - Uses a similarity threshold of 0.55 (-threshold 0.55)
    - Displays 2 lines of context before and after each match (-C 2)
    - Shows line numbers (-n)

The output will show matches with their similarity scores, highlighted words, context, and line numbers.

## Features

- Semantic search using word embeddings (word2vec and glove support out of the box)
- Configurable similarity threshold
- Context display (before and after matching lines)
- Color-coded output for matched words and line numbers
- Support for reading from files or standard input
- Configurable via JSON file and command-line arguments

## Installation

Two files are absolutely needed: 
1. the sgrep binary
2. the vector embedding model file
3. (Optionally, a config.json file to tell sgrep where the embedding model is)

**Using install script**:

```bash
# clone
git clone https://github.com/arunsupe/semantic-grep.git
cd semantic-grep

# run install:
#   compiles using the local go compiler, installs in user/bin, 
#   downloads the model to $HOME/.config/semantic-grep
#   makes config.json
bash install.sh
``` 
**Binary**:

1. Download the latest binary release
2. Download a vector embedding model (see below)
3. Optionally, download the config.json to configure model location there (or do this from the command line)

**From source (linux/osx)**:

```bash
# clone
git clone https://github.com/arunsupe/semantic-grep.git
cd semantic-grep

# build
go build -o sgrep

# download a word2vec model using this helper script (see "Word Embedding Model" below)
bash download-model.sh
```

## Usage

Basic usage:

./sgrep [options] <query> [file]

If no file is specified, sgrep reads from standard input.

### Command-line Options

- `-model_path`: Path to the Word2Vec model file ('models/glove/glove.6B.300d.bin'). Overrides config file.
- `-threshold`: Similarity threshold for matching (default: 0.7)
- `-A`: Number of lines to display after a match
- `-B`: Number of lines to display before a match
- `-C`: Number of lines to display before and after a match
- `-n`: Print line numbers

## Configuration

- sgrep can be configured using a JSON file. By default, it looks for `config.json` in the current directory.

Example `config.json`:

```json
{
    "model_path": "models/glove/glove.6B.300d.bin"
}
```

## Word Embedding Model
`sgrep` requires a word embedding model in __binary__ format. The default model loader expects vectors of dimension 300, as 32 bit floats or as 8 bit ints (quantized). It uses the model file's extension to determine the type (.bin, .8bit.int). Compatible model files are provided in this repo ([models/](models/)). Download one of the .bin files from the `models/` directory and update the path in config.json.

Alternatively, you can use pre-trained models (like Google's Word2Vec) or train your own using tools like gensim. Note though that there does not seem to be a standardized binary format (google's is different to facebook's fasttext or gensim's default _save()_). For `sgrep`, because efficiently loading the large model is key for performance, I have elected to keep the simplest format. 


## A word about word2vec vs glove
Word2Vec focuses on local context (words occuring close to each other in text) while GloVe leverages global co-occurrence statistics ([explaination](https://machinelearninginterview.com/topics/natural-language-processing/what-is-the-difference-between-word2vec-and-glove/)). For practical purposes, they seem equivalent.


## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.


## License and attribution:
The code in this project is licensed under the MIT [License](LICENSE). 

**Word2Vec Model**:

This project uses a mirrored version of the word2vec-slim model, which is stored in the `models/googlenews-slim` directory. This model is distributed under the Apache License 2.0. For more information about the model, its original authors, and the license, please see the `models/googlenews-slim/ATTRIBUTION.md` file.

**GloVe word vectors**:

This project uses a processed version of the GloVe word vectors, which is stored in the `models/glove` directory. This work is distributed under the Public Domain Dedication and License v1.0. For more information about the model, its original authors, and the license, please see the `models/glove/ATTRIBUTION.md` file.

**Fasttext word vectors**:

This project uses a processed version of the fasttext word vectors, which is stored in the `models/fasttext` directory. This work is distributed under the Creative Commons Attribution-Share-Alike License 3.0. For more information about the model, its original authors, and the license, please see the `models/fasttext/ATTRIBUTION.md` file.


## Sources of models in the web
- Google's Word2Vec: from https://github.com/mmihaltz/word2vec-GoogleNews-vectors
- A slim version of the above: GoogleNews-vectors-negative300-SLIM.bin.gz model from https://github.com/eyaler/word2vec-slim/
- Stanford NLP group's Global Vectors for Word Representation (glove) model [source](https://nlp.stanford.edu/projects/glove/): binary version is in mirrored in [models/glove/](models/glove/).  