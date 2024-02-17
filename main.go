package main

import (
	"fmt"
	"pedal/cmd"

	rl "github.com/gen2brain/raylib-go/raylib"
)


type PedalState struct {
    currentWorkoutFile string
}

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


    // Needel
    // note: X changes on every second
    // and based on X I have to get the current block
    startPos := rl.Vector2{
        X: 100,
        Y: canvas.Y,
    }
    endPos := rl.Vector2{
        X: 100,
        Y: canvas.Y + canvas.Height,
    }

    rl.DrawLineV(startPos, endPos, rl.Red)

    return canvas
}

