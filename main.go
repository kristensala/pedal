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

    workoutData := make([]cmd.Step, 0)
    droppedFile := make([]string, 0)

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
        if (rl.IsFileDropped()) {
            fmt.Println("here")
            droppedFile = rl.LoadDroppedFiles()

            if (len(droppedFile) > 0) {
                workoutData = cmd.ParseWorkoutFile(droppedFile[0])
                rl.UnloadDroppedFiles()
            }
        }

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

        if (len(workoutData) > 0) {
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

func renderWorkoutView(data []cmd.Step) {
    renderGraph(data, 200)
}

func renderGraph(data []cmd.Step, height float32) rl.Rectangle {
    canvasX, canvasY := 0, float32(rl.GetScreenHeight()) - height

    canvas := rl.Rectangle{
        X: float32(canvasX),
        Y: canvasY,
        Height: height,
        Width: float32(rl.GetScreenWidth()),
    }

    // draw canvas element (the parent element)
    rl.DrawRectangleRec(canvas, rl.Blue)

    // TODO:
    // test the steps (this is random testing currently)
    // need to calculate correct pixle gap and use canvas
    // as the parent element
    // first set the high target and then low (otherwise 
    // high will cover the smaller rectangle)
    
    // X is seconds (time)
    // Y is power

    // default baseline power is 200 so 50% canvas height is 200W
    // knowing that max canvas height is 400w

    // POWER distribution:
    // if canvas height is 400px then 1w = 1px => 400px / 400W
    // if canvas height = 500px then 1w = 500 / 400 and so on

    // TIME distribution:
    // Data needed: canvas width and total workout time in seconds
    // calculation same as it is for power

    // each step will be a simple rectangle

    /*
    firstStep := data[0]
    stepOneOne := rl.Rectangle{
        X: 0,
        Y: float32(rl.GetScreenHeight()) - float32(firstStep.TargetHigh),
        Height: float32(firstStep.TargetHigh),
        Width: 100,
    }
    rl.DrawRectangleRec(stepOneOne, rl.Green)

    stepOne := rl.Rectangle{
        X: 0,
        Y: float32(rl.GetScreenHeight()) - float32(firstStep.TargetLow),
        Height: float32(firstStep.TargetLow),
        Width: 100,
    }
    rl.DrawRectangleRec(stepOne, rl.Yellow)
    */


    return canvas
}

