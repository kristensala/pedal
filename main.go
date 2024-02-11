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
		rl.DrawText("Drop a .FIT workout file here!",
            190,
            200,
            20,
            rl.LightGray)

        if (len(workoutData) > 0) {
            renderWorkoutView(workoutData)
        }

		rl.EndDrawing()
	}
}

// needs to be a .fit workout file
// not a workout summary
func validateWorkoutFile() {

}

// todo:
// render layout for graph and
// display power and time and hr etc
func renderWorkoutView(data []cmd.Step) {
    renderGraph(200)
}


// todo: get the fit workout file data to render the graph
func renderGraph(height float32) rl.Rectangle {
    canvas := rl.Rectangle{
        X: 0,
        Y: float32(rl.GetScreenHeight()) - height,
        Height: height,
        Width: float32(rl.GetScreenWidth()),
    }

    rl.DrawRectangleRec(canvas, rl.Blue)

    return canvas
}
