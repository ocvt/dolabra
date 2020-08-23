package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/ocvt/dolabra/utils"
	"google.golang.org/api/drive/v3"
)

/* HELPERS */
func getPhotos(w http.ResponseWriter, tripFolderId string) ([]map[string]string, []map[string]string, bool) {
	// Use Google Application Default Credentials
	service, err := drive.NewService(context.Background())
	if !checkError(w, err) {
		return nil, nil, false
	}

	// Get trip photos
	query := fmt.Sprintf("'%s' in parents", tripFolderId)
	fileListStruct, err := service.Files.List().Q(query).Fields("files(id, name)").Do()
	if !checkError(w, err) {
		return nil, nil, false
	}

	// If exists, mainphoto is list containing single image
	var mainphoto []map[string]string
	var imageList []map[string]string
	for i := 0; i < len(fileListStruct.Files); i++ {
		if strings.HasPrefix(fileListStruct.Files[i].Name, "mainphoto") {
			mainphoto = append(imageList, map[string]string{
				"name": fileListStruct.Files[i].Name,
				"url":  utils.GetConfig().ApiUrl + "/photo/" + fileListStruct.Files[i].Id,
			})
		} else {
			imageList = append(imageList, map[string]string{
				"name": fileListStruct.Files[i].Name,
				"url":  utils.GetConfig().ApiUrl + "/photo/" + fileListStruct.Files[i].Id,
			})
		}
	}

	return mainphoto, imageList, true
}

func getTripFolderId(w http.ResponseWriter, tripId string) (string, bool) {
	// Use Google Application Default Credentials
	service, err := drive.NewService(context.Background())
	if !checkError(w, err) {
		return "", false
	}

	// Lookup trip folder
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and "+
		"'%s' in parents and "+
		"name = '%s'",
		utils.GetConfig().GDriveTripsFolderId, tripId)
	folderListStruct, err := service.Files.List().Q(query).Fields("files/id").Do()
	if !checkError(w, err) {
		return "", false
	}
	if len(folderListStruct.Files) > 0 {
		return folderListStruct.Files[0].Id, true
	}

	// Create new trip folder if it doesn't exist
	newFolder := &drive.File{
		Name:     tripId,
		Parents:  []string{utils.GetConfig().GDriveTripsFolderId},
		MimeType: "application/vnd.google-apps.folder",
	}
	newFolder, err = service.Files.Create(newFolder).Do()
	if !checkError(w, err) {
		return "", false
	}

	return newFolder.Id, true
}

func uploadTripPhoto(w http.ResponseWriter, r *http.Request, tripId string, fileName string) bool {
	// Get photo
	file, _, err := r.FormFile("photoFile")
	if !checkError(w, err) {
		return false
	}
	defer file.Close()

	// Use Google Application Default Credentials env var
	service, err := drive.NewService(context.Background())
	if !checkError(w, err) {
		return false
	}

	tripFolderId, ok := getTripFolderId(w, tripId)
	if !ok {
		return false
	}

	if fileName == "" {
		// Get random file name
		idListStruct, err := service.Files.GenerateIds().Count(1).Do()
		if !checkError(w, err) {
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
	return checkError(w, err)
}

/* MAIN FUNCTIONS */
func GetAllTripsPhotos(w http.ResponseWriter, r *http.Request) {
	// Use Google Application Default Credentials
	service, err := drive.NewService(context.Background())
	if !checkError(w, err) {
		return
	}

	// Get trip photos
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and "+
		"'%s' in parents", utils.GetConfig().GDriveTripsFolderId)
	fileListStruct, err := service.Files.List().Q(query).Fields("files(id, name)").Do()
	if !checkError(w, err) {
		return
	}

	var imageList []map[string]string
	for i := 0; i < len(fileListStruct.Files); i++ {
		_, imageListTmp, ok := getPhotos(w, fileListStruct.Files[i].Id)
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
	_, imageList, ok := getPhotos(w, utils.GetConfig().GDriveHomePhotosFolderId)
	if !ok {
		return
	}

	if imageList == nil {
		imageList = []map[string]string{}
	}

	respondJSON(w, http.StatusOK, map[string][]map[string]string{"images": imageList})
}

func GetPhoto(w http.ResponseWriter, r *http.Request) {
	log.Printf("URI: " + r.URL.RequestURI() + "\n")
	photoId := chi.URLParam(r, "photoId")

	// Use Google Application Default Credentials
	service, err := drive.NewService(context.Background())
	if !checkError(w, err) {
		return
	}

	// Download & return photo
	photoRes, err := service.Files.Get(photoId).Download()
	if !checkError(w, err) {
		return
	}

	_, err = io.Copy(w, photoRes.Body)
	if err != nil {
		log.Printf("Failed writing response: " + err.Error())
	}
}

func GetTripsPhotos(w http.ResponseWriter, r *http.Request) {
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Lookup trip folder
	tripFolderId, ok := getTripFolderId(w, strconv.Itoa(tripId))
	if !ok {
		return
	}

	mainphoto, imageList, ok := getPhotos(w, tripFolderId)
	if !ok {
		return
	}

	if mainphoto == nil {
		mainphoto = []map[string]string{}
	}
	if imageList == nil {
		imageList = []map[string]string{}
	}

	respondJSON(w, http.StatusOK, map[string][]map[string]string{"mainphoto": mainphoto, "images": imageList})
}

func PatchTripsMainphoto(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
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
	tripId, ok := getURLIntParam(w, r, "tripId")
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
