package main

import (
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/gorilla/mux"
	sqlx "github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rengas/tn-covid19-bed-alert/client"
	"github.com/rengas/tn-covid19-bed-alert/handler"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"html/template"
	"log"
	"net/http"
)

func main() {
	log.Println("Init tn-covid19-bed-alert")
	initConfig()
	db := initDatabase()
	fcm := initFcm()

	homeTmpl := template.Must(template.ParseFiles("./static/index.html")) // Parse template file.
	statusTmpl := template.Must(template.ParseFiles("./static/sub_status.html"))
	env := viper.GetString("GO_ENV")

	router := mux.NewRouter()
	router.HandleFunc("/", handler.HomeHandle{homeTmpl, db}.HomeHandler)
	router.HandleFunc("/health", handler.HealthHandler)
	router.HandleFunc("/sync", handler.SyncHandle{db,}.SyncHandler)
	router.HandleFunc("/subscribe", handler.SubHandle{db,fcm,env}.SubscribeHandler)
	router.HandleFunc("/unsubscribe", handler.SubHandle{db,fcm,env}.UnSubscribeHandler)
	router.HandleFunc("/message", handler.MessageHandle{db}.ViewMessageHandler)
	router.HandleFunc("/notify", handler.NotifyHandle{statusTmpl, db, fcm,env}.NotifyHandler)

	// Not a proper way as its adds the header to all the static resources
	changeHeaderThenServe := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set some header.
			w.Header().Add("Keep-Alive", "300")
			w.Header().Add("Service-Worker-Allowed", "/")

			// Serve with the actual handler.
			h.ServeHTTP(w, r)
		}
	}

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		changeHeaderThenServe(http.FileServer(http.Dir("./static/")))))

	http.Handle("/", router)
	port := viper.GetInt64("PORT")

	addr := fmt.Sprintf(":%d", port)
	log.Printf("printing port...%s",addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func initConfig() {
	viper.AutomaticEnv()
	environment := viper.GetString("GO_ENV")
	if environment=="dev"{
		viper.AddConfigPath(".")
		filename := "config-" + environment + ".json"
		log.Printf("setting up config from %s", filename)

		viper.SetConfigFile(filename)
		viper.SetConfigType("json")

		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}

}

func initFcm() *client.FCMClient {
	creds := viper.GetString("GOOGLE_CREDS")
	opt := option.WithCredentialsJSON([]byte(creds))
	config := &firebase.Config{ProjectID: "tn-covid-bed-alert"}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	c, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("unable to intialise fcm %v", err)
	}
	return &client.FCMClient{
		c,
	}
}

func initDatabase() *sqlx.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=%s",
			viper.GetString("HOST"),
			viper.GetInt64("DB_PORT"),
			viper.GetString("USER_NAME"),
			viper.GetString("PASSWORD"),
			viper.GetString("DATABASE_NAME"),
			viper.GetString("SSL_MODE"),
			)

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("unable to open connection to database %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("unable to ping to database %v", err)
	}

	fmt.Println("Successfully connected!")
	return db
}
