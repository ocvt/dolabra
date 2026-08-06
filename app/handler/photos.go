package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/ocvt/dolabra/utils"
	"google.golang.org/api/drive/v3"
)

const PHOTO_CACHE_DIR = "data/photo-cache"

// Drive file ids are url-safe base64; anything else is rejected before the
// id is used as a cache filename
var photoIdRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Shared drive service; constructing one per request leaks the underlying
// transport and its connection pool
var driveService *drive.Service
var driveServiceOnce sync.Once
var driveServiceErr error

func getDriveService() (*drive.Service, error) {
	driveServiceOnce.Do(func() {
		driveService, driveServiceErr = drive.NewService(context.Background())
	})
	return driveService, driveServiceErr
}

// Folder listings hit the drive api; cache them briefly so page views don't
// each pay a drive round trip. Uploads invalidate their folder.
const PHOTO_LIST_CACHE_TTL = 5 * time.Minute

type photoListEntry struct {
	mainphoto []map[string]string
	imageList []map[string]string
	expires   time.Time
}

var photoListCache = map[string]photoListEntry{}
var photoListMutex sync.Mutex

/* HELPERS */
func getPhotos(w http.ResponseWriter, tripFolderId string) ([]map[string]string, []map[string]string, bool) {
	photoListMutex.Lock()
	entry, hit := photoListCache[tripFolderId]
	photoListMutex.Unlock()
	if hit && time.Now().Before(entry.expires) {
		return entry.mainphoto, entry.imageList, true
	}

	service, err := getDriveService()
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
	mainphoto := []map[string]string{}
	imageList := []map[string]string{}
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

	photoListMutex.Lock()
	photoListCache[tripFolderId] = photoListEntry{
		mainphoto: mainphoto,
		imageList: imageList,
		expires:   time.Now().Add(PHOTO_LIST_CACHE_TTL),
	}
	photoListMutex.Unlock()

	return mainphoto, imageList, true
}

func getTripFolderId(w http.ResponseWriter, tripId string) (string, bool, bool) {
	service, err := getDriveService()
	if !checkError(w, err) {
		return "", false, false
	}

	// Lookup trip folder
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and "+
		"'%s' in parents and "+
		"name = '%s'",
		utils.GetConfig().GDriveTripsFolderId, tripId)
	folderListStruct, err := service.Files.List().Q(query).Fields("files/id").Do()
	if !checkError(w, err) {
		return "", false, false
	}
	if len(folderListStruct.Files) > 0 {
		return folderListStruct.Files[0].Id, true, true
	}

	return "", false, true
}

func newTripFolderId(w http.ResponseWriter, tripId string) (string, bool) {
	service, err := getDriveService()
	if !checkError(w, err) {
		return "", false
	}

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

	service, err := getDriveService()
	if !checkError(w, err) {
		return false
	}

	tripFolderId, exists, ok := getTripFolderId(w, tripId)
	if !ok {
		return false
	}
	if !exists {
		tripFolderId, ok = newTripFolderId(w, tripId)
		if !ok {
			return false
		}
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
	if !checkError(w, err) {
		return false
	}

	// New photo; drop the folder's cached listing
	photoListMutex.Lock()
	delete(photoListCache, tripFolderId)
	photoListMutex.Unlock()

	return true
}

/* MAIN FUNCTIONS */
func GetAllTripsPhotos(w http.ResponseWriter, r *http.Request) {
	service, err := getDriveService()
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
	photoId := chi.URLParam(r, "photoId")
	if !photoIdRegexp.MatchString(photoId) {
		respondError(w, http.StatusBadRequest, "Invalid photo id")
		return
	}

	// Serve from disk cache if present
	cachePath := filepath.Join(PHOTO_CACHE_DIR, photoId)
	if cached, err := os.Open(cachePath); err == nil {
		defer cached.Close()
		_, err = io.Copy(w, cached)
		if err != nil {
			log.Print("Failed writing response: " + err.Error())
		}
		return
	}

	service, err := getDriveService()
	if !checkError(w, err) {
		return
	}

	// Download photo, teeing it into the cache while responding
	photoRes, err := service.Files.Get(photoId).Download()
	if !checkError(w, err) {
		return
	}
	defer photoRes.Body.Close()

	tmp, err := os.CreateTemp(PHOTO_CACHE_DIR, photoId+".tmp*")
	if err != nil {
		// Cache dir unavailable; still serve the photo
		log.Print("Photo cache unavailable: " + err.Error())
		_, err = io.Copy(w, photoRes.Body)
		if err != nil {
			log.Print("Failed writing response: " + err.Error())
		}
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = io.Copy(w, io.TeeReader(photoRes.Body, tmp))
	if err != nil {
		// Client hung up or download failed; drop the partial cache file
		log.Print("Failed writing response: " + err.Error())
		return
	}
	err = os.Rename(tmp.Name(), cachePath)
	if err != nil {
		log.Print("Failed caching photo: " + err.Error())
	}
}

func GetTripsPhotos(w http.ResponseWriter, r *http.Request) {
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Lookup trip folder
	tripFolderId, exists, ok := getTripFolderId(w, strconv.Itoa(tripId))
	if !ok {
		return
	}

	mainphoto := []map[string]string{}
	imageList := []map[string]string{}
	if exists {
		mainphoto, imageList, ok = getPhotos(w, tripFolderId)
		if !ok {
			return
		}
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
