package main

import (
	"fmt"
	"image/color"
	"pedal/internal/fit"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// NOTE: maybe shoud use DrawingTexture
// to group the whole thing
func (state *AppState) drawWorkoutGraph() {
    rl.DrawText(fmt.Sprint(state.DataSet.TotalDurationSeconds),
        190, 200, 20,
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

func renderCanvas(data fit.DataSet, height float32) rl.Rectangle {
    canvasX, canvasY := 0, float32(rl.GetScreenHeight()) - height

    canvas := rl.Rectangle{
        X: float32(canvasX),
        Y: canvasY,
        Height: height,
        Width: float32(rl.GetScreenWidth()),
    }

    // draw canvas element (the parent element)
    rl.DrawRectangleRec(canvas, rl.Black)

    // calculate 1sec and 1w pixels
    timeGap := float64(canvas.Width) / float64(data.TotalDurationSeconds)
    powerGap := float64(canvas.Height) / float64(canvasMaxPowerDisplay)

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

func (state *AppState) moveNeedleBasedOnElapsedTime() {
    needlePosX = float64(state.WorkoutElapsedTime) * needleIncrementX
    needlePosPercent = (needlePosX * 100) / float64(rl.GetScreenWidth())
}

func (state *AppState) setIntervalBasedOnElapsedTime() {
    if (len(state.DataSet.Intervals) == 0) {
        return
    }

    // fix: do not index into array every time
    // save the block into some sort of an application state
    state.CurrentInterval = state.DataSet.Intervals[state.CurrentIntervalNumber]
    if (state.CurrentIntervalNumber == 0) {
        state.NextIntervalStartsAt = int(state.CurrentInterval.DurationSeconds)
    }

    if (state.WorkoutElapsedTime >= uint32(state.NextIntervalStartsAt)) {
        state.CurrentIntervalNumber += 1

        // TODO: send some sort of a signal
        // save the workout
        // change the screen
        // show the completed workout
        if (state.CurrentIntervalNumber == len(state.DataSet.Intervals)) {
            fmt.Println("workout ENDED")
            return;
        }

        nextInterval := state.DataSet.Intervals[state.CurrentIntervalNumber]
        state.NextIntervalStartsAt = int(state.WorkoutElapsedTime) + int(nextInterval.DurationSeconds)
    }
}

// @todo: should have utils file
func convertSecondsToHHMMSS(seconds int32) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	sec := seconds % 60
	
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, sec)
}
