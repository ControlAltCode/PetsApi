package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ControlAltCode/pets/api/auth"
	"github.com/ControlAltCode/pets/api/models"
	"github.com/ControlAltCode/pets/api/responses"
	"github.com/ControlAltCode/pets/api/utils/formaterror"
	"github.com/gorilla/mux"
)

func (server *Server) CreateVeterinary(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	veterinary := models.Veterinary{}
	err = json.Unmarshal(body, &veterinary)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	veterinary.Prepare()
	err = veterinary.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if uid != veterinary.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}
	veterinaryCreated, err := veterinary.SaveVeterinary(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, veterinaryCreated.ID))
	responses.JSON(w, http.StatusCreated, veterinaryCreated)
}

func (server *Server) GetVeterinaries(w http.ResponseWriter, r *http.Request) {

	veterinary := models.Veterinary{}

	veterinaries, err := veterinary.FindAllVeterinaries(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, veterinaries)
}

func (server *Server) GetVeterinary(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	veterinary := models.Veterinary{}

	veterinaryReceived, err := veterinary.FindVeterinaryByID(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, veterinaryReceived)
}

func (server *Server) UpdateVeterinary(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Check if the veterinary id is valid
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//CHeck if the auth token is valid and  get the user id from it
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the veterinary exist
	veterinary := models.Veterinary{}
	err = server.DB.Debug().Model(models.Veterinary{}).Where("id = ?", pid).Take(&veterinary).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Veterinary not found"))
		return
	}

	// If a user attempt to update a veterinary not belonging to him
	if uid != veterinary.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	// Read the data posted
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Start processing the request data
	veterinaryUpdate := models.Veterinary{}
	err = json.Unmarshal(body, &veterinaryUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	//Also check if the request user id is equal to the one gotten from token
	if uid != veterinaryUpdate.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	veterinaryUpdate.Prepare()
	err = veterinaryUpdate.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	veterinaryUpdate.ID = veterinary.ID //this is important to tell the model the veterinary id to update, the other update field are set above

	veterinaryUpdated, err := veterinaryUpdate.UpdateAVeterinary(server.DB)

	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	responses.JSON(w, http.StatusOK, veterinaryUpdated)
}

func (server *Server) DeleteVeterinary(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Is a valid veterinary id given to us?
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Is this user authenticated?
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the veterinary exist
	veterinary := models.Veterinary{}
	err = server.DB.Debug().Model(models.Veterinary{}).Where("id = ?", pid).Take(&veterinary).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}

	// Is the authenticated user, the owner of this veterinary?
	if uid != veterinary.UserID {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	_, err = veterinary.DeleteAVeterinary(server.DB, pid, uid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", pid))
	responses.JSON(w, http.StatusNoContent, "")
}
