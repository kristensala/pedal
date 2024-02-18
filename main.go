package main

import (
	"fmt"
	"pedal/cmd"

	rl "github.com/gen2brain/raylib-go/raylib"
	//gui "github.com/gen2brain/raylib-go/raygui"
)

type AppState struct {
    dataSet cmd.DataSet
    screen CurrentScreen
}

// Get the current workout state
// through the needle pos
var (
    needlePosX float64 = 0
    needlePosPercent float64 = 0
    needleIncrementX float64 = 0
    currentWorkoutBlockNr int = 0
    workoutBlockChangeAtSecond = 0
)

type CurrentScreen int
const (
    TitleScreen CurrentScreen = iota
    WorkoutScreen
)

func main() {
    const windowHeight = 450
    const windowWidth = 800

    droppedFile := make([]string, 0)
    appState := AppState{
        screen: TitleScreen,
    }

	rl.InitWindow(windowWidth, windowHeight, "Pedal")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

    // works for my enormous screen
    rl.SetWindowPosition(
        (rl.GetMonitorWidth(0) - (windowWidth / 2)),
        (rl.GetMonitorHeight(0) / 2) - (windowHeight / 2))

	for !rl.WindowShouldClose() {
        // Update
        //------------------------------------
        switch appState.screen {
        case TitleScreen: {
            if (rl.IsFileDropped()) {
                droppedFile = rl.LoadDroppedFiles()

                if (len(droppedFile) > 0) {
                    appState.dataSet = cmd.ParseWorkoutFile(droppedFile[0])
                    appState.screen = WorkoutScreen
                    rl.UnloadDroppedFiles()
                }
            }
            break;
        }
        case WorkoutScreen: {
            // note: to move the needle manually
            if (rl.IsKeyDown(rl.KeyRight)) {
                needlePosX = needlePosX + needleIncrementX
                needlePosPercent = (needlePosX * 100) / float64(rl.GetScreenWidth())
                getBlockBasedOnNeedlePos(appState.dataSet)
            }
            break;
        }
        default: break;
        }
        //------------------------------------

        // Draw
        //------------------------------------
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

        switch appState.screen {
        case TitleScreen: {
            rl.DrawText("Drop a .FIT workout file here!",
                190,
                200,
                20,
                rl.LightGray)
            break;
        }
        case WorkoutScreen: {
            if (len(appState.dataSet.Blocks) > 0) {
                appState.drawWorkoutCanvas()
            }
            break;
        }
        default: break;
        }

		rl.EndDrawing()
	}
}

func (state AppState) drawWorkoutCanvas() {
    rl.DrawText(fmt.Sprint(state.dataSet.TotalDurationSeconds),
        190,
        200,
        20,
        rl.Green)

    // canvas is always 50% of the screen height
    canvasHeight := float32(rl.GetScreenHeight()) * float32(0.5)
    canvas := renderCanvas(state.dataSet, canvasHeight)

    // makes sure that on window resize 
    // the needle is in the correct poisition
    // Assumes that canvas width is window width
    needlePosX = (float64(needlePosPercent) * float64(canvas.Width)) / 100

    // Draw the Needle
    startPos := rl.Vector2{
        X: float32(needlePosX),
        Y: canvas.Y,
    }
    endPos := rl.Vector2{
        X: float32(needlePosX),
        Y: canvas.Y + canvas.Height,
    }

    rl.DrawLineV(startPos, endPos, rl.Red)
}

func renderCanvas(data cmd.DataSet, height float32) rl.Rectangle {
    canvasX, canvasY := 0, float32(rl.GetScreenHeight()) - height

    canvas := rl.Rectangle{
        X: float32(canvasX),
        Y: canvasY,
        Height: height,
        Width: float32(rl.GetScreenWidth()),
    }

    // draw canvas element (the parent element)
    rl.DrawRectangleRec(canvas, rl.Blue)

    // calculate 1s and 1w pixels
    timeGap := float64(canvas.Width) / float64(data.TotalDurationSeconds)
    powerGap := float64(canvas.Height) / 600 // 600W is max power to display

    needleIncrementX = timeGap
    
    blockX := 0.0
    for _, b := range data.Blocks {
        blockHighEndHeight := float64(b.TargetHigh) * powerGap
        blockLowEndHeight := float64(b.TargetLow) * powerGap
        blockWidth := float64(b.DurationSeconds) * timeGap

        // blocks are dependent of the canvas position
        // if canvas heigh and location is changed
        // block will move with it
        // or at least should
        block := rl.Rectangle{
            X: float32(blockX),
            Y: canvas.Y + canvas.Height - float32(blockHighEndHeight),
            Height: float32(blockHighEndHeight),
            Width: float32(blockWidth),
        }
        rl.DrawRectangleRec(block, rl.Yellow)

        lowEndBlock := rl.Rectangle{
            X: float32(blockX),
            Y: canvas.Y + canvas.Height - float32(blockLowEndHeight),
            Height: float32(blockLowEndHeight),
            Width: float32(blockWidth),
        }
        rl.DrawRectangleRec(lowEndBlock, rl.LightGray) //fix: better colors

        blockX = blockX + blockWidth
    }

    return canvas
}

// Values needed
// Dataset; Needle pos and needle increment; total duration in seconds
func getBlockBasedOnNeedlePos(dataSet cmd.DataSet) {
    // gives me the actual second we are on
    trueNeedlePos := uint32(needlePosX / needleIncrementX)

    if (len(dataSet.Blocks) == 0) {
        return
    }

    // fix: do not index into array every time
    // save the block into some sort of an application state
    block := dataSet.Blocks[currentWorkoutBlockNr]
    fmt.Printf("%d; %d ;", int(trueNeedlePos), block.DurationSeconds)

    if (currentWorkoutBlockNr == 0) {
        workoutBlockChangeAtSecond = int(block.DurationSeconds)
    }

    if (trueNeedlePos > uint32(workoutBlockChangeAtSecond)) {
        currentWorkoutBlockNr = currentWorkoutBlockNr + 1

        // TODO: send some sort of a signal
        if (currentWorkoutBlockNr > len(dataSet.Blocks)) {
            fmt.Println("workout ENDED")
            return;
        }

        newBlock := dataSet.Blocks[currentWorkoutBlockNr]
        
        workoutBlockChangeAtSecond = int(trueNeedlePos) + int(newBlock.DurationSeconds)
        fmt.Printf("next block starts at: %d second\n", workoutBlockChangeAtSecond)

    } else {
        fmt.Println(block.TargetLow)
    }
}

