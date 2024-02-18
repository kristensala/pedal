package main

import (
	"fmt"
	"image/color"
	"log"
	"pedal/cmd"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// local application settings
// if needed
type AppState struct {
    DataSet cmd.DataSet
    Screen ApplicationScreen
    WorkoutElapsedTime uint32
    CurrentInterval cmd.Interval
    CurrentIntervalNumber int
    NextIntervalStartsAt int
}

// Needle state on canvas
// TODO:
type NeedleState struct {
    PosX float64
    PosPercentage float64
    IncrementX float64
}

// Get the current workout state
// through the needle pos
var (
    needlePosX float64 = 0
    needlePosPercent float64 = 0
    needleIncrementX float64 = 0
)

type ApplicationScreen int
const (
    TitleScreen ApplicationScreen = iota
    WorkoutScreen
)

func InitApp() (state AppState) {
    state.Screen = TitleScreen
    state.WorkoutElapsedTime = 0
    state.CurrentIntervalNumber = 0
    state.NextIntervalStartsAt = 0
    return state
}

func main() {
    // Init
    //-------------------------------------
    const (
        windowHeight = 450
        windowWidth = 800
    )

    droppedFile := make([]string, 0)
    appState := InitApp()

	rl.InitWindow(windowWidth, windowHeight, "Pedal")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

    // works for my enormous screen
    rl.SetWindowPosition(
        (rl.GetMonitorWidth(0) - (windowWidth / 2)),
        (rl.GetMonitorHeight(0) / 2) - (windowHeight / 2))
    //--------------------------------------

	for !rl.WindowShouldClose() {
        // Update
        //------------------------------------
        switch appState.Screen {
        case TitleScreen: 
            if (rl.IsFileDropped()) {
                droppedFile = rl.LoadDroppedFiles()

                if (len(droppedFile) > 0) {
                    appState.DataSet = cmd.ParseWorkoutFile(droppedFile[0])
                    rl.UnloadDroppedFiles()

                    appState.Screen = WorkoutScreen
                }
            }
            break;

        case WorkoutScreen:
            // note: to move the needle manually
            if (rl.IsKeyDown(rl.KeyRight)) {
                needlePosX = needlePosX + needleIncrementX
                needlePosPercent = (needlePosX * 100) / float64(rl.GetScreenWidth())
                appState.GetBlockBasedOnNeedlePos()
            }
            break;

        default:
            break;
        }
        //------------------------------------

        // Draw
        //------------------------------------
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

        switch appState.Screen {
        case TitleScreen:
            rl.DrawText("Drop a .FIT workout file here!",
                190,
                200,
                20,
                rl.LightGray)
            break;

        case WorkoutScreen:
            if (len(appState.DataSet.Intervals) > 0) {
                gui.Button(rl.Rectangle{
                    X: float32(rl.GetScreenWidth()) - (10 + 30),
                    Y: 10,
                    Width: 30,
                    Height: 30,
                }, gui.IconText(gui.ICON_GEAR_BIG, ""))

                rl.DrawText(
                    fmt.Sprintf("Target power: %d - %d", appState.CurrentInterval.TargetLow, 
                        appState.CurrentInterval.TargetHigh),
                    10,
                    10,
                    20,
                    rl.Black)

                rl.DrawText(
                    fmt.Sprintf("Interval: %d", appState.CurrentInterval.DurationSeconds),
                    10,
                    30,
                    20,
                    rl.Black)

                rl.DrawText(
                    fmt.Sprintf("Elapsed time: %d", appState.WorkoutElapsedTime),
                    10,
                    50,
                    20,
                    rl.Black)

                appState.drawWorkoutCanvas()
            } else {
                log.Print("Could not read any data from fit file")
                appState.Screen = TitleScreen
            }
            break;

        default:
            break;
        }

		rl.EndDrawing()
	}
}

// NOTE: maybe shoud use DrawingTexture
// to group the whole thing
func (state *AppState) drawWorkoutCanvas() {
    rl.DrawText(fmt.Sprint(state.DataSet.TotalDurationSeconds),
        190,
        200,
        20,
        rl.Green)

    // canvas is always 50% of the screen height
    canvasHeight := float32(rl.GetScreenHeight()) * float32(0.5)
    canvas := renderCanvas(state.DataSet, canvasHeight)

    // makes sure that on window resize 
    // the needle is in the correct poisition
    // Assumes that canvas width is same as window width
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
    rl.DrawRectangleRec(canvas, rl.Black)

    // calculate 1s and 1w pixels
    timeGap := float64(canvas.Width) / float64(data.TotalDurationSeconds)
    powerGap := float64(canvas.Height) / 600 // 600W is max power to display

    needleIncrementX = timeGap
    
    blockX := 0.0
    for _, b := range data.Intervals {
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
        rl.DrawRectangleRec(block, color.RGBA{38, 210, 66, 255})

        lowEndBlock := rl.Rectangle{
            X: float32(blockX),
            Y: canvas.Y + canvas.Height - float32(blockLowEndHeight),
            Height: float32(blockLowEndHeight),
            Width: float32(blockWidth),
        }
        rl.DrawRectangleRec(lowEndBlock, color.RGBA{30, 167, 53, 255})

        blockX = blockX + blockWidth
    }

    // Draw canvas power guide lines
    rl.DrawText("600",
        canvas.ToInt32().X,
        canvas.ToInt32().Y,
        18,
        rl.White)

    rl.DrawText("300",
        canvas.ToInt32().X,
        canvas.ToInt32().Y + (canvas.ToInt32().Height / 2),
        18,
        rl.White)

    rl.DrawLine(
        0,
        canvas.ToInt32().Y + (canvas.ToInt32().Height / 2),
        canvas.ToInt32().Width,
        canvas.ToInt32().Y + (canvas.ToInt32().Height / 2),
        rl.White)

    return canvas
}

// Values needed
// Dataset; Needle pos and needle increment; total duration in seconds
func (state *AppState) GetBlockBasedOnNeedlePos() {
    // gives me the actual second we are on
    state.WorkoutElapsedTime = uint32(needlePosX / needleIncrementX)

    if (len(state.DataSet.Intervals) == 0) {
        return
    }

    // fix: do not index into array every time
    // save the block into some sort of an application state
    state.CurrentInterval = state.DataSet.Intervals[state.CurrentIntervalNumber]
    if (state.CurrentIntervalNumber == 0) {
        state.NextIntervalStartsAt = int(state.CurrentInterval.DurationSeconds)
    }

    if (state.WorkoutElapsedTime > uint32(state.NextIntervalStartsAt)) {
        state.CurrentIntervalNumber += 1

        // TODO: send some sort of a signal
        // save the workout
        if (state.CurrentIntervalNumber == len(state.DataSet.Intervals)) {
            fmt.Println("workout ENDED")
            return;
        }

        nextInterval := state.DataSet.Intervals[state.CurrentIntervalNumber]
        state.NextIntervalStartsAt = int(state.WorkoutElapsedTime) + int(nextInterval.DurationSeconds)
    }
}

