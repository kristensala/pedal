package cmd

import (
	"log"
	"time"

	"tinygo.org/x/bluetooth"
)

type BluetoothControl struct {
    Adapter *bluetooth.Adapter
}

func InitBt() (bt BluetoothControl) {
    bt.Adapter = bluetooth.DefaultAdapter
    return bt
}

func (bt *BluetoothControl) Scan(ch chan string) {
    err := bt.Adapter.Enable()
    if err != nil {
        log.Printf("Could not enable bluetooth adapter %s", err)
        return;
    }

    log.Println("Start scanning")

    scanDuration := 10 * time.Second
    endScanTime := time.Now().Add(scanDuration)
    err = bt.Adapter.Scan(func(a *bluetooth.Adapter, sr bluetooth.ScanResult) {
        //fmt.Printf("Device: %s; %s \n", sr.Address.String(), sr.LocalName())

        ch <-sr.LocalName()

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

func (bt *BluetoothControl) Connect(device string) {

}

func (bt *BluetoothControl) Listen() {

}

