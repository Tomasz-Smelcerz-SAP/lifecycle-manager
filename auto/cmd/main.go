package main

import (
	"errors"
	"fmt"

	"github.com/kyma-project/lifecycle-manager/auto"
)

var ErrInvalidReason = errors.New("invalid reason")

func main() {
	ok, err := doBusinessLogic("exaample")
	// ok, err := doBusinessLogic("")
	if err != nil {
		if errors.Is(err, ErrDivideByZero) {
			logger("Division by zero occurred")
		}
		if errors.Is(err, auto.ErrWrapped) {
			logger("AutoWrapped error occurred: ")
			inst := &auto.WrappedError{}
			if errors.As(err, &inst) {
				logger(inst.String())
			}
		}
		logger("\nError:", err)
		return
	}

	if ok {
		logger("Success")
	} else {
		logger("Failure")
	}
}

func logger(args ...any) {
	fmt.Println(args...) //nolint:forbidigo // demo
}

func doBusinessLogic(reason string) (bool, error) {
	if reason == "" {
		return false, auto.Wrap(ErrInvalidReason)
	}
	objects := make([]int, len(reason))
	for idx, r := range reason {
		objects[idx] = int(r)
	}

	arranged, err := doLowLevelStuff(objects)
	if err != nil {
		return false, fmt.Errorf("failed to arrange objects: %w", err)
	}

	prev := arranged[0]
	for idx := 1; idx < len(arranged); idx++ {
		if arranged[idx] < prev {
			return false, nil
		}
	}
	return true, nil
}

func doLowLevelStuff(objects []int) ([]int, error) {
	res, err := arrangeObjects(objects)
	if err != nil {
		return nil, auto.Wrap(err)
	}
	return res, nil
}

func arrangeObjects(objects []int) ([]int, error) {
	if len(objects) < 2 { //nolint:mnd // demo
		return nil, errors.New("not enough objects to arrange") //nolint:err113 // demo
	}

	res := []int{}
	for idx := 1; idx < len(objects); idx++ {
		current := objects[idx]
		previous := objects[idx-1]
		mathRes, err := doMath(current, previous)
		if err != nil {
			return nil, auto.Wrap(err)
		}
		res = append(res, mathRes)
	}
	return res, nil
}

func doMath(a, b int) (int, error) {
	res, err := Divide(a, b)
	if err != nil {
		return 0, fmt.Errorf("cannot divide: %w", err)
	}
	return int(res), nil
}
