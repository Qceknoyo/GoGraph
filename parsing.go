package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"strconv"
	"strings"
)

type Measurement struct {
	StepTime string  `json:"step_time"`
	Strain   float64 `json:"strain"`
	Stress   float64 `json:"stress"`
}

type AnalysisResult struct {
	MaxStress   float64 `json:"max_stress"`
	AvgStress   float64 `json:"avg_stress"`
	BreakPoint  float64 `json:"break_point"`
	TotalPoints int     `json:"total_points"`
}

// функция анализа данных
func analyzeMeasurements(points []Measurement) AnalysisResult {

	var maxStress float64
	var totalStress float64
	var breakPoint float64

	for i, p := range points {

		if p.Stress > maxStress {
			maxStress = p.Stress
		}

		totalStress += p.Stress

		if i > 0 {
			prev := points[i-1].Stress
			if prev-p.Stress > 5 {
				breakPoint = p.Strain
			}
		}
	}

	avgStress := totalStress / float64(len(points))

	return AnalysisResult{
		MaxStress:   maxStress,
		AvgStress:   avgStress,
		BreakPoint:  breakPoint,
		TotalPoints: len(points),
	}
}

func parseTXT(content []byte) []Measurement {
	lines := strings.Split(string(content), "\n")
	var measurements []Measurement

	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		stepTime := fields[0]
		strain, err1 := strconv.ParseFloat(fields[1], 64)
		stress, err2 := strconv.ParseFloat(fields[2], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		measurements = append(measurements, Measurement{
			StepTime: stepTime,
			Strain:   strain,
			Stress:   stress,
		})
	}
	return measurements
}

func parseCSV(content []byte) []Measurement {
	r := csv.NewReader(bytes.NewReader(content))
	r.Comma = ';'

	if _, err := r.Read(); err != nil {
		return nil
	}

	var measurements []Measurement
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 3 {
			continue
		}
		stepTime := record[0]
		strain, err1 := strconv.ParseFloat(record[1], 64)
		stress, err2 := strconv.ParseFloat(record[2], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		measurements = append(measurements, Measurement{
			StepTime: stepTime,
			Strain:   strain,
			Stress:   stress,
		})
	}
	return measurements
}
