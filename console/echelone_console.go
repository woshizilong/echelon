package console

import (
	"bufio"
	"github.com/cirruslabs/echelon/node"
	"os"
	"strings"
	"time"
)

type EchelonConsole struct {
	output            *bufio.Writer
	root              *node.EchelonNode
	currentFrameLines []string
	refreshRate       time.Duration
	renderRoot        bool
}

func NewConsole(output *os.File, root *node.EchelonNode, renderRoot bool) *EchelonConsole {
	return &EchelonConsole{
		output:      bufio.NewWriter(output),
		root:        root,
		renderRoot:  renderRoot,
		refreshRate: 200 * time.Millisecond,
	}
}

func (console *EchelonConsole) StartDrawing() {
	for console.root != nil && console.root.IsRunning() {
		console.renderFrame()
		time.Sleep(console.refreshRate)
	}
	console.renderFrame()
}

func (console *EchelonConsole) renderFrame() {
	var newFrameLines []string
	if console.renderRoot {
		newFrameLines = console.root.Render()
	} else {
		newFrameLines = console.root.RenderChildren()
	}
	calculateIncrementalUpdate(console.output, console.currentFrameLines, newFrameLines)
	console.currentFrameLines = newFrameLines
}

func calculateIncrementalUpdate(output *bufio.Writer, linesBefore []string, linesAfter []string) {
	const moveUp = "\u001B[A"
	const moveDown = "\u001B[B"
	const moveBeginningOfLine = "\r"
	const eraseLine = "\u001B[K" // move to the beginning and erase
	const savePosition = "\u001B[s"
	const restorePosition = "\u001B[u"
	commonElements := commonElementsCount(linesBefore, linesAfter)
	if commonElements > 0 {
		calculateIncrementalUpdate(output, linesBefore[commonElements:], linesAfter[commonElements:])
		return
	}
	linesBeforeCount := len(linesBefore)
	linesAfterCount := len(linesAfter)
	if linesBeforeCount > linesAfterCount {
		// there will be less lines so let's clear some
		output.WriteString(strings.Repeat(moveUp+eraseLine, linesBeforeCount-linesAfterCount))
		output.WriteString(savePosition)
		// and move up for the rest
		output.WriteString(strings.Repeat(moveUp, linesAfterCount))
	} else {
		output.WriteString(savePosition)
		output.WriteString(strings.Repeat(moveUp, linesBeforeCount))
	}
	for i := 0; i < linesAfterCount; i++ {
		if i < linesBeforeCount {
			// line existed before so let's replace it
			if linesBefore[i] != linesAfter[i] {
				output.WriteString(eraseLine)
				output.WriteString(linesAfter[i])
				output.WriteString(moveBeginningOfLine)
			}
			output.WriteString(moveDown)
		} else {
			output.WriteString(linesAfter[i])
			output.WriteString("\n")
		}
	}
	output.WriteString(restorePosition)
	output.Flush()
}

func commonElementsCount(one []string, two []string) int {
	oneCount := len(one)
	twoCount := len(two)
	minCount := oneCount
	if twoCount < minCount {
		minCount = twoCount
	}
	for i := 0; i < minCount; i++ {
		if one[i] != two[i] {
			return i
		}
	}
	return minCount
}
