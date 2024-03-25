package bluetoothctl

import "fmt"

func (device *BluetoothDevice) ToString() string {
    return fmt.Sprintf("%s - %s", device.Name, device.Address.String())
}

func ListToString(devices *[]BluetoothDevice) []string {
    result := []string{}
    for _, device := range *devices {
        result = append(result, device.ToString())
    }

    return result
}
