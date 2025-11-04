# Text Differencing Algorithms

Diff algorithms determine the smallest set of operations to make two sequences identical.
They are essential to tools like `git`, `rsync`, and file synchronization systems.

## The Hunt–McIlroy Algorithm

Developed by James W. Hunt and M. Douglas McIlroy in 1976, this algorithm underlies the original Unix `diff` utility.
Unlike Myers, it relies on finding **longest common subsequences (LCS)** to compute differences.

### Core Principles

- Operates on the *longest common subsequence* problem.
- Identifies matching lines using hash-based comparison.
- Produces intuitive, human-readable diffs.

### Simplified Outline

```text
match = longest_common_subsequence(A, B)
for each segment not in match:
    emit insertion or deletion
```

### Advantages

- Generates results similar to human intuition.
- Performs well on structured text like source code.
- Simple to implement and debug.

### Limitations

- May not always yield the shortest possible edit script.
- Space complexity can grow for large inputs.

## Comparison to Myers

| Feature    | Myers             | Hunt–McIlroy       |
| ---------- | ----------------- | ------------------ |
| Complexity | O(ND)             | O(N log N) typical |
| Output     | Minimal           | Readable           |
| Origin     | 1986              | 1976               |
| Use Cases  | Modern diff tools | Unix `diff`        |

## References

- Hunt, J. W. & McIlroy, M. D. (1976). *An Algorithm for Differential File Comparison.*
- Research on Longest Common Subsequence algorithms.
