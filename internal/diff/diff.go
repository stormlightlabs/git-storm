package diff

// EditKind defines the type of diff operation.
type EditKind int

const (
	Equal   EditKind = iota // context: lines unchanged
	Insert                  // added lines ('+')
	Delete                  // removed lines ('-')
	Replace                 // changed lines (shown as Delete + Insert in unified view)
)

func (e EditKind) String() string {
	switch e {
	case Equal:
		return "Equal"
	case Insert:
		return "Insert"
	case Delete:
		return "Delete"
	case Replace:
		return "Replace"
	default:
		return "Unknown"
	}
}

// Edit represents a single edit operation in a diff.
type Edit struct {
	Kind    EditKind // Equal, Insert, or Delete
	AIndex  int      // index in original sequence
	BIndex  int      // index in new sequence
	Content string   // the line or token
}

// Diff represents a generic diffing algorithm.
type Diff interface {
	// Compute computes the edit operations needed to transform a into b.
	Compute(a, b []string) ([]Edit, error)

	// Name returns the human-readable algorithm name (e.g., "LCS", "Hunt–McIlroy").
	Name() string
}

// LCS implements a shortest-edit-script diff algorithm.
//
// For maintainability we use the classic dynamic-programming formulation based on the longest common subsequence.
// While the original Myers paper achieves O(ND) time, this O(NM) variant is simpler and still practical for the small inputs handled by this package.
type LCS struct{}

func (l *LCS) Name() string { return "LCS" }

// Compute computes the shortest edit script using the LCS diff algorithm.
//
// It builds an LCS matrix and walks it to emit the sequence of Equal, Insert, and Delete operations required to transform a into b.
func (l *LCS) Compute(a, b []string) ([]Edit, error) {
	n := len(a)
	lenB := len(b)

	if n == 0 && lenB == 0 {
		return []Edit{}, nil
	}

	if n == 0 {
		edits := make([]Edit, lenB)
		for i := range lenB {
			edits[i] = Edit{Kind: Insert, AIndex: -1, BIndex: i, Content: b[i]}
		}
		return edits, nil
	}

	if lenB == 0 {
		edits := make([]Edit, n)
		for i := range n {
			edits[i] = Edit{Kind: Delete, AIndex: i, BIndex: -1, Content: a[i]}
		}
		return edits, nil
	}

	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, lenB+1)
	}

	for i := n - 1; i >= 0; i-- {
		for j := lenB - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	edits := make([]Edit, 0, n+lenB)

	i, j := 0, 0
	for i < n && j < lenB {
		switch {
		case a[i] == b[j]:
			edits = append(edits, Edit{
				Kind:    Equal,
				AIndex:  i,
				BIndex:  j,
				Content: a[i],
			})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			edits = append(edits, Edit{
				Kind:    Delete,
				AIndex:  i,
				BIndex:  -1,
				Content: a[i],
			})
			i++
		default:
			edits = append(edits, Edit{
				Kind:    Insert,
				AIndex:  -1,
				BIndex:  j,
				Content: b[j],
			})
			j++
		}
	}

	for i < n {
		edits = append(edits, Edit{
			Kind:    Delete,
			AIndex:  i,
			BIndex:  -1,
			Content: a[i],
		})
		i++
	}

	for j < lenB {
		edits = append(edits, Edit{
			Kind:    Insert,
			AIndex:  -1,
			BIndex:  j,
			Content: b[j],
		})
		j++
	}

	return edits, nil
}

// Myers implements the Myers algorithm.
type Myers struct{}

// Name returns algorithm name.
func (m *Myers) Name() string {
	return "Myers"
}

// Compute computes the diff edits needed to transform a into b.
func (m *Myers) Compute(a, b []string) ([]Edit, error) {
	n := len(a)
	mLen := len(b)
	max := n + mLen

	if n == 0 && mLen == 0 {
		return []Edit{}, nil
	}

	if n == 0 {
		edits := make([]Edit, mLen)
		for i := range mLen {
			edits[i] = Edit{Kind: Insert, AIndex: -1, BIndex: i, Content: b[i]}
		}
		return edits, nil
	}

	if mLen == 0 {
		edits := make([]Edit, n)
		for i := range n {
			edits[i] = Edit{Kind: Delete, AIndex: i, BIndex: -1, Content: a[i]}
		}
		return edits, nil
	}

	offset := max
	size := 2*max + 1
	V := make([]int, size)
	trace := make([][]int, max+1)

	if offset+1 < size {
		V[offset+1] = 0
	}
	for D := 0; D <= max; D++ {
		currentV := make([]int, size)
		copy(currentV, V)
		trace[D] = currentV

		for k := -D; k <= D; k += 2 {
			idx := offset + k

			var x int
			if k == -D || (k != D && V[idx-1] < V[idx+1]) {
				x = V[idx+1]
			} else {
				x = V[idx-1] + 1
			}
			y := x - k

			for x < n && y < mLen && a[x] == b[y] {
				x++
				y++
			}

			V[idx] = x

			if x >= n && y >= mLen {
				return m.buildEdits(a, b, trace, D, offset), nil
			}
		}
	}

	return nil, nil
}

// buildEdits reconstructs the edit script from the trace of V arrays.
func (m *Myers) buildEdits(a, b []string, trace [][]int, D, offset int) []Edit {
	var edits []Edit
	x := len(a)
	y := len(b)

	for d := D; d > 0; d-- {
		V := trace[d]
		k := x - y
		idx := offset + k

		var prevK int
		if k == -d || (k != d && V[idx-1] < V[idx+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := V[offset+prevK]
		prevY := prevX - prevK

		var xStart, yStart int
		if prevK == k-1 {
			xStart = prevX + 1
			yStart = prevY
		} else {
			xStart = prevX
			yStart = prevY + 1
		}

		for x > xStart && y > yStart {
			x--
			y--
			edits = append(edits, Edit{
				Kind:    Equal,
				AIndex:  x,
				BIndex:  y,
				Content: a[x],
			})
		}

		if xStart == prevX+1 {
			x--
			edits = append(edits, Edit{
				Kind:    Delete,
				AIndex:  x,
				BIndex:  -1,
				Content: a[x],
			})
		} else {
			y--
			edits = append(edits, Edit{
				Kind:    Insert,
				AIndex:  -1,
				BIndex:  y,
				Content: b[y],
			})
		}

		x = prevX
		y = prevY
	}

	for x > 0 && y > 0 {
		if a[x-1] == b[y-1] {
			x--
			y--
			edits = append(edits, Edit{
				Kind:    Equal,
				AIndex:  x,
				BIndex:  y,
				Content: a[x],
			})
		} else {
			break
		}
	}

	for x > 0 {
		x--
		edits = append(edits, Edit{
			Kind:    Delete,
			AIndex:  x,
			BIndex:  -1,
			Content: a[x],
		})
	}
	for y > 0 {
		y--
		edits = append(edits, Edit{
			Kind:    Insert,
			AIndex:  -1,
			BIndex:  y,
			Content: b[y],
		})
	}

	for i, j := 0, len(edits)-1; i < j; i, j = i+1, j-1 {
		edits[i], edits[j] = edits[j], edits[i]
	}
	return edits
}

// ApplyEdits applies a sequence of edits to reconstruct the target sequence to verify that the diff is correct.
func ApplyEdits(_ []string, edits []Edit) []string {
	result := make([]string, 0)
	for _, edit := range edits {
		switch edit.Kind {
		case Equal, Insert:
			result = append(result, edit.Content)
		case Delete:
			// Skip deleted lines
		}
	}
	return result
}

// CountEditKinds returns a map counting occurrences of each [EditKind].
func CountEditKinds(edits []Edit) map[EditKind]int {
	counts := make(map[EditKind]int)
	for _, edit := range edits {
		counts[edit.Kind]++
	}
	return counts
}
