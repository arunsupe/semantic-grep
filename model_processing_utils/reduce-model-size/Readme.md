# Decreasing the size of the model files

The model files are large (Gigabytes). Each word is typically represented using a 300 dimension, 32 bit floating point vector. While highly accurate, these lareg models are memory intensive and slow. Reducing dimensionality, to 100 or 150 dimensions, can produce smaller, memory efficient, faster, more performant models with minimal (maybe even better) accuracy. 

In `model_processing_utils/reduce-model-size/PCA-dimension-reduction.go`, I have written a small utility to reduce model's dimensions. This will take as input the model file, the output path and optionally, vector dimensions and reduce the model's size using Principal Component Analysis. In my testing, the optimal vector dimensions are somewhere between 100-150. Smaller than that, and accuracy may be compromised. Note: thresholds will be different with the new, smaller, model compared to the large models. Optimize through trial and error.

This can be used to reduce the size of any word2vec binary model used by w2vgrep. Use this like so:

```bash
# build
cd model_processing_utils/reduce-model-size
go build .

# run on large GoogleNews-vectors-negative300-SLIM.bin model (346MB) to make smaller
# GoogleNews-vectors-negative100-SLIM.bin model (117MB)
./reduce-pca -input ../../models/googlenews-slim/GoogleNews-vectors-negative300-SLIM.bin -output ../../models/googlenews-slim/GoogleNews-vectors-negative100-SLIM.bin -dim 100

# use this smaller model in w2vgrep like so
curl -s 'https://gutenberg.ca/ebooks/hemingwaye-oldmanandthesea/hemingwaye-oldmanandthesea-00-t.txt' | bin/w2vgrep.linux.amd64 -n -t 0.5 -m models/googlenews-slim/GoogleNews-vectors-negative100-SLIM.bin --line-number death
```

Please try this if performance is a bottle neck.