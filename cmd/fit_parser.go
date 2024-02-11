package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tormoder/fit"
)

type Data struct {
    StepIdx int
    DurationSeconds float64
    TargetLow int32
    TargetHigh int32
}

func ParseWorkoutFile(fullFilePath string) []*fit.WorkoutStepMsg {
    file := filepath.Join(fullFilePath)

    data, err := os.ReadFile(file)
    if err != nil {
        fmt.Println(err)
        return nil
    }

    fit, err := fit.Decode(bytes.NewReader(data))
    if err != nil {
        fmt.Println(err)
        return nil
    }

    workoutFile, err := fit.Workout()
    if err != nil {
        fmt.Println(err)
        return nil
    }


    // laps
    for _, step := range workoutFile.WorkoutSteps {
        //fmt.Printf("%+v\n", step)

        if step.DurationType.String() == "RepeatUntilStepsCmplt" {
            fmt.Printf("%s, from step: %d; times: %d \n",
                step.DurationType.String(),
                step.DurationValue,
                step.TargetValue)
        }

        if step.DurationType.String() == "Time" {
            powerHighInter := step.CustomTargetValueHigh - 1000
            powerLowInter := step.CustomTargetValueLow - 1000

            fmt.Printf("Step duration: %f s ", step.GetDurationValue())
            fmt.Printf("%s => %d - %d \n",
                step.MessageIndex,
                powerLowInter,
                powerHighInter)

        }
    }


    return workoutFile.WorkoutSteps
}
