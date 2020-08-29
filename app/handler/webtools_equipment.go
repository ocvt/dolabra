package handler

import (
	"encoding/json"
	"net/http"
)

type equipmentStruct struct {
	/* Managed server side */
	CreateDatetime string `json:"createDatetime,omitempty"`
	/* Required for adding equipment */
	// Id & Count only for PATCH, included in URL param
	Id          int    `json:"id,omitempty"`
	Count       int    `json:"count"`
	Description string `json:"description"`
}

func GetWebtoolsEquipment(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT * FROM equipment
		ORDER BY datetime(create_datetime) DESC`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var equipment = []*equipmentStruct{}
	i := 0
	for rows.Next() {
		equipment = append(equipment, &equipmentStruct{})
		err = rows.Scan(
			&equipment[i].Id,
			&equipment[i].CreateDatetime,
			&equipment[i].Description,
			&equipment[i].Count)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*equipmentStruct{"equipment": equipment})
}

func PatchWebtoolsEquipment(w http.ResponseWriter, r *http.Request) {
	// Get equipmentId, count
	equipmentId, ok := getURLIntParam(w, r, "equipmentId")
	if !ok {
		return
	}
	count, ok := getURLIntParam(w, r, "count")
	if !ok {
		return
	}

	stmt := `
		UPDATE equipment
		SET count = ?
		WHERE id = ?`
	_, err := db.Exec(stmt, count, equipmentId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostWebtoolsEquipment(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var equipment equipmentStruct
	err := decoder.Decode(&equipment)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stmt := `
		INSERT INTO equipment (
			create_datetime,
			description,
			count)
		VALUES (datetime('now'), ?, ?)`
	_, err = db.Exec(stmt, equipment.Description, equipment.Count)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
