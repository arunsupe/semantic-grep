# Interesting/helpful utilities to manage word embedding models

`cluster.go`
    A program to collect similar words into a text file, with one cluster per line. It optionally takes nunber of clusters to find as input

`synonym-finder.go`
    A program to find all words in the model above a similarity threshold to the qurey word. Essentially, finds synonyms

`fasttext-to-bin.go`
    A utility to convert FastText text model files to Word2Vec binary format for use with w2vgrep.

    