package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"google.golang.org/api/drive/v3"
)

var HOME_PHOTOS_FOLDER_ID string
var TRIPS_FOLDER_ID string

/* HELPERS */
func getTripFolderId(w http.ResponseWriter, tripId string) (string, bool) {
	// Use Google Application Default Credentials
	service, err := drive.NewService(context.Background())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return "", false
	}

	// Lookup trip folder
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and "+
		"'%s' in parents and "+
		"name = '%s'",
		TRIPS_FOLDER_ID, tripId)
	folderListStruct, err := service.Files.List().Q(query).Fields("files/id").Do()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return "", false
	}
	if len(folderListStruct.Files) > 0 {
		return folderListStruct.Files[0].Id, true
	}

	// Create new trip folder if it doesn't exist
	newFolder := &drive.File{
		Name:     tripId,
		Parents:  []string{TRIPS_FOLDER_ID},
		MimeType: "application/vnd.google-apps.folder",
	}
	newFolder, err = service.Files.Create(newFolder).Do()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return "", false
	}

	return newFolder.Id, true
}

func getPhotos(w http.ResponseWriter, tripFolderId string) ([]map[string]string, bool) {
	// Use Google Application Default Credentials
	service, err := drive.NewService(context.Background())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return nil, false
	}

	// Get trip photos
	query := fmt.Sprintf("'%s' in parents", tripFolderId)
	fileListStruct, err := service.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return nil, false
	}

	var imageList []map[string]string
	for i := 0; i < len(fileListStruct.Files); i++ {
		imageList = append(imageList, map[string]string{
			"name": fileListStruct.Files[i].Name,
			"url":  fmt.Sprintf("https://drive.google.com/uc?id=%s&export=view", fileListStruct.Files[i].Id),
		})
	}

	return imageList, true
}

func uploadTripPhoto(w http.ResponseWriter, r *http.Request, tripId string, fileName string) bool {
	// Get photo
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return false
	}
	defer file.Close()

	// Use Google Application Default Credentials env var
	service, err := drive.NewService(context.Background())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return false
	}

	tripFolderId, ok := getTripFolderId(w, tripId)
	if !ok {
		return false
	}

	if fileName == "" {
		// Get random file name
		idListStruct, err := service.Files.GenerateIds().Count(1).Do()
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return false
		}
		fileName = idListStruct.Ids[0]
	}

	// Upload photo to trip folder
	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{tripFolderId},
	}

	_, err = service.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return false
	}

	return true
}

/* MAIN FUNCTIONS */
func GetAllTripsPhotos(w http.ResponseWriter, r *http.Request) {
	// Use Google Application Default Credentials
	service, err := drive.NewService(context.Background())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get trip photos
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and "+
		"'%s' in parents", TRIPS_FOLDER_ID)
	fileListStruct, err := service.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var imageList []map[string]string
	for i := 0; i < len(fileListStruct.Files); i++ {
		imageListTmp, ok := getPhotos(w, fileListStruct.Files[i].Id)
		if !ok {
			return
		}
		imageList = append(imageList, imageListTmp...)
	}

	if imageList == nil {
		imageList = []map[string]string{}
	}

	respondJSON(w, http.StatusOK, map[string][]map[string]string{"images": imageList})
}

func GetHomePhotos(w http.ResponseWriter, r *http.Request) {
	imageList, ok := getPhotos(w, HOME_PHOTOS_FOLDER_ID)
	if !ok {
		return
	}

	if imageList == nil {
		imageList = []map[string]string{}
	}

	respondJSON(w, http.StatusOK, map[string][]map[string]string{"images": imageList})
}

func GetTripsPhotos(w http.ResponseWriter, r *http.Request) {
	tripId, ok := checkURLParam(w, r, "tripId")
	if !ok {
		return
	}

	// Lookup trip folder
	tripFolderId, ok := getTripFolderId(w, strconv.Itoa(tripId))
	if !ok {
		return
	}

	imageList, ok := getPhotos(w, tripFolderId)
	if !ok {
		return
	}

	if imageList == nil {
		imageList = []map[string]string{}
	}

	respondJSON(w, http.StatusOK, map[string][]map[string]string{"images": imageList})
}

func PostTripsMainphoto(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := checkURLParam(w, r, "tripId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsureTripLeader(w, tripId, memberId) {
		return
	}

	if !uploadTripPhoto(w, r, strconv.Itoa(tripId), "mainphoto") {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostTripsPhotos(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := checkURLParam(w, r, "tripId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsureMemberIsOnTrip(w, tripId, memberId) {
		return
	}

	if !uploadTripPhoto(w, r, strconv.Itoa(tripId), "") {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
