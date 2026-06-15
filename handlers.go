package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func historyHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(userKey).(string)

	rows, err := db.Query(`
		SELECT
			e.id,
			e.created_at,
			MAX(m.stress) as max_stress,
			AVG(m.stress) as avg_stress,
			COUNT(m.id) as total_points
		FROM experiments e
		JOIN measurements m
		ON m.experiment_id = e.id
		WHERE e.user_id = $1
		GROUP BY e.id
		ORDER BY e.created_at DESC
	`, userID)

	if err != nil {
		http.Error(w, "Ошибка получения истории", 500)
		return
	}

	defer rows.Close()

	type Experiment struct {
		ID          int     `json:"id"`
		CreatedAt   string  `json:"created_at"`
		MaxStress   float64 `json:"max_stress"`
		AvgStress   float64 `json:"avg_stress"`
		TotalPoints int     `json:"total_points"`
	}

	var experiments []Experiment

	for rows.Next() {
		var exp Experiment
		err := rows.Scan(
			&exp.ID,
			&exp.CreatedAt,
			&exp.MaxStress,
			&exp.AvgStress,
			&exp.TotalPoints,
		)
		if err != nil {
			continue
		}
		experiments = append(experiments, exp)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data":   experiments,
	})
}

func clearHistoryHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", 405)
		return
	}

	userID := r.Context().Value(userKey).(string)

	_, err := db.Exec(`
        DELETE FROM measurements
        USING experiments
        WHERE measurements.experiment_id = experiments.id
        AND experiments.user_id = $1
    `, userID)

	if err != nil {
		http.Error(w, "Ошибка удаления measurements", 500)
		return
	}

	_, err = db.Exec(`
        DELETE FROM experiments
        WHERE user_id = $1
    `, userID)

	if err != nil {
		http.Error(w, "Ошибка удаления experiments", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// экспорт данных в txt
func exportHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(userKey).(string)

	rows, err := db.Query(`
        SELECT
            e.created_at,
            MAX(m.stress),
            AVG(m.stress),
            COUNT(m.id)
        FROM experiments e
        JOIN measurements m
        ON m.experiment_id = e.id
        WHERE e.user_id = $1
        GROUP BY e.id
    `, userID)

	if err != nil {
		http.Error(w, "Ошибка экспорта", 500)
		return
	}

	defer rows.Close()

	var report strings.Builder

	report.WriteString("ОТЧЁТ ПО ИССЛЕДОВАНИЯМ\n\n")

	for rows.Next() {
		var created string
		var max float64
		var avg float64
		var total int

		rows.Scan(&created, &max, &avg, &total)

		report.WriteString(fmt.Sprintf(`
Дата: %s

Макс. напряжение: %.2f МПа
Среднее напряжение: %.2f МПа
Точек: %d

-----------------------

`,
			created, max, avg, total,
		))
	}

	w.Header().Set("Content-Disposition", "attachment; filename=report.txt")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(report.String()))
}

// загрузка и парсинг файлов
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n НОВЫЙ ЗАПРОС")

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Не получилось распарсить", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "Нет файлов", http.StatusBadRequest)
		return
	}

	finalData := make(map[string][]Measurement)
	var overallAnalysis AnalysisResult

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}

		fileName := filepath.Base(fileHeader.Filename)
		dstPath := filepath.Join("uploads", fileName)
		os.MkdirAll("uploads", os.ModePerm)
		_ = os.WriteFile(dstPath, content, 0644)

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		var points []Measurement

		switch ext {
		case ".txt":
			points = parseTXT(content)
		case ".csv":
			points = parseCSV(content)
		default:
			continue
		}

		if len(points) == 0 {
			fmt.Printf("Файл %s: нет корректных точек\n", fileName)
			continue
		}

		analysis := analyzeMeasurements(points)
		overallAnalysis = analysis

		userID := r.Context().Value(userKey).(string)

		var experimentID int
		queryExp := `
            INSERT INTO experiments (title, material, operator, user_id)
            VALUES ($1, $2, $3, $4)
            RETURNING id
        `
		err = db.QueryRow(queryExp, fileName, "Полимерная пленка", "Admin", userID).Scan(&experimentID)
		if err != nil {
			fmt.Printf("Ошибка добавления эксперимента %s: %v\n", fileName, err)
			continue
		}
		fmt.Printf("Создан эксперимент в БД с ID: %d\n", experimentID)

		tx, err := db.Begin()
		if err != nil {
			fmt.Printf("Ошибка старта транзакции: %v\n", err)
			continue
		}

		stmt, err := tx.Prepare("INSERT INTO measurements (experiment_id, strain, stress, extra_data) VALUES ($1, $2, $3, $4)")
		if err != nil {
			tx.Rollback()
			fmt.Printf("Ошибка Prepare: %v\n", err)
			continue
		}

		insertErr := false
		for _, p := range points {
			extraMap := map[string]string{"step_time": p.StepTime}
			extraJSONBytes, err := json.Marshal(extraMap)
			if err != nil {
				fmt.Printf("Ошибка маршалинга JSON: %v\n", err)
				insertErr = true
				break
			}
			extraJSON := string(extraJSONBytes)

			_, err = stmt.Exec(experimentID, p.Strain, p.Stress, extraJSON)
			if err != nil {
				fmt.Printf("Ошибка записи точки: %v\n", err)
				insertErr = true
				break
			}
		}

		stmt.Close()
		if insertErr {
			tx.Rollback()
			fmt.Printf("Транзакция откачена для эксперимента ID %d\n", experimentID)
		} else {
			err = tx.Commit()
			if err != nil {
				fmt.Printf("Ошибка фиксации транзакции (Commit): %v\n", err)
			} else {
				fmt.Printf("Успешно записано %d точек для эксперимента ID %d\n", len(points), experimentID)
			}
		}

		finalData[fileName] = points
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"payload":  finalData,
		"analysis": overallAnalysis,
	})
}

// возможность сессии на других устройствах
func meHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userKey).(string)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID,
	})
}
