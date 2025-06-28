package bluetoothctl

import (
	"log"
	"time"
	"encoding/binary"

	"tinygo.org/x/bluetooth"
)

type BluetoothControl struct {
    Adapter *bluetooth.Adapter
    HrMonitorConnected bool
    SmartTrainerConnected bool
}

type BluetoothDevice struct {
    Name string
    Address bluetooth.Address
    Type DeviceType
}

type DeviceType int
const (
    Unknown DeviceType = iota
    HeartRateMonitor
    SmartTrainer
)

const (
    scanDuration = 10 * time.Second
)

var (
    heartRateServiceUUID = bluetooth.ServiceUUIDHeartRate
    heartRateCharacteristicUUID = bluetooth.CharacteristicUUIDHeartRateMeasurement

    cyclingPowerServiceUUID = bluetooth.ServiceUUIDCyclingPower
    cyclingPowerCharacteristicUUID = bluetooth.CharacteristicUUIDCyclingPowerMeasurement
    cyclingPowerFeatureCharacteristicUUID = bluetooth.CharacteristicUUIDCyclingPowerFeature
    cyclingPowerVectorCharacteristicUUID = bluetooth.CharacteristicUUIDCyclingPowerVector

	indoorBikeData = bluetooth.CharacteristicUUIDIndoorBikeData

    cyclingSpeedAndCadenceServiceUUID = bluetooth.ServiceUUIDCyclingSpeedAndCadence
)

func Init() (bt BluetoothControl) {
    bt.Adapter = bluetooth.DefaultAdapter
    return bt
}

// if devices list is not empty, scan until devices found and connected
// otherwise scan for 10s
func (bt *BluetoothControl) Scan(ch chan BluetoothDevice, devices []string) {
    err := bt.Adapter.Enable()
    if err != nil {
        log.Printf("Could not enable bluetooth adapter %s", err)
        return;
    }

    log.Println("Start scanning")

    endScanTime := time.Now().Add(scanDuration)
    err = bt.Adapter.Scan(func(a *bluetooth.Adapter, sr bluetooth.ScanResult) {
        advPayload := sr.AdvertisementPayload
        if advPayload.HasServiceUUID(bluetooth.New16BitUUID(heartRateServiceUUID.Get16Bit())) {
            ch <- BluetoothDevice{
                Name: sr.LocalName(),
                Address: sr.Address,
                Type: HeartRateMonitor,
            }
        }

        if advPayload.HasServiceUUID(bluetooth.New16BitUUID(cyclingPowerServiceUUID.Get16Bit())) {
            ch <- BluetoothDevice{
                Name: sr.LocalName(),
                Address: sr.Address,
                Type: SmartTrainer,
            }
        }

        if (time.Now().After(endScanTime)) {
            bt.Adapter.StopScan()
            log.Println("Scan ended")

            close(ch)
        }
    })

    if err != nil {
        log.Printf("Could not start the scan %s", err)
        return;
    }
}

func (bt *BluetoothControl) ConnectToHrMonitor(deviceAddress bluetooth.Address, ch chan uint8) {
    device, err := bt.Adapter.Connect(deviceAddress, bluetooth.ConnectionParams{})
    if err != nil {
        log.Printf("Could not connect to device. Error: %s \n", err)
        close(ch)
        return;
    }

    // note: not sure if this is in the correct place
    // or is it even needed
    bt.Adapter.StopScan()

    services, err := device.DiscoverServices([]bluetooth.UUID{heartRateServiceUUID})
    if err != nil {
        log.Printf("Failed to discover services. Error: %s \n", err)
        device.Disconnect()
        close(ch)
        return
    }

    if len(services) == 0 {
        log.Println("Could not find any services")
        device.Disconnect()
        close(ch)
        return
    }

    service := services[0]
    characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{heartRateCharacteristicUUID})
    if err != nil {
        log.Printf("Failed to discover characteristics. Error: %s \n", err)
        device.Disconnect()
        close(ch)
        return
    }

    if len(characteristics) == 0 {
        log.Println("Could not find any characteristics")
        device.Disconnect()
        close(ch)
        return
    }

    bt.HrMonitorConnected = true

    characteristic := characteristics[0]
    log.Printf("Found characteristic %s", characteristic.UUID().String())

    err = characteristic.EnableNotifications(func(buf []byte) {
        value := uint8(buf[1])
        ch <-value
    })

    if err != nil {
        bt.HrMonitorConnected = false
        device.Disconnect()

        log.Printf("Error reading value: %s", err)
        close(ch)
    }
}

// @todo: should return  error if failed to connect
// also, what happens if connection is lost?
func (bt *BluetoothControl) ConnectToSmartTrainer(deviceAddress bluetooth.Address, ch chan uint16) {
    bt.Adapter.StopScan();

    device, err := bt.Adapter.Connect(deviceAddress, bluetooth.ConnectionParams{})
    if err != nil {
        log.Printf("Could not connect to smart trainer: %s", err)
        close(ch)
        return
    }

    services, err := device.DiscoverServices([]bluetooth.UUID{cyclingPowerServiceUUID})
    if err != nil {
        log.Printf("Failed to discover services. Error: %s \n", err)
        device.Disconnect()
        close(ch)
        return
    }

    service := services[0]
    characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{cyclingPowerCharacteristicUUID})
    if err != nil {
        log.Printf("Failed to discover characteristics. Error: %s \n", err)
        device.Disconnect()
        close(ch)
        return
    }

    if len(characteristics) == 0 {
        log.Println("Could not find any characteristics")
        device.Disconnect()
        close(ch)
        return
    }

    characteristic := characteristics[0]
    log.Printf("Found characteristics %v", characteristic)

	offset := 2
    err = characteristic.EnableNotifications(func(buf []byte) {
        //log.Printf("Power measurement: %d", buf)

		byteSlice := buf[offset : offset + 2]
		integerValue := binary.LittleEndian.Uint16(byteSlice)
        //log.Printf("Power: %d", integerValue)

        ch <-integerValue
    })

    if err != nil {
        //bt.HrMonitorConnected = false
        device.Disconnect()

        log.Printf("Error reading value: %s", err)
        close(ch)
    }

}

