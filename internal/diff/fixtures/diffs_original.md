# Text Differencing Algorithms

Text differencing algorithms compute the minimal set of edits required to transform one sequence into another.
They are widely used in version control systems, compilers, and data synchronization tools.

## The Myers Algorithm

Eugene Myers proposed a diff algorithm in 1986 that computes the shortest edit script (SES) between two sequences.
It models the problem as a traversal over a grid, where diagonal moves represent matches and horizontal or vertical moves represent insertions and deletions.

### Key Ideas

- Based on the concept of *edit graph traversal*.
- Uses a dynamic programming approach optimized with linear space.
- Achieves **O(ND)** time complexity where `N` is sequence length and `D` is the edit distance.

### Pseudocode

```text
for D from 0 to MAX:
    for k in range(-D, D+1, 2):
        choose move (insert or delete)
        extend along diagonal as far as possible
        if end reached: return path
```

### Strengths

- Produces minimal diffs.
- Works efficiently for typical text files.
- Used by `git diff`, `diffutils`, and many modern tools.

### Weaknesses

- Complexity increases with extremely long or highly divergent sequences.
- Implementation details are tricky due to path tracing.

## References

- Myers, E. W. (1986). *An O(ND) Difference Algorithm and Its Variations.*
- GNU diffutils documentation.
