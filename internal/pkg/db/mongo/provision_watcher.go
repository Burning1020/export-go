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
package mongo

import (
	"github.com/edgexfoundry/export-go/internal/pkg/db"
	"github.com/edgexfoundry/export-go/pkg/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Internal version of the provision watcher struct
// Use this to handle DBRef
type mongoProvisionWatcher struct {
	models.ProvisionWatcher
}

// Custom marshaling into mongo
func (mpw mongoProvisionWatcher) GetBSON() (interface{}, error) {
	return struct {
		models.BaseObject `bson:",inline"`
		Id                bson.ObjectId         `bson:"_id,omitempty"`
		Name              string                `bson:"name"`           // unique name and identifier of the addressable
		Identifiers       map[string]string     `bson:"identifiers"`    // set of key value pairs that identify type of of address (MAC, HTTP,...) and address to watch for (00-05-1B-A1-99-99, 10.0.0.1,...)
		Profile           mgo.DBRef             `bson:"profile"`        // device profile that should be applied to the devices available at the identifier addresses
		Service           mgo.DBRef             `bson:"service"`        // device service that owns the watcher
		OperatingState    models.OperatingState `bson:"operatingState"` // operational state - either enabled or disabled
	}{
		BaseObject:     mpw.BaseObject,
		Id:             mpw.Id,
		Name:           mpw.Name,
		Identifiers:    mpw.Identifiers,
		Profile:        mgo.DBRef{Collection: db.DeviceProfile, Id: mpw.Profile.Id},
		Service:        mgo.DBRef{Collection: db.DeviceService, Id: mpw.Service.Service.Id},
		OperatingState: mpw.OperatingState,
	}, nil
}

// Custom unmarshaling out of mongo
func (mpw *mongoProvisionWatcher) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		models.BaseObject `bson:",inline"`
		Id                bson.ObjectId         `bson:"_id,omitempty"`
		Name              string                `bson:"name"`           // unique name and identifier of the addressable
		Identifiers       map[string]string     `bson:"identifiers"`    // set of key value pairs that identify type of of address (MAC, HTTP,...) and address to watch for (00-05-1B-A1-99-99, 10.0.0.1,...)
		Profile           mgo.DBRef             `bson:"profile"`        // device profile that should be applied to the devices available at the identifier addresses
		Service           mgo.DBRef             `bson:"service"`        // device service that owns the watcher
		OperatingState    models.OperatingState `bson:"operatingState"` // operational state - either enabled or disabled
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	mpw.BaseObject = decoded.BaseObject
	mpw.Id = decoded.Id
	mpw.Name = decoded.Name
	mpw.Identifiers = decoded.Identifiers
	mpw.OperatingState = decoded.OperatingState

	// De-reference the DBRef fields
	m, err := getCurrentMongoClient()
	if err != nil {
		return err
	}
	s := m.session.Copy()
	defer s.Close()

	profCol := s.DB(m.database.Name).C(db.DeviceProfile)
	servCol := s.DB(m.database.Name).C(db.DeviceService)

	var mdp mongoDeviceProfile
	var mds mongoDeviceService

	if err := profCol.FindId(decoded.Profile.Id).One(&mdp); err != nil {
		return err
	}

	if err := servCol.FindId(decoded.Service.Id).One(&mds); err != nil {
		return err
	}

	mpw.Profile = mdp.DeviceProfile
	mpw.Service = mds.DeviceService

	return nil
}
