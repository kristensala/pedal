package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tormoder/fit"
)

type DataSet struct {
    TotalDurationSeconds int
    ThresholdPower uint32
    Intervals []Interval
}

type Interval struct {
    Number uint16
    DurationSeconds uint32
    TargetLow uint32
    TargetHigh uint32
}

func ParseWorkoutFile(fullFilePath string) DataSet {
    dataSet := DataSet{}

    file := filepath.Join(fullFilePath)

    data, err := os.ReadFile(file)
    if err != nil {
        fmt.Println(err)
        return dataSet
    }

    fit, err := fit.Decode(bytes.NewReader(data))
    if err != nil {
        fmt.Println(err)
        return dataSet
    }

    workoutFile, err := fit.Workout()
    if err != nil {
        fmt.Println(err)
        return dataSet
    }


    result := buildBlocks(workoutFile.WorkoutSteps)
    totalDuration := getTotalDurationInSeconds(result)

    dataSet = DataSet{
        TotalDurationSeconds: totalDuration,
        Intervals: result,
        ThresholdPower: 200,
    }

    /*resultJson, _ := json.MarshalIndent(dataSet, ""," ")
    fmt.Println(string(resultJson))*/

    return dataSet
}

func buildBlocks(messages []*fit.WorkoutStepMsg) []Interval {
    steps := []Interval{}

    for _, stepMsg := range messages {
        if stepMsg.DurationType.String() == "Time" {
            duration := stepMsg.DurationValue / 1000
            powerHigh := stepMsg.CustomTargetValueHigh - 1000
            powerLow := stepMsg.CustomTargetValueLow - 1000

            newStep := Interval{
                Number: uint16(stepMsg.MessageIndex),
                DurationSeconds: duration,
                TargetLow: powerLow,
                TargetHigh: powerHigh,
            }

            steps = append(steps, newStep)
        }

        if stepMsg.DurationType.String() == "RepeatUntilStepsCmplt" {
            repeatFrom := stepMsg.DurationValue
            repeatTimes := stepMsg.TargetValue - 1 // minus one because one entry is already in array

            sliceToRepeat := steps[repeatFrom:]

            for i := 0; i < int(repeatTimes); i++ {
                steps = append(steps, sliceToRepeat...)
            }
        }
    }

    return steps
}

func getTotalDurationInSeconds(blocks []Interval) int {
    totalDuration := 0

    for _, b := range blocks {
        totalDuration = totalDuration + int(b.DurationSeconds)
    }

    return totalDuration
}

