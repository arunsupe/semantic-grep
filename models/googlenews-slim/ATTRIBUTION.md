# Word2Vec Slim Model Attribution

This directory contains a mirrored version of the word2vec-slim model, which is derived from the Google News dataset (about 100 billion words).

Original source: https://github.com/eyaler/word2vec-slim/

The word2vec model used here is a slimmed-down version of the original Google News model, created by Eyal Gruss.

## Original Attribution

Pre-trained vectors trained on part of Google News dataset (about 100 billion words).
Model contains 300-dimensional vectors for 3 million words and phrases.

The original model was created by Mikolov et al. and is available here:
https://code.google.com/archive/p/word2vec/

## License

This model is distributed under the Apache License 2.0. See the LICENSE file in this directory for the full license text.

## Changes Made

Quantized 8 bit int models are derived from the original 32 bit float model.