package hpk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	//	"github.com/kylelemons/godebug/pretty"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MongoConnect(): Generates a mgo session using environment variables
func MongoConnect() *mgo.Session {
	mongoHost := "localhost"
	mongoPort := "27017"
	if mp := os.Getenv("MONGO_HOST"); mp != "" {
		mongoHost = mp
	}
	if mp := os.Getenv("MONGO_PORT"); mp != "" {
		mongoPort = mp
	}

	session, err := mgo.Dial(mongoHost + ":" + mongoPort)
	if err != nil {
		log.Panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	return session
}

func KApi() {
	r := mux.NewRouter()

	kubeconfig := ParseKubeconfig()

	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		ConfigHandler(w, r, kubeconfig)
	}).Methods("GET")
	r.HandleFunc("/config/{key}", func(w http.ResponseWriter, r *http.Request) {
		ConfigKeyHandler(w, r, kubeconfig)
	}).Methods("GET", "PUT", "POST", "DELETE")
	r.HandleFunc("/allocations", AllocationsCategoryHandler).Methods("GET")
	r.HandleFunc("/allocations/{name}", AllocationsHandler).Methods("GET", "PUT", "POST")
	r.HandleFunc("/jobs", JobsCategoryHandler).Methods("GET")
	r.HandleFunc("/jobs_use/{id}", JobsUseHandler).Methods("GET")
	r.HandleFunc("/jobs/{id}", JobsHandler).Methods("GET")
	r.HandleFunc("/jobs/{id}/logs", JobsLogsHandler).Methods("GET")
	r.PathPrefix("/dashboard/").Handler(http.StripPrefix("/dashboard/", http.FileServer(http.Dir("./kdash/static/"))))

	http.Handle("/", r)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "kJob API Server")
	}
}

func AllocationsCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		session := MongoConnect()
		jobsQuery := session.DB("hpk").C("allocations")

		InitializeAllocations(jobsQuery)

		var m []bson.M
		jobsQuery.Find(nil).Sort("name").All(&m)
		w.WriteHeader(http.StatusOK)
		jsonData, err := json.Marshal(m)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
}

func AllocationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := MongoConnect()
	jobsQuery := session.DB("hpk").C("allocations")
	if r.Method == "GET" {
		InitializeAllocations(jobsQuery)
		var m []bson.M
		jobsQuery.Find(bson.M{"name": vars["name"]}).All(&m)
		w.WriteHeader(http.StatusOK)
		jsonData, err := json.Marshal(m)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
	if r.Method == "POST" || r.Method == "PUT" {
		c := session.DB("hpk").C("allocations")
		object, err := c.Upsert(bson.M{"name": vars["name"]}, bson.M{
			"name":    vars["name"],
			"balance": r.PostFormValue("balance"),
		})
		if err != nil {
			log.Info(err)
		}
		jsonData, err := json.Marshal(object)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
}

func ConfigHandler(w http.ResponseWriter, r *http.Request, kubeconfig *string) {
	if r.Method == "GET" {
		annotations := GetFullConfig(kubeconfig)

		w.WriteHeader(http.StatusOK)

		jsonData, err := json.Marshal(annotations)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
}

func ConfigKeyHandler(w http.ResponseWriter, r *http.Request, kubeconfig *string) {
	vars := mux.Vars(r)
	annotations := GetFullConfig(kubeconfig)

	//pretty.Print(r)

	if r.Method == "GET" {
		annotations = make(map[string]string)
		annotations[vars["key"]] = r.PostFormValue("val")
	}

	if r.Method == "POST" || r.Method == "PUT" {
		annotations = UpdateConfigKey(kubeconfig, vars["key"], r.PostFormValue("val"))
	}
	if r.Method == "DELETE" {
		annotations = RemoveConfigKey(kubeconfig, vars["key"])
	}
	jsonData, err := json.Marshal(annotations)
	if err != nil {
		log.Info(err)
	}
	fmt.Fprintf(w, "%s", string(jsonData[:]))
}
func JobsCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		session := MongoConnect()
		jobsQuery := session.DB("hpk").C("jobs")

		var m []bson.M
		jobsQuery.Find(nil).Sort("-name").All(&m)
		w.WriteHeader(http.StatusOK)
		jsonData, err := json.Marshal(m)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
}

func JobsUseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		vars := mux.Vars(r)
		session := MongoConnect()
		jobsQuery := session.DB("hpk").C("jobs_usage")

		var m []bson.M
		jobsQuery.Find(bson.M{"name": vars["id"]}).All(&m)
		w.WriteHeader(http.StatusOK)
		jsonData, err := json.Marshal(m)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
}

func JobsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := MongoConnect()
	jobsQuery := session.DB("hpk").C("jobs")
	if r.Method == "GET" {
		var m []bson.M

		jobsQuery.Find(bson.M{"name": vars["id"]}).All(&m)
		w.WriteHeader(http.StatusOK)
		jsonData, err := json.Marshal(m)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
}

func JobsLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := MongoConnect()
	jobsQuery := session.DB("hpk").C("job_logs")
	if r.Method == "GET" {
		var m []bson.M
		jobsQuery.Find(bson.M{"jobName": vars["id"]}).All(&m)
		w.WriteHeader(http.StatusOK)
		jsonData, err := json.Marshal(m)
		if err != nil {
			log.Info(err)
		}
		fmt.Fprintf(w, "%s", string(jsonData[:]))
	}
}
