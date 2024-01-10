package gui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/null93/waterfall/sdk/aws"
)

type View string

type State struct {
	screen            tcell.Screen
	dataSet           *aws.DataSet
	CurrentView       View
	SelectedStack     string
	SelectedOperation string
	selectedIndex     int
	AllStacks         bool
	AllOperations     bool
	LastRefreshed     time.Time
}

const (
	VIEW_WATERFALL  View = "waterfall"
	VIEW_HELP       View = "help"
	VIEW_STACKS     View = "stacks"
	VIEW_OPERATIONS View = "operations"
	VIEW_DETAILS    View = "details"
)

var (
	DefaultStyle     = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	HighlightedStyle = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
)

func NewState(screen tcell.Screen, dataSet *aws.DataSet) *State {
	screen.SetStyle(DefaultStyle)
	return &State{
		screen:            screen,
		dataSet:           dataSet,
		selectedIndex:     0,
		CurrentView:       VIEW_WATERFALL,
		SelectedStack:     dataSet.OriginalStackArn,
		SelectedOperation: "",
		AllStacks:         false,
		AllOperations:     false,
		LastRefreshed:     time.Now(),
	}
}

func (s *State) ResetSelectedIndex() {
	s.selectedIndex = 0
}

func (s *State) IncrementSelected() {
	total := len(s.dataSet.GetSortedIntervals(s.SelectedStack, s.SelectedOperation, s.AllStacks, s.AllOperations))
	if s.selectedIndex < total-1 {
		s.selectedIndex = s.selectedIndex + 1
	} else {
		s.selectedIndex = 0
	}
}

func (s *State) DecrementSelected() {
	total := len(s.dataSet.GetSortedIntervals(s.SelectedStack, s.SelectedOperation, s.AllStacks, s.AllOperations))
	if s.selectedIndex > 0 {
		s.selectedIndex = s.selectedIndex - 1
	} else {
		s.selectedIndex = total - 1
	}
}

func (s *State) IncrementOperationSelected() {
	events := s.dataSet.GetOperations(s.SelectedStack, s.AllStacks)
	if len(events) > 1 {
		for i, event := range events {
			if event.EventId == s.SelectedOperation {
				if i < len(events)-1 {
					s.SelectedOperation = events[i+1].EventId
				} else {
					s.SelectedOperation = events[0].EventId
				}
				break
			}
		}
	}
}

func (s *State) DecrementOperationSelected() {
	events := s.dataSet.GetOperations(s.SelectedStack, s.AllStacks)
	if len(events) > 1 {
		for i, event := range events {
			if event.EventId == s.SelectedOperation {
				if i > 0 {
					s.SelectedOperation = events[i-1].EventId
				} else {
					s.SelectedOperation = events[len(events)-1].EventId
				}
				break
			}
		}
	}
}

func (s *State) IncrementSelectedStack() {
	stackArns := s.dataSet.GetStackArns()
	if len(stackArns) > 1 {
		for i, stackArn := range stackArns {
			if stackArn == s.SelectedStack {
				if i < len(stackArns)-1 {
					s.SelectedStack = stackArns[i+1]
				} else {
					s.SelectedStack = stackArns[0]
				}
				break
			}
		}
	}
}

func (s *State) DecrementSelectedStack() {
	stackArns := s.dataSet.GetStackArns()
	if len(stackArns) > 1 {
		for i, stackArn := range stackArns {
			if stackArn == s.SelectedStack {
				if i > 0 {
					s.SelectedStack = stackArns[i-1]
				} else {
					s.SelectedStack = stackArns[len(stackArns)-1]
				}
				break
			}
		}
	}
}

func (s *State) Render() {
	width, _ := s.screen.Size()
	s.screen.Clear()
	s.renderTopBar()
	row := 14
	switch s.CurrentView {
	case VIEW_WATERFALL:
		s.drawText(row, 3, width, DefaultStyle, "STACK", nil)
		s.drawText(row, 57, width, DefaultStyle, "INTERVAL", nil)
		s.renderWaterfall(row + 1)
	case VIEW_HELP:
		s.renderLegend(row + 1)
	case VIEW_STACKS:
		s.renderStacks(row + 1)
	case VIEW_OPERATIONS:
		s.renderOperation(row + 1)
	case VIEW_DETAILS:
		s.renderDetails(row + 1)
	}
	s.screen.Show()
}

func (s *State) renderTopBar() {
	width, _ := s.screen.Size()

	switch s.CurrentView {
	case VIEW_WATERFALL:
		s.drawText(0, 0, width, DefaultStyle, "Quit: <Esc>, Help: h, Stacks: s, Operations: o, Details: <Enter>", nil)
	case VIEW_HELP:
		s.drawText(0, 0, width, DefaultStyle, "Quit: <Esc>, Timeline: <Enter>, Stacks: s, Operations: o", nil)
	case VIEW_STACKS:
		s.drawText(0, 0, width, DefaultStyle, "Quit: <Esc>, Timeline: <Enter>, Help: h, Operations: o", nil)
	case VIEW_OPERATIONS:
		s.drawText(0, 0, width, DefaultStyle, "Quit: <Esc>, Timeline: <Enter>, Help: h, Stacks: s", nil)
	case VIEW_DETAILS:
		s.drawText(0, 0, width, DefaultStyle, "Quit: <Esc>, Timeline: <Enter>, Help: h, Stacks: s, Operations: o", nil)
	}

	selectionHelp := "Selection: <Up> or <Down>"
	if !s.AllOperations {
		selectionHelp += ", Cycle Operation: <Left> or <Right>"
	}
	if !s.AllStacks {
		selectionHelp += ", Cycle Stack: <Tab> or <Shift><Tab>"
	}
	s.drawText(1, 0, width, DefaultStyle, selectionHelp, nil)

	allStacksMessage := "All Stacks"
	combineOperationsMessage := "All Operations"
	if s.AllStacks {
		allStacksMessage = "Specific Stack"
	}
	if s.AllOperations {
		combineOperationsMessage = "Specific Operation"
	}
	s.drawText(2, 0, width, DefaultStyle, "Refresh Data: r, "+allStacksMessage+": S, "+combineOperationsMessage+": O", nil)

	intervals := s.dataSet.GetSortedIntervals(s.SelectedStack, s.SelectedOperation, s.AllStacks, s.AllOperations)

	fillerRune := '━'
	activeTabStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	activeTabTextStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)

	s.drawText(3, 0, width, DefaultStyle, "┏━━━━━━━━━━━┓┏━━━━━━┓┏━━━━━━━━┓┏━━━━━━━━━━━━┓┏━━━━━━━━━┓", nil)
	s.drawText(4, 0, width, DefaultStyle, "┃ WATERFALL ┃┃ HELP ┃┃ STACKS ┃┃ OPERATIONS ┃┃ DETAILS ┃", nil)
	s.drawText(5, 0, width, DefaultStyle, "┻━━━━━━━━━━━┻┻━━━━━━┻┻━━━━━━━━┻┻━━━━━━━━━━━━┻┻━━━━━━━━━┻", &fillerRune)

	switch s.CurrentView {
	case VIEW_WATERFALL:
		s.drawText(3, 0, width, activeTabStyle, "▄▄▄▄▄▄▄▄▄▄▄▄▄", nil)
		s.drawText(4, 0, width, activeTabStyle, "█████████████", nil)
		s.drawText(5, 0, width, activeTabStyle, "▀▀▀▀▀▀▀▀▀▀▀▀▀", nil)
		s.drawText(4, 2, width, activeTabTextStyle, "WATERFALL", nil)
	case VIEW_HELP:
		s.drawText(3, 13, width, activeTabStyle, "▄▄▄▄▄▄▄▄", nil)
		s.drawText(4, 13, width, activeTabStyle, "████████", nil)
		s.drawText(5, 13, width, activeTabStyle, "▀▀▀▀▀▀▀▀", nil)
		s.drawText(4, 15, width, activeTabTextStyle, "HELP", nil)
	case VIEW_STACKS:
		s.drawText(3, 21, width, activeTabStyle, "▄▄▄▄▄▄▄▄▄▄", nil)
		s.drawText(4, 21, width, activeTabStyle, "██████████", nil)
		s.drawText(5, 21, width, activeTabStyle, "▀▀▀▀▀▀▀▀▀▀", nil)
		s.drawText(4, 23, width, activeTabTextStyle, "STACKS", nil)
	case VIEW_OPERATIONS:
		s.drawText(3, 31, width, activeTabStyle, "▄▄▄▄▄▄▄▄▄▄▄▄▄▄", nil)
		s.drawText(4, 31, width, activeTabStyle, "██████████████", nil)
		s.drawText(5, 31, width, activeTabStyle, "▀▀▀▀▀▀▀▀▀▀▀▀▀▀", nil)
		s.drawText(4, 33, width, activeTabTextStyle, "OPERATIONS", nil)
	case VIEW_DETAILS:
		s.drawText(3, 45, width, activeTabStyle, "▄▄▄▄▄▄▄▄▄▄▄", nil)
		s.drawText(4, 45, width, activeTabStyle, "███████████", nil)
		s.drawText(5, 45, width, activeTabStyle, "▀▀▀▀▀▀▀▀▀▀▀", nil)
		s.drawText(4, 47, width, activeTabTextStyle, "DETAILS", nil)
	}

	s.drawText(6, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "Last Refresh:", s.LastRefreshed.Format(time.TimeOnly)), nil)

	if s.AllStacks {
		s.drawText(7, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "Stack:", "<ALL>"), nil)
	} else {
		s.drawText(7, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "Stack:", aws.ExtractStackNameFromArn(s.SelectedStack)), nil)
	}
	if s.AllOperations {
		s.drawText(8, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "Operation:", "<ALL>"), nil)
	} else {
		s.drawText(8, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "Operation:", s.SelectedOperation), nil)
	}
	s.drawText(9, 0, width, DefaultStyle, fmt.Sprintf("%-22s %d", "Stack Count:", len(s.dataSet.GetStackArns())), nil)
	s.drawText(10, 0, width, DefaultStyle, fmt.Sprintf("%-22s %d", "Operation Count:", len(s.dataSet.GetOperations(s.SelectedStack, s.AllStacks))), nil)
	s.drawText(11, 0, width, DefaultStyle, fmt.Sprintf("%-22s %d", "Interval Count:", len(intervals)), nil)
	if len(intervals) > 0 {
		windowInterval := aws.GetWindowInterval(&intervals)
		s.drawText(12, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "Duration:", windowInterval.End.Timestamp.Sub(windowInterval.Start.Timestamp)), nil)
	}
	s.drawText(13, 0, width, DefaultStyle, "", &fillerRune)
}

func (s *State) renderDetails(row int) {
	width, _ := s.screen.Size()
	intervals := s.dataSet.GetSortedIntervals(s.SelectedStack, s.SelectedOperation, s.AllStacks, s.AllOperations)

	if len(intervals) == 0 {
		return
	}

	interval := intervals[s.selectedIndex]

	if interval.Start != nil {
		s.drawText(row, 0, width, DefaultStyle, "START EVENT", nil)
		row += 2
		row += renderEvent(s, interval.Start, row)
	}

	for _, event := range interval.Intermediate {
		row += 1
		s.drawText(row, 0, width, DefaultStyle, "INTERMEDIATE EVENT", nil)
		row += 2
		row += renderEvent(s, event, row)
	}

	if interval.End != nil && interval.End.EventId != "" {
		row += 1
		s.drawText(row, 0, width, DefaultStyle, "END EVENT", nil)
		row += 2
		renderEvent(s, interval.End, row)
	}
}

func renderEvent(s *State, event *aws.Event, row int) int {
	if event == nil {
		return 0
	}
	width, _ := s.screen.Size()
	s.drawText(row, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "EventId:", event.EventId), nil)
	s.drawText(row+1, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "StackId:", aws.ExtractStackNameFromArn(event.StackId)), nil)
	s.drawText(row+2, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "Timestamp:", event.Timestamp), nil)
	s.drawText(row+3, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "ResourceStatus:", event.ResourceStatus), nil)
	s.drawText(row+4, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "ResourceType:", event.ResourceType), nil)
	s.drawText(row+5, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "LogicalResourceId:", event.LogicalResourceId), nil)
	s.drawText(row+6, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "PhysicalResourceId:", event.PhysicalResourceId), nil)
	s.drawText(row+7, 0, width, DefaultStyle, fmt.Sprintf("%-22s %s", "ResourceStatusReason:", event.ResourceStatusReason), nil)
	return 8
}

func (s *State) renderStacks(row int) {
	width, _ := s.screen.Size()
	s.drawText(
		row,
		0,
		width,
		DefaultStyle,
		"STACK ARN",
		nil,
	)
	for i, stackArn := range s.dataSet.GetStackArns() {
		textStyle := DefaultStyle
		if stackArn == s.SelectedStack && !s.AllStacks {
			textStyle = HighlightedStyle
		}
		s.drawText(
			row+i+1,
			0,
			width,
			textStyle,
			aws.ExtractStackNameFromArn(stackArn),
			nil,
		)
	}
}

func (s *State) renderOperation(row int) {
	width, _ := s.screen.Size()
	events := s.dataSet.GetOperations(s.SelectedStack, s.AllStacks)
	s.drawText(
		row,
		0,
		width,
		DefaultStyle,
		fmt.Sprintf("%-36s  %-20s  %-18s  %s", "EVENT ID", "TIMESTAMP", "RESOURCE STATUS", "LOGICAL RESOURCE ID"),
		nil,
	)
	for i, _ := range events {
		event := events[i]
		textStyle := DefaultStyle
		if event.EventId == s.SelectedOperation && !s.AllOperations {
			textStyle = HighlightedStyle
		}
		s.drawText(
			row+i+1,
			0,
			width,
			textStyle,
			fmt.Sprintf("%-36s  %-20s  %-18s  %s", event.EventId, event.Timestamp.Format(time.RFC3339), event.ResourceStatus, event.LogicalResourceId),
			nil,
		)
	}
}

func (s *State) renderWaterfall(row int) {
	textWidth := 52
	width, height := s.screen.Size()
	intervals := s.dataSet.GetSortedIntervals(s.SelectedStack, s.SelectedOperation, s.AllStacks, s.AllOperations)
	if len(intervals) == 0 {
		s.drawText(row+1, 3, textWidth, DefaultStyle, "No intervals found", nil)
		return
	}
	windowInterval := aws.GetWindowInterval(&intervals)
	totalIntervals := len(intervals)
	totalRows := height - row
	startIndex := 0
	endIndex := totalRows
	if s.selectedIndex+totalRows >= totalIntervals {
		startIndex = totalIntervals - totalRows
		endIndex = totalIntervals - 1
	} else {
		startIndex = s.selectedIndex
		endIndex = s.selectedIndex + totalRows - 1
	}
	drawCount := 0
	for i, interval := range intervals {
		if i < startIndex || i > endIndex {
			continue
		}
		textStyle := DefaultStyle
		backgroundColor := tcell.ColorReset
		logicalResourceId := "-"
		if interval.Start != nil {
			logicalResourceId = interval.Start.LogicalResourceId
		}
		var fillerRunePtr *rune = nil
		if i == s.selectedIndex {
			textStyle = HighlightedStyle
			backgroundColor = tcell.ColorWhite
			fillerRune := ' '
			fillerRunePtr = &fillerRune
		}
		isStackIndicator := "   "
		if interval.Start.IsOperation() {
			isStackIndicator = " ◯ "
		}
		s.drawText(row+drawCount, 0, 5, textStyle, isStackIndicator, fillerRunePtr)
		s.drawText(row+drawCount, 3, textWidth+4, textStyle, logicalResourceId, fillerRunePtr)
		s.drawText(row+drawCount, textWidth+4, textWidth+5, textStyle, " ", fillerRunePtr)
		drawInterval(s.screen, row+drawCount, textWidth+5, width, windowInterval, interval, backgroundColor)
		drawCount++
	}
}

func (s *State) renderLegend(row int) {
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)
	purple := tcell.StyleDefault.Foreground(tcell.ColorPurple)
	orange := tcell.StyleDefault.Foreground(tcell.ColorOrange)
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	yellow := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	darkOrange := tcell.StyleDefault.Foreground(tcell.ColorDarkOrange)
	gray := tcell.StyleDefault.Foreground(tcell.ColorGray)
	s.drawText(row+0, 0, 18, green, "□□□□□□", nil)
	s.drawText(row+0, 8, 64, DefaultStyle, "CREATE_IN_PROGRESS", nil)
	s.drawText(row+1, 0, 18, green, "■■■■■■", nil)
	s.drawText(row+1, 8, 64, DefaultStyle, "CREATE_COMPLETE", nil)
	s.drawText(row+2, 0, 18, green, "◩◩◩◩◩◩", nil)
	s.drawText(row+2, 8, 64, DefaultStyle, "CREATE_FAILED", nil)
	s.drawText(row+3, 0, 18, red, "□□□□□□", nil)
	s.drawText(row+3, 8, 64, DefaultStyle, "DELETE_IN_PROGRESS", nil)
	s.drawText(row+4, 0, 18, red, "■■■■■■", nil)
	s.drawText(row+4, 8, 64, DefaultStyle, "DELETE_COMPLETE", nil)
	s.drawText(row+5, 0, 18, red, "◩◩◩◩◩◩", nil)
	s.drawText(row+5, 8, 64, DefaultStyle, "DELETE_FAILED", nil)
	s.drawText(row+6, 0, 18, gray, "□□□□□□", nil)
	s.drawText(row+6, 8, 64, DefaultStyle, "REVIEW_IN_PROGRESS", nil)
	s.drawText(row+7, 0, 18, purple, "□□□□□□", nil)
	s.drawText(row+7, 8, 64, DefaultStyle, "IMPORT_IN_PROGRESS", nil)
	s.drawText(row+8, 0, 18, purple, "■■■■■■", nil)
	s.drawText(row+8, 8, 64, DefaultStyle, "IMPORT_COMPLETE", nil)
	s.drawText(row+9, 0, 18, orange, "□□□□□□", nil)
	s.drawText(row+9, 8, 64, DefaultStyle, "ROLLBACK_IN_PROGRESS", nil)
	s.drawText(row+10, 0, 18, orange, "■■■■■■", nil)
	s.drawText(row+10, 8, 64, DefaultStyle, "ROLLBACK_COMPLETE", nil)
	s.drawText(row+11, 0, 18, orange, "◩◩◩◩◩◩", nil)
	s.drawText(row+11, 8, 64, DefaultStyle, "ROLLBACK_FAILED", nil)
	s.drawText(row+12, 0, 18, blue, "□□□□□□", nil)
	s.drawText(row+12, 8, 64, DefaultStyle, "UPDATE_IN_PROGRESS", nil)
	s.drawText(row+13, 0, 18, blue, "■■■■■■", nil)
	s.drawText(row+13, 8, 64, DefaultStyle, "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS", nil)
	s.drawText(row+14, 0, 18, blue, "■■■■■■", nil)
	s.drawText(row+14, 8, 64, DefaultStyle, "UPDATE_COMPLETE", nil)
	s.drawText(row+15, 0, 18, blue, "◩◩◩◩◩◩", nil)
	s.drawText(row+15, 8, 64, DefaultStyle, "UPDATE_FAILED", nil)
	s.drawText(row+16, 0, 18, yellow, "□□□□□□", nil)
	s.drawText(row+16, 8, 64, DefaultStyle, "UPDATE_ROLLBACK_IN_PROGRESS", nil)
	s.drawText(row+17, 0, 18, yellow, "□□□□□□", nil)
	s.drawText(row+17, 8, 64, DefaultStyle, "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS", nil)
	s.drawText(row+18, 0, 18, yellow, "■■■■■■", nil)
	s.drawText(row+18, 8, 64, DefaultStyle, "UPDATE_ROLLBACK_COMPLETE", nil)
	s.drawText(row+19, 0, 18, yellow, "◩◩◩◩◩◩", nil)
	s.drawText(row+19, 8, 64, DefaultStyle, "UPDATE_ROLLBACK_FAILED", nil)
	s.drawText(row+20, 0, 18, darkOrange, "□□□□□□", nil)
	s.drawText(row+20, 8, 64, DefaultStyle, "IMPORT_ROLLBACK_IN_PROGRESS", nil)
	s.drawText(row+21, 0, 18, darkOrange, "■■■■■■", nil)
	s.drawText(row+21, 8, 64, DefaultStyle, "IMPORT_ROLLBACK_COMPLETE", nil)
	s.drawText(row+22, 0, 18, darkOrange, "◩◩◩◩◩◩", nil)
	s.drawText(row+22, 8, 64, DefaultStyle, "IMPORT_ROLLBACK_FAILED", nil)
}

func getIntervalRune(interval aws.Interval) rune {
	if interval.End == nil {
		return '■'
	}
	status := string(interval.End.ResourceStatus)
	if strings.HasSuffix(status, "_IN_PROGRESS") {
		return '□'
	}
	if strings.HasSuffix(status, "_FAILED") {
		return '◩'
	}
	return '■'
}

func getIntervalColor(interval aws.Interval) tcell.Color {
	status := string(interval.End.ResourceStatus)
	switch true {
	case strings.HasPrefix(status, "UPDATE_ROLLBACK_"):
		return tcell.ColorYellow
	case strings.HasPrefix(status, "IMPORT_ROLLBACK_"):
		return tcell.ColorDarkOrange
	case strings.HasPrefix(status, "CREATE_"):
		return tcell.ColorGreen
	case strings.HasPrefix(status, "DELETE_"):
		return tcell.ColorRed
	case strings.HasPrefix(status, "UPDATE_"):
		return tcell.ColorBlue
	case strings.HasPrefix(status, "IMPORT_"):
		return tcell.ColorGray
	case strings.HasPrefix(status, "REVIEW_"):
		return tcell.ColorPurple
	case strings.HasPrefix(status, "ROLLBACK_"):
		return tcell.ColorOrange
	}
	return tcell.ColorReset
}

func drawInterval(s tcell.Screen, row, colStart, colEnd int, windowInterval, interval aws.Interval, backgroundColor tcell.Color) {
	windowEnd := windowInterval.End.Timestamp
	windowStart := windowInterval.Start.Timestamp
	intStart := interval.Start.Timestamp
	intEnd := time.Now()
	if interval.End != nil {
		intEnd = interval.End.Timestamp
	}
	colWidth := colEnd - colStart
	secondsInCol := windowEnd.Sub(windowStart).Seconds() / float64(colWidth)
	carryOver := windowStart.Add(0)
	intervalRune := getIntervalRune(interval)
	intervalColor := getIntervalColor(interval)
	intervalStyle := tcell.StyleDefault.Background(backgroundColor).Foreground(intervalColor).Bold(true)
	lineStyle := tcell.StyleDefault.Background(backgroundColor).Foreground(tcell.ColorGray).Dim(true)
	for i := 0; i < colWidth; i++ {
		nextCarryOver := carryOver.Add(time.Duration(secondsInCol * float64(time.Second)))
		if intStart.Before(nextCarryOver) && intEnd.After(carryOver) {
			s.SetContent(colStart+i, row, intervalRune, nil, intervalStyle)
		} else {
			s.SetContent(colStart+i, row, '─', nil, lineStyle)
		}
		carryOver = nextCarryOver
	}
}

func (s *State) drawText(row, colStart, colEnd int, style tcell.Style, text string, filler *rune) {
	maxLength := colEnd - colStart
	if len(text) > maxLength {
		text = text[:maxLength-1]
		text = text + "…"
	}

	col := colStart
	for _, r := range []rune(text) {
		if col >= colEnd {
			return
		}
		s.screen.SetContent(col, row, r, nil, style)
		col++
	}
	if filler != nil {
		for col < colEnd {
			s.screen.SetContent(col, row, *filler, nil, style)
			col++
		}
	}
}
