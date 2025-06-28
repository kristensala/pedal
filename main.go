package main

import (
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

    if state.Screen == WorkoutScreen {
		state.WorkoutScreenUpdate()
    }

    if state.Screen == DevicesScreen {
		state.DevicesScreenUpdate()
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

    if state.Screen == WorkoutScreen {
		state.WorkoutScreenDraw()
    }

    if (state.Screen == DevicesScreen) {
		state.DevicesScreenDraw()
    }
}

