package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/null93/waterfall/sdk/aws"
	"github.com/null93/waterfall/sdk/gui"
	"github.com/spf13/cobra"
)

var (
	Version            = "0.0.1"
	VerboseOutput      = false
	AwsProfile         = ""
	RefreshInterval    = 15
	IgnoreNestedStacks = false
	Debug              = false
)

var RootCmd = &cobra.Command{
	Use:     "waterfall STACK_NAME",
	Version: Version,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// initialize aws config and make sure stack exists

		config, configErr := aws.GetConfig(AwsProfile)

		if configErr != nil {
			exitWithError(1, "failed to load aws config", configErr)
		}

		authErr := aws.CheckAuth(config)

		if authErr != nil {
			exitWithError(2, "failed to authenticate", authErr)
		}

		arn, stackArnErr := aws.GetStackArnFromStackName(config, args[0])

		if stackArnErr != nil {
			if stackArnErr == aws.StackNotFoundErr {
				exitWithError(3, "stack not found", stackArnErr)
			}
			exitWithError(4, "unknown error", stackArnErr)
		}

		// pull stack data from aws

		dataSet := aws.NewDataSet(config, arn)

		if !IgnoreNestedStacks {
			if nestedErr := dataSet.AddNestedStacks(); nestedErr != nil {
				exitWithError(5, "failed to get nested stacks", nestedErr)
			}
		}

		if refreshErr := dataSet.Refresh(); refreshErr != nil {
			exitWithError(6, "failed to get stack events", refreshErr)
		}

		if len(dataSet.GetStackEvents(arn)) == 0 {
			exitWithError(7, "no events found", nil)
		}

		// print debug info and exit if debug mode is enabled

		if Debug {
			for _, event := range dataSet.GetAllStackEvents() {
				fmt.Printf("%s - %s - %s\n", event.Timestamp.Format(time.RFC3339), event.ResourceStatus, event.LogicalResourceId)
			}
			fmt.Println()
			intervals := dataSet.GetSortedIntervals("", "", true, true)
			for _, interval := range intervals {
				if interval.Start.IsOperation() {
					fmt.Println("--- User Initiated ---")
					fmt.Println()
				}
				fmt.Printf("Start: %20s - %-35s - %s\n", interval.Start.Timestamp.Format(time.RFC3339), interval.Start.ResourceStatus, interval.Start.LogicalResourceId)
				for _, event := range interval.Intermediate {
					fmt.Printf("       %20s - %-35s - %s\n", event.Timestamp.Format(time.RFC3339), event.ResourceStatus, event.LogicalResourceId)
				}
				fmt.Printf("End:   %20s - %-35s - %s\n", interval.End.Timestamp.Format(time.RFC3339), interval.End.ResourceStatus, interval.End.LogicalResourceId)
				fmt.Println()
			}
			return
		}

		// initialize screen

		screen, screenErr := tcell.NewScreen()
		defer screen.Fini()

		if screenErr != nil {
			exitWithError(8, "failed to create screen", screenErr)
		}

		if screenInitErr := screen.Init(); screenInitErr != nil {
			exitWithError(9, "failed to initialize screen", screenInitErr)
		}

		// initialize output state

		output := gui.NewState(screen, dataSet)
		output.CurrentView = gui.VIEW_WATERFALL
		output.SelectedOperation = dataSet.GetLatestOperation(output.SelectedStack, output.AllStacks)
		output.Render()

		// run ticker to update data and render

		if RefreshInterval > 0 {
			ticker := time.NewTicker(time.Second * time.Duration(RefreshInterval))
			defer ticker.Stop()

			go func() {
				for range ticker.C {
					if !dataSet.IsLoading() {
						output.LastRefreshed = time.Now()
						if !IgnoreNestedStacks {
							dataSet.AddNestedStacks()
						}
						dataSet.Refresh()
						output.Render()
					}
				}
			}()
		}

		// main loop

		for {
			event := screen.PollEvent()
			switch event := event.(type) {
			case *tcell.EventResize:
				output.Render()
				screen.Sync()
			case *tcell.EventKey:
				if event.Key() == tcell.KeyCtrlC || event.Key() == tcell.KeyEscape {
					screen.Fini()
					os.Exit(0)
				}
				if event.Key() == tcell.KeyUp {
					if output.CurrentView == gui.VIEW_WATERFALL {
						output.DecrementSelected()
						output.Render()
					}
				}
				if event.Key() == tcell.KeyDown {
					if output.CurrentView == gui.VIEW_WATERFALL {
						output.IncrementSelected()
						output.Render()
					}
				}
				if (event.Key() == tcell.KeyRight && !output.AllOperations) || (event.Key() == tcell.KeyDown && output.CurrentView == gui.VIEW_OPERATIONS) {
					output.IncrementOperationSelected()
					output.ResetSelectedIndex()
					output.Render()
				}
				if (event.Key() == tcell.KeyLeft && !output.AllOperations) || (event.Key() == tcell.KeyUp && output.CurrentView == gui.VIEW_OPERATIONS) {
					output.DecrementOperationSelected()
					output.ResetSelectedIndex()
					output.Render()
				}
				if (event.Key() == tcell.KeyTab && !output.AllStacks) || (event.Key() == tcell.KeyDown && output.CurrentView == gui.VIEW_STACKS) {
					output.IncrementSelectedStack()
					output.SelectedOperation = dataSet.GetLatestOperation(output.SelectedStack, output.AllStacks)
					output.ResetSelectedIndex()
					output.Render()
				}
				if (event.Key() == tcell.KeyBacktab && !output.AllStacks) || (event.Key() == tcell.KeyUp && output.CurrentView == gui.VIEW_STACKS) {
					output.DecrementSelectedStack()
					output.SelectedOperation = dataSet.GetLatestOperation(output.SelectedStack, output.AllStacks)
					output.ResetSelectedIndex()
					output.Render()
				}
				if event.Rune() == 'h' {
					output.CurrentView = gui.VIEW_HELP
					output.Render()
				}
				if event.Rune() == 's' {
					output.CurrentView = gui.VIEW_STACKS
					output.Render()
				}
				if event.Rune() == 'o' {
					output.CurrentView = gui.VIEW_OPERATIONS
					output.Render()
				}
				if event.Rune() == 'r' && !dataSet.IsLoading() {
					output.LastRefreshed = time.Now()
					if !IgnoreNestedStacks {
						dataSet.AddNestedStacks()
					}
					dataSet.Refresh()
					output.Render()
				}
				if event.Rune() == 'O' {
					output.AllOperations = !output.AllOperations
					output.SelectedOperation = dataSet.GetLatestOperation(output.SelectedStack, output.AllStacks)
					output.ResetSelectedIndex()
					output.Render()
				}
				if event.Rune() == 'S' {
					output.AllStacks = !output.AllStacks
					output.SelectedOperation = dataSet.GetLatestOperation(output.SelectedStack, output.AllStacks)
					output.ResetSelectedIndex()
					output.Render()
				}

				if event.Key() == tcell.KeyEnter {
					if output.CurrentView == gui.VIEW_WATERFALL {
						output.CurrentView = gui.VIEW_DETAILS
					} else {
						output.CurrentView = gui.VIEW_WATERFALL
					}
					output.Render()
				}
			}
		}

	},
}

func exitWithError(exitCode int, message string, err error) {
	fmt.Printf("Error: %s\n", message)
	if VerboseOutput {
		fmt.Println(err)
	}
	os.Exit(exitCode)
}

func init() {
	RootCmd.Flags().SortFlags = true
	RootCmd.Flags().StringVarP(&AwsProfile, "profile", "p", AwsProfile, "aws profile name")
	RootCmd.Flags().BoolVarP(&VerboseOutput, "verbose", "v", VerboseOutput, "verbose output")
	RootCmd.Flags().IntVarP(&RefreshInterval, "refresh", "r", RefreshInterval, "refresh interval in secs, 0 to disable")
	RootCmd.Flags().BoolVarP(&Debug, "debug", "d", Debug, "debug mode")
	RootCmd.Flags().BoolVarP(&IgnoreNestedStacks, "no-nested-stacks", "n", IgnoreNestedStacks, "do not process nested stacks")
	RootCmd.Flags().MarkHidden("debug")
}
