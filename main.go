package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	ErrPersonNotExit = errors.New("person does not exist")
)

type Person struct {
	FirstName  string `json:"firstName"`
	SecondName string `json:"secondName"`
	Age        int    `json:"age"`
}

const addr = ":8080"
const dbFileName = "db.json"

func main() {
	ds := NewDatastore(dbFileName)
	pc := NewController(ds)
	rt := NewRouter(pc)

	if err := rt.Run(); err != nil {
		log.Fatal(err)
	}
}

type Controller struct {
	datastore *Datastore
}

func NewController(ds *Datastore) *Controller {
	return &Controller{
		datastore: ds,
	}
}

func (c *Controller) HandleGetPerson(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	person, err := c.datastore.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(person)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (c *Controller) HandleSetPerson(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := c.datastore.Set(id, person); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *Controller) HandleRemovePerson(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := c.datastore.Remove(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type Router struct{}

func NewRouter(pc *Controller) *Router {
	// swagger:operation GET /people/{id} getPerson
	//
	// Returns a person with sepcified id
	//
	// any existing person
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: person id
	//   required: true
	//   type: integer
	//   format: int32
	// responses:
	//   '200':
	//     description: person response
	//     schema:
	//       type: object
	//       properties:
	//		   firstName:
	//		     type: integer
	//		   secondName:
	//			 type: string
	//		   age:
	//			 type: integer
	//   '500':
	//     description: any error
	http.HandleFunc("GET /people/{id}", pc.HandleGetPerson)

	// swagger:operation POST /people/{id} setPerson
	//
	// Sets person with sepcified id
	//
	// any person
	//
	// ---
	// parameters:
	// - name: id
	//   in: path
	//   description: person id
	//   required: true
	//   type: integer
	//   format: int32
	// responses:
	//   '200':
	//     description: person response
	//   '500':
	//     description: any error
	http.HandleFunc("POST /people/{id}", pc.HandleSetPerson)

	// swagger:operation DELETE /people/{id} removePerson
	//
	// Removes person with sepcified id
	//
	// any existing person
	//
	// ---
	// parameters:
	// - name: id
	//   in: path
	//   description: person id
	//   required: true
	//   type: integer
	//   format: int32
	// responses:
	//   '200':
	//     description: person response
	//   '500':
	//     description: any error
	http.HandleFunc("DELETE /people/{id}", pc.HandleRemovePerson)

	return &Router{}
}

func (r *Router) Run() error {
	log.Printf("Server running %s", addr)
	return http.ListenAndServe(addr, nil)
}

type Datastore struct {
	fileName string
}

func NewDatastore(fileName string) *Datastore {
	return &Datastore{
		fileName: fileName,
	}
}

func (d *Datastore) loadPeople() (map[int]Person, error) {
	file, err := os.ReadFile(d.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[int]Person), nil
		}

		return nil, err
	}

	var people map[int]Person
	err = json.Unmarshal(file, &people)
	if err != nil {
		return nil, err
	}

	return people, nil
}

func (d *Datastore) savePeople(people map[int]Person) error {
	data, err := json.MarshalIndent(people, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(d.fileName, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (d *Datastore) Get(id int) (Person, error) {
	people, err := d.loadPeople()
	if err != nil {
		return Person{}, err
	}

	if person, ok := people[id]; ok {
		return person, nil
	}

	return Person{}, ErrPersonNotExit
}

func (d *Datastore) Set(id int, person Person) error {
	people, err := d.loadPeople()
	if err != nil {
		return err
	}

	people[id] = person

	if err := d.savePeople(people); err != nil {
		return err
	}

	return nil
}

func (d *Datastore) Remove(id int) error {
	people, err := d.loadPeople()
	if err != nil {
		return err
	}

	delete(people, id)

	if err := d.savePeople(people); err != nil {
		return err
	}

	return nil
}
