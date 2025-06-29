package main

import (
	"fmt"
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
	InactivityTime uint32
}

const (
    windowHeight = 550
    windowWidth = 1000
    fontSize = 30

    defaultBtnSize float32 = 30
    canvasMaxPowerDisplay int32 = 600

    // Colors

    // Fonts

	visualDebugMode bool = true
)

var (
    needlePosX float64 = 0
    needlePosPercent float64 = 0
    needleIncrementX float64 = 0

    backBtnClicked bool = false
    scanBtnClicked bool = false
    devicesBtnClicked bool = false

	// Saves current workout and resets data
    endWorkoutClicked bool = false
    startWorkoutClicked bool = false

	// TODO: remove.
	// Free ride should be default and if I want to use a workout
	// then just drop a file and that automagically loads the needed stuff.
	// Also drop the title screen
    freeRideBtnClicked bool = false

    scannedDevices []bt.BluetoothDevice = []bt.BluetoothDevice{}
    selectedDeviceIdx int32 = int32(0)
    listViewBounds rl.Rectangle

    currentHeartRate uint8
    currentPower uint16
    currentCadence uint8

    ticker *time.Ticker
    stopTicker chan struct{}

    workoutInProgress bool = false

	// @todo: new fields, should use as an alternative to workoutInProgress probably
	workoutStarted bool = false
	workoutPaused bool = false

	heart_icon_texture rl.Texture2D
	power_icon_texture rl.Texture2D
	time_icon_texture rl.Texture2D

	inactivityTimer *time.Ticker = nil

	visualDebugColor rl.Color = rl.Blank
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
    state.Screen = WorkoutScreen
    state.WorkoutElapsedTime = 0
    state.CurrentIntervalNumber = 0
    state.NextIntervalStartsAt = 0
    state.BluetoothCtl = bt.Init()
    return state
}

func main() {
	if visualDebugMode {
		visualDebugColor = rl.Orange
	}

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

	// @note: pos window in the center of the screen
	currentMonitor := rl.GetCurrentMonitor()
    rl.SetWindowPosition(
        (rl.GetMonitorWidth(currentMonitor) - (windowWidth / 2)),
        (rl.GetMonitorHeight(currentMonitor) / 2) - (windowHeight / 2))

	for !rl.WindowShouldClose() {
        appState.update()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

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
        if (startWorkoutClicked || currentPower > 0) && !workoutInProgress {
            workoutInProgress = true
            ticker = time.NewTicker(1 * time.Second)
            stopTicker = make(chan struct{})

            go func() {
                for {
                    select {
                    case <-ticker.C:
                        state.WorkoutElapsedTime += 1
                        state.moveNeedleBasedOnElapsedTime()
                        state.setIntervalBasedOnElapsedTime()
                    case <-stopTicker:
                        workoutInProgress = false
                        ticker.Stop()

						if inactivityTimer != nil {
							inactivityTimer = nil
						}
                        return
                    }
                }
            }()
        }

		if workoutInProgress && currentPower == 0 && inactivityTimer == nil {
			inactivityTimer = time.NewTicker(1 * time.Second)
            go func() {
                for {
					if inactivityTimer == nil {
						state.InactivityTime = 0
						break
					}

                    select {
                    case <-inactivityTimer.C:
                        state.InactivityTime += 1
					}
                }
            }()
		}

		// @note: pause workout
		if state.InactivityTime >= 5 && workoutInProgress {
			state.InactivityTime = 0
			close(stopTicker)
		}

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
                    log.Println("Connecting to smart trainer")
                    powerChannel := make(chan uint16)

					go state.BluetoothCtl.ConnectToSmartTrainer(selectedDevice.Address, powerChannel)
                    go func() {
                        for {
                            power, ok := <-powerChannel
                            if !ok {
                                break
                            }

                            currentPower = power
                        }
                    }()
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

        /*rl.DrawText(
            fmt.Sprintf("Target power: %d - %d", state.CurrentInterval.TargetLow, 
                state.CurrentInterval.TargetHigh),
            10, 10, 20,
            rl.Black)

        rl.DrawText(
            fmt.Sprintf("Interval: %d", state.CurrentInterval.DurationSeconds),
            10, 30, 20,
            rl.Black)*/

		mainViewPadding := 10
		mainViewWidth := 350

		mainViewRect := rl.Rectangle{
			X: (windowWidth / 2) - (float32(mainViewWidth) / 2),
			Y: (windowHeight / 2) - (450 / 2),
			Width: float32(mainViewWidth),
			Height: 450,
		}
		rl.DrawRectangleRec(mainViewRect, rl.Blue)

		// power
		power_rect := rl.Rectangle{
			X: mainViewRect.X,
			Y: 100,
			Width: (mainViewRect.Width / 2) - 10,
			Height: 70,
		}
		rl.DrawRectangleRec(power_rect, visualDebugColor)
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
			int32(power_rect.X) + 50, (int32(power_rect.Height) / 2) + int32(power_rect.Y) - 20, 40,
			rl.White)

		// avg power
		avgPowerRect := rl.Rectangle{
			X: mainViewRect.X,
			Y: 180,
			Width: (mainViewRect.Width / 2) - 10,
			Height: 70,
		}
		rl.DrawRectangleRec(avgPowerRect, visualDebugColor)
		rl.DrawText(
			"AVG Power",
			int32(avgPowerRect.X) + 10, (int32(avgPowerRect.Height) / 2) + int32(avgPowerRect.Y) - 30, 20,
			rl.White)

		rl.DrawText(
			fmt.Sprintf("%d", 0),
			int32(avgPowerRect.X) + 10, (int32(avgPowerRect.Height) / 2) + int32(avgPowerRect.Y) - 5, 40,
			rl.White)

		// heart rate
		heartRectWidth := (mainViewRect.Width / 2) - float32(mainViewPadding)
		heart_rect := rl.Rectangle{
			X: mainViewRect.X + mainViewRect.Width - heartRectWidth,
			Y: 100,
			Width: heartRectWidth,
			Height: 70,
		}
		rl.DrawRectangleRec(heart_rect, visualDebugColor)

		rl.DrawTextureEx(
			heart_icon_texture,
			rl.NewVector2(
				heart_rect.X + 10,
				(heart_rect.Height / 2) + heart_rect.Y - (float32(heart_icon_texture.Height) * 0.05) / 2),
			0,
			.05,
			rl.White)

		rl.DrawText(
			fmt.Sprintf("%d", currentHeartRate),
			int32(heart_rect.X) + 50, (int32(heart_rect.Height) / 2) + int32(heart_rect.Y) - 20, 40,
			rl.White)

		// avg hr
		avgHeartRateRectWidth := (mainViewRect.Width / 2) - float32(mainViewPadding)
		avgHeartRateRect := rl.Rectangle{
			X: mainViewRect.X + mainViewRect.Width - avgHeartRateRectWidth,
			Y: 180,
			Width: avgHeartRateRectWidth,
			Height: 70,
		}
		rl.DrawRectangleRec(avgHeartRateRect, visualDebugColor)
		rl.DrawText(
			"AVG HR",
			int32(avgHeartRateRect.X) + 10, (int32(avgHeartRateRect.Height) / 2) + int32(avgHeartRateRect.Y) - 30, 20,
			rl.White)

		rl.DrawText(
			fmt.Sprintf("%d", 0),
			int32(avgHeartRateRect.X) + 10, (int32(avgHeartRateRect.Height) / 2) + int32(avgHeartRateRect.Y) - 5, 40,
			rl.White)

		// ride time data
		ride_time_rect := rl.Rectangle{
			X: mainViewRect.X,
			Y: 260,
			Width: mainViewRect.Width,
			Height: 70,
		}
		rl.DrawRectangleRec(ride_time_rect, visualDebugColor)
		rl.DrawTextureEx(
			time_icon_texture,
			rl.NewVector2(
				ride_time_rect.X + 10,
				(ride_time_rect.Height / 2) + ride_time_rect.Y - (float32(time_icon_texture.Height) / 2)),
			0,
			1,
			rl.White)

		rl.DrawText(
			convertSecondsToHHMMSS(int32(state.WorkoutElapsedTime)),
			int32(ride_time_rect.X) + 50, (int32(ride_time_rect.Height) / 2) + int32(ride_time_rect.Y) - 20, 40,
			rl.White)
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

