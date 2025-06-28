package main

import (
	"time"
	"fmt"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (state *AppState) WorkoutScreenUpdate() {
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

func (state *AppState) WorkoutScreenDraw() {
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
		rl.Black,
	)

	rl.DrawText(
		fmt.Sprintf("Interval: %d", state.CurrentInterval.DurationSeconds),
		10, 30, 20,
		rl.Black,
	)

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
			(heart_rect.Height / 2) + heart_rect.Y - (float32(heart_icon_texture.Height) * 0.05) / 2,
		),
		0,
		.05,
		rl.Black,
	)

	rl.DrawText(
		fmt.Sprintf("%d", currentHeartRate),
		int32(heart_rect.X) + 50, (int32(heart_rect.Height) / 2) + int32(heart_rect.Y) - 10, 20,
		rl.Black,
	)


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
			(power_rect.Height / 2) + power_rect.Y - (float32(power_icon_texture.Height) / 2),
		),
		0,
		1,
		rl.Black,
	)

	rl.DrawText(
		fmt.Sprintf("%d", currentPower),
		int32(power_rect.X) + 50, (int32(power_rect.Height) / 2) + int32(power_rect.Y) - 10, 20,
		rl.Black,
	)

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
			(power_rect.Height / 2) + ride_time_rect.Y - (float32(time_icon_texture.Height) / 2),
		),
		0,
		1,
		rl.Black,
	)

	rl.DrawText(
		convertSecondsToHHMMSS(int32(state.WorkoutElapsedTime)),
		int32(ride_time_rect.X) + 50, (int32(ride_time_rect.Height) / 2) + int32(ride_time_rect.Y) - 10, 20,
		rl.Black,
	)
}
