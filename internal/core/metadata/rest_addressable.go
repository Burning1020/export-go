/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package metadata

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/export-go/internal/pkg/db"
	"github.com/edgexfoundry/export-go/pkg/models"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func restGetAllAddressables(w http.ResponseWriter, _ *http.Request) {
	results := make([]models.Addressable, 0)
	err := dbClient.GetAddressables(&results)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(results) > Configuration.Service.ReadMaxLimit {
		err = errors.New("Max limit exceeded")
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&results)
}

// Add a new addressable
// The name must be unique
func restAddAddressable(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var a models.Addressable
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		if len(a.Name) == 0 {
			err = errors.New("name is required for addressable")
			LoggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	var id bson.ObjectId
	id, err = dbClient.AddAddressable(&a)
	if err != nil {
		if err == db.ErrNotUnique {
			http.Error(w, "Duplicate name for addressable", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error(), "")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id.Hex()))
}

// Update addressable by ID or name (ID used first)
// 404 Not found if no addressable is found to update
func restUpdateAddressable(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var ra models.Addressable
	if err := json.NewDecoder(r.Body).Decode(&ra); err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the addressable exists
	var res models.Addressable
	err := dbClient.GetAddressableById(&res, ra.Id.Hex())
	if err != nil {
		if ra.Id == "" {
			err = dbClient.GetAddressableByName(&res, ra.Name)
		}
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error(), "")
			return
		}
	}

	// If the name is changed, check if the addressable is still in use
	if ra.Name != "" && ra.Name != res.Name {
		isStillInUse, err := isAddressableStillInUse(res)
		if err != nil {
			LoggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if isStillInUse {
			err = errors.New("Data integrity issue: Addressable is still in use")
			LoggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	if err := dbClient.UpdateAddressable(&ra, &res); err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Notify Associates
	notifyAddressableAssociates(res, http.MethodPut)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetAddressableById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars["id"]
	var result models.Addressable
	if err := dbClient.GetAddressableById(&result, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			LoggingClient.Error(err.Error(), "")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
func restDeleteAddressableById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the addressable exists
	var a models.Addressable
	err := dbClient.GetAddressableById(&a, id)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the addressable is still in use
	isStillInUse, err := isAddressableStillInUse(a)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if isStillInUse {
		err = errors.New("Data integrity issue: attempt to delete addressable still in use by a device or device service")
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	err = dbClient.DeleteAddressableById(a.Id.Hex())
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("true"))
}
func restDeleteAddressableByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	// Problems unescaping
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the addressable exists
	var a models.Addressable
	err = dbClient.GetAddressableByName(&a, n)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the addressable is still in use
	isStillInUse, err := isAddressableStillInUse(a)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if isStillInUse {
		err = errors.New("Data integrity issue: attempt to delete addressable still in use by a device or device service")
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err := dbClient.DeleteAddressableById(a.Id.Hex()); err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Helper function to determine if an addressable is still referenced by a device or device service
func isAddressableStillInUse(a models.Addressable) (bool, error) {
	// Check devices
	var d []models.Device
	err := dbClient.GetDevicesByAddressableId(&d, a.Id.Hex())
	if err != nil {
		return false, err
	}
	if len(d) > 0 {
		return true, nil
	}

	// Check device services
	var ds []models.DeviceService
	err = dbClient.GetDeviceServicesByAddressableId(&ds, a.Id.Hex())
	if err != nil {
		return false, err
	}
	if len(ds) > 0 {
		return true, nil
	}

	return false, nil
}
func restGetAddressableByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dn, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	var result models.Addressable
	if err := dbClient.GetAddressableByName(&result, dn); err != nil {
		LoggingClient.Error(err.Error(), "")
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func restGetAddressableByTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	t, err := url.QueryUnescape(vars[TOPIC])
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := make([]models.Addressable, 0)

	err = dbClient.GetAddressablesByTopic(&res, t)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetAddressableByPort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var strp string = vars[PORT]
	p, err := strconv.Atoi(strp)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := make([]models.Addressable, 0)
	if err := dbClient.GetAddressablesByPort(&res, p); err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetAddressableByPublisher(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, err := url.QueryUnescape(vars[PUBLISHER])
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := make([]models.Addressable, 0)
	err = dbClient.GetAddressablesByPublisher(&res, p)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func restGetAddressableByAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a, err := url.QueryUnescape(vars[ADDRESS])
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := make([]models.Addressable, 0)
	err = dbClient.GetAddressablesByAddress(&res, a)
	if err != nil {
		LoggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Notify the associated device services for changes in the device addressable
func notifyAddressableAssociates(a models.Addressable, action string) error {
	// Get the devices
	var d []models.Device
	if err := dbClient.GetDevicesByAddressableId(&d, a.Id.Hex()); err != nil {
		LoggingClient.Error(err.Error(), "")
		return err
	}

	// Get the services for each device
	// Use map as a Set
	dsMap := map[string]models.DeviceService{}
	var ds []models.DeviceService
	for _, device := range d {
		// Only add if not there
		if _, ok := dsMap[device.Service.Service.Id.Hex()]; !ok {
			dsMap[device.Service.Service.Id.Hex()] = device.Service
			ds = append(ds, device.Service)
		}
	}

	if err := notifyAssociates(ds, a.Id.Hex(), action, models.ADDRESSABLE); err != nil {
		LoggingClient.Error(err.Error(), "")
		return err
	}

	return nil
}
