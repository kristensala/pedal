package cmd

import (
	"fmt"
	"log"
	"time"

	"tinygo.org/x/bluetooth"
)

type BluetoothControl struct {
    Adapter *bluetooth.Adapter
    HrMonitorConnected bool
    SmartTrainerConnected bool
}

type BluetoothDevice struct {
    Name string
    Address string
}

var (
    hearRateServiceUUID = bluetooth.ServiceUUIDHeartRate
    hearRateCharacteristicUUID = bluetooth.CharacteristicUUIDHeartRateMeasurement

    cyclingPowerServiceUUID = bluetooth.ServiceUUIDCyclingPower
    cyclingPowerCharacteristicUUID = bluetooth.CharacteristicUUIDCyclingPowerMeasurement
)

func (device *BluetoothDevice) ToString() string {
    return fmt.Sprintf("%s - %s", device.Name, device.Address)
}

func InitBt() (bt BluetoothControl) {
    bt.Adapter = bluetooth.DefaultAdapter
    return bt
}

func (bt *BluetoothControl) Scan(ch chan BluetoothDevice) {
    err := bt.Adapter.Enable()
    if err != nil {
        log.Printf("Could not enable bluetooth adapter %s", err)
        return;
    }

    log.Println("Start scanning")

    scanDuration := 10 * time.Second
    endScanTime := time.Now().Add(scanDuration)
    err = bt.Adapter.Scan(func(a *bluetooth.Adapter, sr bluetooth.ScanResult) {

        // bluetooth headphone testing
        // todo: show only valid devices (hr monitor and smart trainers)
        // do not allow two types on same device connection (no 2 HR monitors)
        advPayload := sr.AdvertisementPayload
        if advPayload.HasServiceUUID(bluetooth.New16BitUUID(0x111E)) {
            log.Printf("%s", sr.LocalName())
        }

        ch <- BluetoothDevice{
            Name: sr.LocalName(),
            Address: sr.Address.String(),
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

// note: currently only heart rate monitor
// fix: add  param which is the device to detect the type of device
// and possibly block, when connecting is not allowed (ex: type of device already connected)
func (bt *BluetoothControl) Connect(deviceAddress bluetooth.Address, ch chan uint8) {
    device, err := bt.Adapter.Connect(deviceAddress, bluetooth.ConnectionParams{})
    if err != nil {
        log.Panicf("Could not connect to device. Error: %s \n", err)
        return;
    }

    services, err := device.DiscoverServices([]bluetooth.UUID{hearRateServiceUUID})
    if err != nil {
        log.Panicf("Failed to discover services. Error: %s \n", err)
        device.Disconnect()
        return
    }

    if len(services) == 0 {
        log.Panicln("Could not find any services")
        device.Disconnect()
        return
    }

    service := services[0]
    characteristics, err := service.DiscoverCharacteristics([]bluetooth.UUID{hearRateCharacteristicUUID})
    if err != nil {
        log.Panicf("Failed to discover characteristics. Error: %s \n", err)
        device.Disconnect()
        return
    }

    if len(characteristics) == 0 {
        log.Panicln("Could not find any characteristics")
        device.Disconnect()
        return
    }

    characteristic := characteristics[0]
    log.Printf("Found characteristic %s", characteristic.UUID().String())

    characteristic.EnableNotifications(func(buf []byte) {
        value := uint8(buf[1])
        ch <-value
    })

    select{}
}

