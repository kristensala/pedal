package main

import (
	"log"
	bt "pedal/internal/bluetoothctl"
	rl "github.com/gen2brain/raylib-go/raylib"
	gui "github.com/gen2brain/raylib-go/raygui"
)

func (state *AppState) DevicesScreenUpdate() {
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

func (state *AppState) DevicesScreenDraw() {
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
