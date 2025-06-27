package main

import (
	"fmt"
	"image/color"
	"log"
	bt "pedal/internal/bluetoothctl"
	"pedal/internal/fit"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type AppState struct {
    DataSet fit.DataSet
    Screen ApplicationScreen
    WorkoutElapsedTime uint32
    CurrentInterval fit.Interval
    CurrentIntervalNumber int
    NextIntervalStartsAt int
    BluetoothCtl bt.BluetoothControl
}

const (
    windowHeight = 450
    windowWidth = 800
    fontSize = 30

    defaultBtnSize float32 = 30
    canvasMaxPowerDisplay int32 = 600

    // Colors

    // Fonts
)

var (
    needlePosX float64 = 0
    needlePosPercent float64 = 0
    needleIncrementX float64 = 0

    backBtnClicked bool = false
    scanBtnClicked bool = false
    devicesBtnClicked bool = false
    endWorkoutClicked bool = false
    startWorkoutClicked bool = false
    freeRideBtnClicked bool = false

    scannedDevices []bt.BluetoothDevice = []bt.BluetoothDevice{}
    selectedDeviceIdx int32 = int32(0)
    listViewBounds rl.Rectangle

    currentHeartRate uint8
    currentPower uint8
    currentCadence uint8

    ticker *time.Ticker
    stopTicker chan struct{}

    workoutInProgress bool = false

	heart_icon_texture rl.Texture2D
	power_icon_texture rl.Texture2D
	time_icon_texture rl.Texture2D
)

type ApplicationScreen int
const (
    TitleScreen ApplicationScreen = iota
    WorkoutScreen
    SettingsScreen
    DevicesScreen
    WorkoutCompletedScreen
)

func initApp() (state AppState) {
    state.Screen = TitleScreen
    state.WorkoutElapsedTime = 0
    state.CurrentIntervalNumber = 0
    state.NextIntervalStartsAt = 0
    state.BluetoothCtl = bt.Init()
    return state
}

func main() {
    appState := initApp()

	rl.InitWindow(windowWidth, windowHeight, "Pedal")

	heart_icon := rl.LoadImage("./assets/heart.png")
	heart_icon_texture = rl.LoadTextureFromImage(heart_icon)
	rl.UnloadImage(heart_icon)

	power_icon := rl.LoadImage("./assets/power.png")
	power_icon_texture = rl.LoadTextureFromImage(power_icon)
	rl.UnloadImage(power_icon)

	time_icon := rl.LoadImage("./assets/clock.png")
	time_icon_texture = rl.LoadTextureFromImage(time_icon)
	rl.UnloadImage(time_icon)

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

    /*rl.SetWindowPosition(
        (rl.GetMonitorWidth(0) - (windowWidth / 2)),
        (rl.GetMonitorHeight(0) / 2) - (windowHeight / 2))*/

	for !rl.WindowShouldClose() {
        appState.update()

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

        appState.draw()

		rl.EndDrawing()
	}

	rl.UnloadTexture(heart_icon_texture)
	rl.UnloadTexture(power_icon_texture)
	rl.UnloadTexture(time_icon_texture)
}

func (state *AppState) update() {
    //================ Title screen =================
    if (state.Screen == TitleScreen) {
        droppedFile := make([]string, 0)

        if (rl.IsFileDropped()) {
            droppedFile = rl.LoadDroppedFiles()

            if (len(droppedFile) > 0) {
                state.DataSet = fit.ParseWorkoutFile(droppedFile[0])
                rl.UnloadDroppedFiles()

                state.Screen = WorkoutScreen
            }
        }

        if freeRideBtnClicked {
            state.Screen = WorkoutScreen
            return
        }

        return
    }

    //================ Workout screen =================
    if state.Screen == WorkoutScreen {
        if startWorkoutClicked && !workoutInProgress {
            workoutInProgress = true
            ticker = time.NewTicker(1 * time.Second)
            stopTicker = make(chan struct{})

            go func() {
                for {
                    select {
                    // TODO: on each tick read
                    // timestamp; power;hr;cadence
                    // and write to DB (could use a queue probably)
                    // TODO later: Also display hr and power and cadence on canvas
                    case <-ticker.C:
                        state.WorkoutElapsedTime += 1
                        state.moveNeedleBasedOnElapsedTime()
                        state.setIntervalBasedOnElapsedTime()
                    case <-stopTicker:
                        workoutInProgress = false
                        ticker.Stop()
                        return
                    }
                }
            }()
        }

        if endWorkoutClicked && workoutInProgress {
            close(stopTicker)
        }
        /*if rl.IsKeyDown(rl.KeyRight) && state.WorkoutElapsedTime < uint32(state.DataSet.TotalDurationSeconds) {
            state.WorkoutElapsedTime += 1
            state.moveNeedleBasedOnElapsedTime()
            state.setIntervalBasedOnElapsedTime()
            return
        }*/

        if devicesBtnClicked {
            state.Screen = DevicesScreen
            return
        }

        return
    }

    //================= Devices screen =================
    if state.Screen == DevicesScreen {
        if (backBtnClicked) {
            state.Screen = WorkoutScreen
            return;
        }

        // Scan devices btn click ---------------------------
        if scanBtnClicked {
            scannedDevices = []bt.BluetoothDevice{}

            ch := make(chan bt.BluetoothDevice)
            go state.BluetoothCtl.Scan(ch, []string{})
            go func() {
                for {
                    bltDevice, ok := <-ch
                    if !ok {
                        break
                    }

                    if len(scannedDevices) == 0 {
                        scannedDevices = append(scannedDevices, bltDevice)
                    } else {
                        exists := false
                        for _, device := range scannedDevices {
                            if device.Address == bltDevice.Address {
                                exists = true
                                break;
                            }
                        }

                        if !exists {
                            scannedDevices = append(scannedDevices, bltDevice)
                        }
                    }
                }
            }()
        }

        // Connect to device click ------------------------
        mousePos := rl.GetMousePosition()
        if rl.CheckCollisionPointRec(mousePos, listViewBounds) {
            if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
                if len(scannedDevices) == 0 {
                    return
                }

				if selectedDeviceIdx < 0 {
					return
				}

                if len(scannedDevices) < int(selectedDeviceIdx) {
                    log.Printf("how did i get here. Selected index: %d", selectedDeviceIdx)
                    return;
                }
                selectedDevice := scannedDevices[selectedDeviceIdx]

                if selectedDevice.Type == bt.HeartRateMonitor && !state.BluetoothCtl.HrMonitorConnected {
                    hrMonitorCh := make(chan uint8)
                    go state.BluetoothCtl.ConnectToHrMonitor(selectedDevice.Address, hrMonitorCh)
                    go func() {
                        for {
                            hrValue, ok := <-hrMonitorCh
                            if !ok {
                                break
                            }

                            log.Printf("HR: %d", hrValue)
                            currentHeartRate = hrValue
                        }
                    }()
                }
 
                if selectedDevice.Type == bt.SmartTrainer && !state.BluetoothCtl.SmartTrainerConnected {
                    //TODO: listen smart trainer power and cadence
                    log.Println("Connecting to smart trainer")
                    powerChannel := make(chan uint8)
                    go state.BluetoothCtl.ConnectToSmartTrainer(selectedDevice.Address, powerChannel)
                }
            }
        }
    }
}

func (state *AppState) draw() {
    if state.Screen == TitleScreen {
        /* Free ride button */
        freeRideBtnClicked = gui.Button(rl.Rectangle{
            X: float32(rl.GetScreenWidth()) - (20 + 100),
            Y: 10,
            Width: 100,
            Height: defaultBtnSize,
        }, "Free ride")

        rl.DrawText("DROP .FIT FILE HERE",
            190, 200, 20,
            rl.LightGray)

        return
    }

	// @todo: make a proper grid
    if state.Screen == WorkoutScreen {
        if len(state.DataSet.Intervals) > 0 {
            state.drawWorkoutGraph()
        }

        // Settings button
        gui.Button(rl.Rectangle{
            X: float32(rl.GetScreenWidth()) - (10 + defaultBtnSize),
            Y: 10,
            Width: defaultBtnSize,
            Height: defaultBtnSize,
        }, gui.IconText(gui.ICON_GEAR_BIG, ""))

        // Devices button
        devicesBtnClicked = gui.Button(rl.Rectangle{
            X: float32(rl.GetScreenWidth()) - (50 + defaultBtnSize),
            Y: 10,
            Width: defaultBtnSize,
            Height: defaultBtnSize,
        }, gui.IconText(gui.ICON_TOOLS, ""))


        endWorkoutClicked = gui.Button(rl.Rectangle{
            X: float32(rl.GetScreenWidth()) - 195,
            Y: 10,
            Width: 100,
            Height: defaultBtnSize,
        }, "End workout")

        startWorkoutClicked = gui.Button(rl.Rectangle{
            X: float32(rl.GetScreenWidth()) - 310,
            Y: 10,
            Width: 100,
            Height: defaultBtnSize,
        }, "Start workout")

        rl.DrawText(
            fmt.Sprintf("Target power: %d - %d", state.CurrentInterval.TargetLow, 
                state.CurrentInterval.TargetHigh),
            10, 10, 20,
            rl.Black)

        rl.DrawText(
            fmt.Sprintf("Interval: %d", state.CurrentInterval.DurationSeconds),
            10, 30, 20,
            rl.Black)

		// heart rate
		heart_rect := rl.Rectangle{
			X: 220,
			Y: 100,
			Width: 100,
			Height: 50,
		}
		rl.DrawRectangleRec(heart_rect, rl.Gray)

		rl.DrawTextureEx(
			heart_icon_texture,
			rl.NewVector2(
				heart_rect.X + 10,
				(heart_rect.Height / 2) + heart_rect.Y - (float32(heart_icon_texture.Height) * 0.05) / 2),
			0,
			.05,
			rl.Black)

		rl.DrawText(
			fmt.Sprintf("%d", currentHeartRate),
			int32(heart_rect.X) + 50, (int32(heart_rect.Height) / 2) + int32(heart_rect.Y) - 10, 20,
			rl.Black)

		// power
		power_rect := rl.Rectangle{
			X: 100,
			Y: 100,
			Width: 100,
			Height: 50,
		}
		rl.DrawRectangleRec(power_rect, rl.Gray)
		rl.DrawTextureEx(
			power_icon_texture,
			rl.NewVector2(
				power_rect.X + 10,
				(power_rect.Height / 2) + power_rect.Y - (float32(power_icon_texture.Height) / 2)),
			0,
			1,
			rl.Black)

		rl.DrawText(
			fmt.Sprintf("%d", currentPower),
			int32(power_rect.X) + 50, (int32(power_rect.Height) / 2) + int32(power_rect.Y) - 10, 20,
			rl.Black)

		// ride time data
		ride_time_rect := rl.Rectangle{
			X: 100,
			Y: 170,
			Width: 220,
			Height: 50,
		}
		rl.DrawRectangleRec(ride_time_rect, rl.Gray)
		rl.DrawTextureEx(
			time_icon_texture,
			rl.NewVector2(
				ride_time_rect.X + 10,
				(power_rect.Height / 2) + ride_time_rect.Y - (float32(time_icon_texture.Height) / 2)),
			0,
			1,
			rl.Black)

		rl.DrawText(
			convertSecondsToHHMMSS(int32(state.WorkoutElapsedTime)),
			int32(ride_time_rect.X) + 50, (int32(ride_time_rect.Height) / 2) + int32(ride_time_rect.Y) - 10, 20,
			rl.Black)
    }

    if (state.Screen == DevicesScreen) {
        backBtnClicked = gui.Button(rl.Rectangle{
            X: 10,
            Y: 10,
            Width: defaultBtnSize,
            Height: defaultBtnSize,
        }, gui.IconText(gui.ICON_ARROW_LEFT, ""))

        scanBtnClicked = gui.Button(rl.Rectangle{
            X: 50,
            Y: 10,
            Width: 100,
            Height: defaultBtnSize,
        }, "Scan devices")


        listViewBounds = rl.Rectangle{
            X: (float32(rl.GetScreenWidth()) / 2) - 200,
            Y: 10,
            Height: float32(rl.GetScreenHeight()) - 20,
            Width: 400,
        }

        selectedDeviceIdx = gui.ListViewEx(
            listViewBounds,
            bt.ListToString(&scannedDevices),
            nil,
            nil,
            selectedDeviceIdx)
    }
}

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
