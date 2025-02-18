package detect

// Location represents a location in a file
type Location struct {
	startLine      int
	endLine        int
	startColumn    int
	endColumn      int
	startLineIndex int
	endLineIndex   int
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func location(fragment Fragment, matchIndex []int) Location {
	var (
		prevNewLine       int
		location          Location
		lineSet           bool
		_lineNum          int
		_newLineByteIndex int
	)

	start := matchIndex[0]
	end := matchIndex[1] - 1

	// If the last character of the fragment is not a newline, then
	// we should include the last character in the match.
	if fragment.Raw[end] != '\n' {
		end = end + 1
	}

	// default startLineIndex to 0
	location.startLineIndex = 0

	// Fixes: https://github.com/zricethezav/gitleaks/issues/1037
	// When a fragment does NOT have any newlines, a default "newline"
	// will be counted to make the subsequent location calculation logic work
	// for fragments will no newlines.
	if len(fragment.newlineIndices) == 0 {
		fragment.newlineIndices = [][]int{
			{len(fragment.Raw), len(fragment.Raw) + 1},
		}
	}

	for lineNum, pair := range fragment.newlineIndices {
		_lineNum = lineNum
		newLineByteIndex := pair[0]
		_newLineByteIndex = newLineByteIndex
		if prevNewLine <= start && start < newLineByteIndex {
			lineSet = true
			location.startLine = lineNum
			location.endLine = lineNum
			location.startColumn = max(1, (start - prevNewLine))
			location.startLineIndex = prevNewLine
			location.endLineIndex = newLineByteIndex
		}
		if prevNewLine < end && end <= newLineByteIndex {
			location.endLine = lineNum
			location.endColumn = (end - prevNewLine)
			location.endLineIndex = newLineByteIndex
		}
		prevNewLine = pair[0]
	}

	// If the end of the match is on the last line of the fragment
	// and the end column has not been set, then set it.
	if end > _newLineByteIndex && location.endColumn == 0 {
		location.endColumn = end
		location.endLine = _lineNum + 1
	}

	if !lineSet {
		// if lines never get set then that means the secret is most likely
		// on the last line of the diff output and the diff output does not have
		// a newline
		location.startColumn = max(1, (start - prevNewLine))
		location.endColumn = (end - prevNewLine)
		location.startLine = _lineNum + 1
		location.endLine = _lineNum + 1

		// search for new line byte index
		i := 0
		for end+i < len(fragment.Raw) {
			if fragment.Raw[end+i] == '\n' {
				break
			}
			if fragment.Raw[end+i] == '\r' {
				break
			}
			i++
		}
		location.endLineIndex = end + i
	}

	return location
}
