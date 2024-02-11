package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tormoder/fit"
)

type Step struct {
    Number uint16
    DurationSeconds uint32
    TargetLow uint32
    TargetHigh uint32
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
            fmt.Printf("%d => %d - %d \n",
                uint16(step.MessageIndex),
                powerLowInter,
                powerHighInter)

        }
    }


    return workoutFile.WorkoutSteps
}

func buildSteps(messages []*fit.WorkoutStepMsg) []Step{
    steps := []Step{}

    for _, stepMsg := range messages {
        if stepMsg.DurationType.String() == "Time" {
            duration := stepMsg.DurationValue
            powerHigh := stepMsg.CustomTargetValueHigh - 1000
            powerLow := stepMsg.CustomTargetValueLow - 1000

            newStep := Step{
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

