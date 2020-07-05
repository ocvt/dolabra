package handler

import (
	"encoding/json"
	"net/http"
)

type newsStruct struct {
	/* Managed server side */
	Id             int    `json:"id"`
	CreateDatetime string `json:"createDatetime,omitempty"`
	// from member table
	FirstName string `json:"firstName,omitempty"`
	/* Required fields for creating news */
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Content string `json:"content"`
}

func DeleteWebtoolsNews(w http.ResponseWriter, r *http.Request) {
	// Get newsId
	newsId, ok := getURLIntParam(w, r, "newsId")
	if !ok {
		return
	}

	stmt := `
		UPDATE news
		SET publish = false
		WHERE id = ?`
	_, err := db.Exec(stmt, newsId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func GetNews(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT
			member.first_name,
			news.id,
			news.create_datetime,
			news.title,
			news.summary,
			news.content
		FROM member
		INNER JOIN news ON news.member_id = member.id
		WHERE news.publish = true
		ORDER BY datetime(news.create_datetime) DESC
		LIMIT 10`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var news = []*newsStruct{}
	i := 0
	for rows.Next() {
		news = append(news, &newsStruct{})
		err = rows.Scan(
			&news[i].FirstName,
			&news[i].Id,
			&news[i].CreateDatetime,
			&news[i].Title,
			&news[i].Summary,
			&news[i].Content)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*newsStruct{"news": news})
}

func GetNewsArchive(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT
			member.first_name,
			news.id,
			news.create_datetime,
			news.title,
			news.summary,
			news.content
		FROM member
		INNER JOIN news ON news.member_id = member.id
		WHERE news.publish = true`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var news = []*newsStruct{}
	i := 0
	for rows.Next() {
		news = append(news, &newsStruct{})
		err = rows.Scan(
			&news[i].FirstName,
			&news[i].Id,
			&news[i].CreateDatetime,
			&news[i].Title,
			&news[i].Summary,
			&news[i].Content)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*newsStruct{"news": news})
}

func PostWebtoolsNews(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var news newsStruct
	err := decoder.Decode(&news)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stmt := `
		INSERT INTO news (
			member_id,
			create_datetime,
			title,
			summary,
			content)
		VALUES (?, datetime('now'), ?, ?, ?)`
	_, err = db.Exec(
		stmt,
		memberId,
		news.Title,
		news.Summary,
		news.Content)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
