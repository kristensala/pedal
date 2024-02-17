package main

import (
	"fmt"
	"pedal/cmd"

	rl "github.com/gen2brain/raylib-go/raylib"
)


type PedalState struct {
    currentWorkoutFile string
}

// Get the current workout state
var needlePositionOnCanvas float64 = 0
var needleIncrementOnCanvas float64 = 0
var currentWorkoutBlockNr int = 0
var blockChangeAtSecond = 0

func main() {
	rl.InitWindow(800, 450, "Pedal")
	defer rl.CloseWindow()

    workoutData := cmd.DataSet{}
    droppedFile := make([]string, 0)

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
        if (rl.IsFileDropped()) {
            droppedFile = rl.LoadDroppedFiles()

            if (len(droppedFile) > 0) {
                workoutData = cmd.ParseWorkoutFile(droppedFile[0])
                rl.UnloadDroppedFiles()
            }
        }

        if (rl.IsKeyDown(rl.KeyRight)) {
            needlePositionOnCanvas = needlePositionOnCanvas + needleIncrementOnCanvas
            getBlockBasedOnNeedlePos(workoutData)
        }

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

        if (len(workoutData.Blocks) > 0) {
            renderWorkoutView(workoutData)
        } else {
            rl.DrawText("Drop a .FIT workout file here!",
                190,
                200,
                20,
                rl.LightGray)
        }

		rl.EndDrawing()
	}
}

func renderWorkoutView(data cmd.DataSet) {
    rl.DrawText(fmt.Sprint(data.TotalDurationSeconds),
        190,
        200,
        20,
        rl.Green)

    // canvas is always 50% of the screen height
    canvasHeight := float32(rl.GetScreenHeight()) * float32(0.5)
    renderGraph(data, canvasHeight)
}

func renderGraph(data cmd.DataSet, height float32) rl.Rectangle {
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

    needleIncrementOnCanvas = timeGap
    
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


    // Needle
    // note: X changes on every second
    // and based on X I have to get the current block
    startPos := rl.Vector2{
        X: float32(needlePositionOnCanvas),
        Y: canvas.Y,
    }
    endPos := rl.Vector2{
        X: float32(needlePositionOnCanvas),
        Y: canvas.Y + canvas.Height,
    }

    rl.DrawLineV(startPos, endPos, rl.Red)

    return canvas
}

// Values needed
// Dataset; Needle pos and needle increment; total duration in seconds
func getBlockBasedOnNeedlePos(dataSet cmd.DataSet) {
    // gives me the actual second we are on
    trueNeedlePos := uint32(needlePositionOnCanvas / needleIncrementOnCanvas)

    if (len(dataSet.Blocks) == 0) {
        return
    }

    // fix: do not index into array every time
    // save the block into some sort of an application state
    block := dataSet.Blocks[currentWorkoutBlockNr]
    fmt.Printf("%d; %d ;", int(trueNeedlePos), block.DurationSeconds)

    if (currentWorkoutBlockNr == 0) {
        blockChangeAtSecond = int(block.DurationSeconds)
    }

    if (trueNeedlePos > uint32(blockChangeAtSecond)) {
        currentWorkoutBlockNr = currentWorkoutBlockNr + 1

        // TODO: send some sort of a signal
        if (currentWorkoutBlockNr > len(dataSet.Blocks)) {
            fmt.Println("workout ENDED")
            return;
        }

        newBlock := dataSet.Blocks[currentWorkoutBlockNr]
        
        blockChangeAtSecond = int(trueNeedlePos) + int(newBlock.DurationSeconds)
        fmt.Printf("next block starts at: %d second\n", blockChangeAtSecond)

    } else {
        fmt.Println(block.TargetLow)
    }


}
