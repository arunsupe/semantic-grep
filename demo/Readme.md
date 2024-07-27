# Demo of Semantic Grep (sgrep)

Overview:
This demo shows how the semantic grep program works by finding similar text segments based on different similarity thresholds.

How It Works:
1. **Similarity Scoring**: 
   - The program uses cosine similarity to score word similarity on a scale of 0 to 1.
   - A score of 1 means the words are very similar.

2. **Threshold Adjustment**: 
   - Lowering the similarity_threshold allows the program to match less similar words.

3. **Test Passage**: 
   - We created a passage (test-passage.txt) with various topics and some semantic overlap.
   - With a high threshold, we expect sepcific matches. As we relax the threshold, we expect loss of specificity.

4. **Execution**:
    ```bash
    cat demo/test-passage.txt | bin/sgrep --query="diagnosis technology" --window=40 --similarity_threshold=0.3
    ```

5. **Results**:
   - The outputs from `sgrep` are shown below, including the matched text segments and their calculated similarity scores.
   - You can adjust the similarity threshold using the `similarity_threshold` command line flag to control which matches are displayed. Choosing the correct threshold is a matter of trial and error. 

High similarity matches:
```
Similarity: 0.6481
1. Technology and Innovation: "Artificial intelligence and machine learning are revolutionizing industries. Robotics, automation, and data science are driving innovation in fields ranging from healthcare to finance. The integration of IoT devices and the advancement of quantum computing are opening
...

Similarity: 0.6148
Advances in medical research are providing new treatments for chronic diseases. Preventive care and early diagnosis are key to improving health outcomes. Mindfulness and stress management techniques are gaining popularity." 4. Literature and Art: "Literature and art are reflections of
```

Low similarity matches:
```
Similarity: 0.3156
a pressing global issue. Renewable energy sources like solar and wind power are being developed to reduce carbon emissions. Conservation efforts are crucial to protect endangered species and preserve natural habitats. Sustainable practices in agriculture and industry are essential for

Similarity: 0.3249
art are reflections of human experience. Classic novels and contemporary fiction explore themes of love, loss, and identity. Visual arts, from painting to digital media, offer diverse perspectives and emotional depth. The evolution of artistic expression continues to shape cultural
```


This demo highlights how sgrep identifies semantic connections in text and how adjusting the similarity threshold affects the results.

