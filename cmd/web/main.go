package main

import (
	"database/sql" // New import
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql" // New import
	"github.com/golangcollege/sessions"
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include fields for the two custom loggers
// we'll add more to it as the build progresses.
// Add a snippets field to the application struct. This will allow us to
// make the SnippetModel object available to our handlers.
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	session       *sessions.Session
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
}

// Update the signature for the routes() method so that it returns a
// http.Handler instead of *http.ServeMux.
func (app *application) routes() http.Handler {
	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Create a new middleware chain containing the middleware specific to
	// our dynamic application routes. For now, this chain will only contain
	// the session middleware but we'll add more to it later.
	dynamicMiddleware := alice.New(app.session.Enable)

	// mux := http.NewServeMux()
	// mux.HandleFunc("/", app.home)
	// mux.HandleFunc("/snippet", app.showSnippet)
	// mux.HandleFunc("/snippet/create", app.createSnippet)
	mux := pat.New()
	// mux.Get("/", http.HandlerFunc(app.home))
	// mux.Get("/snippet/create", http.HandlerFunc(app.createSnippetForm))
	// mux.Post("/snippet/create", http.HandlerFunc(app.createSnippet))
	// mux.Get("/snippet/:id", http.HandlerFunc(app.showSnippet))
	
	// Update these routes to use the new dynamic middleware chain followed
	// by the appropriate handler function. 
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/snippet/create", dynamicMiddleware.ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))

	// Create a file server which serves files out of the "./ui/static" directory.
	// Note that the path given to the http.Dir function is relative to the project
	// directory root.
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// Use the mux.Handle() function to register the file server as the handler
	// all URL paths that start with "/static/". For matching paths, we strip the
	// "/static/" prefix before the request reaches the file server.
	// mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// Pass the servemux as the 'next' parameter to the secureHeaders middleware
	// Because secureHeaders is just a function, and the function returns a
	// http.Handler we don't need to do anything else.
	// return secureHeaders(mux)

	// Wrap the existing chain with the logRequest middleware.
	// return app.logRequest(secureHeaders(mux))

	// Wrap the existing chain with the recoverPanic middleware.
	// return app.recoverPanic(app.logRequest(secureHeaders(mux)))

	// Return the 'standard' middleware chain followed by the servemux.
	return standardMiddleware.Then(mux)
}

func main() {

	// Define a new command-line flag with the name "addr", a default value of ":8000",
	// and some short help text explaining what the flag controls. The value of the
	// flag will be stored in the addr variable at runtime.
	addr := flag.String("addr", ":8000", "http network address")

	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "fei:fei@tcp(172.20.0.2)/snippetbox?parseTime=true", "MySQL data")

	// Define a new command-line flag for the session secret (a random key whic
	// will be used to encrypt and authenticate session cookies). It should be
	// bytes long.
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")

	// Importantly, we use the flag.Parse() function to parse the command-line
	// This reads in the command-line flag value and assigns it to the addr variable.
	// You need to call this *before* you use the addr variable
	// ohterwise ist will always contain the default value of ":8000", If any errors encountered
	// during parsing the application will be terminated.
	flag.Parse()

	// Use log.new() to create a logger for writing information messages. This
	// three parameters: the destination to write the logs to (os.Stdout), a String
	// prefix for message (INFO followed by a tab), and flags to indicate what
	// additional information to include (local date and time). Note that the flags
	// are joined using the bitwise OR operator (|).
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a logger for writing error messages in the same way, but use Stderr
	// as the destination and use the log.Lshortfile flag to include the relevant
	// file and line number.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// To keep the main() function tidy I've put the code for creating a connection
	// pool into the separate openDB() function below. We pass openDB() the DSN
	// from the command-line flag.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Use the sessions.New() function to initialize a new session manager,
	// passing in the secret key as the parameter. Then we configure it so
	// sessions always expires after 12 hours.
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Initialize a mysql.SnippetModel instance and add it to the application
	// dependencies.
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
	}

	// Initialize a new http.Server struct. We set the Addr and Handler fields
	// that the server uses the same network address and routes as before. and
	// the ErrorLog field so that the server now uses the custom errorLog logger
	// the event of any problems
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(), // Call the new app.routes() method
	}

	// The value returned from the flag.String() function is a pointer to the flag
	// value, not the value itself. So we need to dereference the pointer(i.e.
	// prefix it with the * symbol) before using it.
	infoLog.Printf("Starting server on %s", *addr)
	// err = srv.ListenAndServe()
	// Use the ListenAndServeTLS() method to start the HTTPS server. We
	// pass in the paths to the TLS certificate and corresponding private key a
	// the two parameters.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}