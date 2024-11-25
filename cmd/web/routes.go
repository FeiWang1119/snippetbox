package main

import (
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
	"net/http"
)

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

	// Add the five new routes.
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.ThenFunc(app.logoutUser))

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
