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
	Kind       EditKind // Equal, Insert, Delete, or Replace
	AIndex     int      // index in original sequence (-1 for Insert-only)
	BIndex     int      // index in new sequence (-1 for Delete-only)
	Content    string   // the line or token (old content for Replace)
	NewContent string   // new content (only used for Replace operations)
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

// MergeReplacements merges Delete+Insert pairs into Replace operations for better side-by-side rendering.
//
// This function identifies blocks of Delete and Insert operations and pairs them up based on similarity.
// When a Delete and Insert represent the same logical line being modified (e.g., version bump),
// they are merged into a Replace operation that can be rendered on a single line.
//
// The function uses a similarity heuristic to determine if a Delete and Insert pair should be merged:
// - They must share a common prefix of at least 70% of the shorter line's length
// - This prevents merging unrelated changes (e.g., different package names)
//
// The algorithm processes edits in windows, looking ahead up to 10 positions to find matching pairs.
func MergeReplacements(edits []Edit) []Edit {
	if len(edits) <= 1 {
		return edits
	}

	type mergeInfo struct {
		partnIndex int // index of the partner edit
		isDelete   bool
	}

	merged := make(map[int]mergeInfo)
	const lookAheadWindow = 50

	for i := range edits {
		if _, exists := merged[i]; exists || edits[i].Kind != Delete {
			continue
		}

		found := false
		for j := i + 1; j < len(edits) && j < i+lookAheadWindow; j++ {
			if _, exists := merged[j]; exists || edits[j].Kind != Insert {
				continue
			}

			if areSimilarLines(edits[i].Content, edits[j].Content) {
				merged[i] = mergeInfo{partnIndex: j, isDelete: true}
				merged[j] = mergeInfo{partnIndex: i, isDelete: false}
				found = true
				break
			}
		}

		if !found {
			for j := i - 1; j >= 0 && j >= i-lookAheadWindow; j-- {
				if _, exists := merged[j]; exists || edits[j].Kind != Insert {
					continue
				}

				if areSimilarLines(edits[i].Content, edits[j].Content) {
					merged[i] = mergeInfo{partnIndex: j, isDelete: true}
					merged[j] = mergeInfo{partnIndex: i, isDelete: false}
					break
				}
			}
		}
	}

	for i := 0; i < len(edits); i++ {
		if _, exists := merged[i]; exists || edits[i].Kind != Insert {
			continue
		}

		for j := max(0, i-lookAheadWindow); j < i; j++ {
			if _, exists := merged[j]; exists || edits[j].Kind != Delete {
				continue
			}

			if areSimilarLines(edits[j].Content, edits[i].Content) {
				merged[j] = mergeInfo{partnIndex: i, isDelete: true}
				merged[i] = mergeInfo{partnIndex: j, isDelete: false}
				break
			}
		}
	}

	type outputEdit struct {
		edit         Edit
		origPosition int
	}

	outputs := make([]outputEdit, 0, len(edits))

	for i := range edits {
		info, isMerged := merged[i]
		if !isMerged {
			outputs = append(outputs, outputEdit{
				edit:         edits[i],
				origPosition: i,
			})
		} else if info.isDelete {
			outputs = append(outputs, outputEdit{
				edit: Edit{
					Kind:       Replace,
					AIndex:     edits[i].AIndex,
					BIndex:     edits[info.partnIndex].BIndex,
					Content:    edits[i].Content,
					NewContent: edits[info.partnIndex].Content,
				},
				origPosition: i,
			})
		}
	}

	for i := 0; i < len(outputs); i++ {
		for j := i + 1; j < len(outputs); j++ {
			ei := outputs[i].edit
			ej := outputs[j].edit

			keyI := ei.BIndex
			if keyI == -1 {
				keyI = ei.AIndex
			}

			keyJ := ej.BIndex
			if keyJ == -1 {
				keyJ = ej.AIndex
			}

			if keyI > keyJ {
				outputs[i], outputs[j] = outputs[j], outputs[i]
			}
		}
	}

	result := make([]Edit, 0, len(outputs))
	for _, out := range outputs {
		result = append(result, out.edit)
	}

	return result
}

// areSimilarLines determines if two lines are similar enough to be considered a replacement.
//
// Uses a two-phase similarity check:
// 1. Common prefix must be at least 70% of the shorter line
// 2. Remaining suffixes must be at least 60% similar (Levenshtein-like check)
func areSimilarLines(a, b string) bool {
	if a == b {
		return true
	}

	minLen := min(len(b), len(a))

	if minLen == 0 {
		return false
	}

	commonPrefix := 0
	for i := 0; i < minLen; i++ {
		if a[i] == b[i] {
			commonPrefix++
		} else {
			break
		}
	}

	prefixThreshold := float64(minLen) * 0.7
	if float64(commonPrefix) < prefixThreshold {
		return false
	}

	suffixA := a[commonPrefix:]
	suffixB := b[commonPrefix:]

	suffixLenA := len(suffixA)
	suffixLenB := len(suffixB)

	if suffixLenA == 0 && suffixLenB == 0 {
		return true
	}

	lenDiff := suffixLenA - suffixLenB
	if lenDiff < 0 {
		lenDiff = -lenDiff
	}

	maxSuffixLen := max(suffixLenB, suffixLenA)

	if maxSuffixLen > 0 && float64(lenDiff)/float64(maxSuffixLen) > 0.3 {
		return false
	}

	return true
}
