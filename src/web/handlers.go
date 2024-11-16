package main

import (
	"fmt"
	// "html/template"
	"net/http"
	"snippetbox/src/pkg/models"
	"strconv"
)

// Change the signature of the home handler so it is defined as a method against
// *application.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Because Pat matches the "/" path exactly, we can now remove the manual check
	// of r.URL.Path != "/" from this handler.
	// if r.URL.Path != "/" {
	//  app.notFound(w) // Use the notFound() helper.
	//  return
	// }

	// panic("oops! something went wrong")

	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the render() helper.
	app.render(w, r, "home.page.html", &templateData{Snippets: s})

	// Create an instance of a templateData struct holding the slice of snippets.
	//
	// data := &templateData{Snippets: s}

	// Initialize a slice containing the paths to the two files. Note that the home.page.template
	// file must be the "first" file in the slice.
	// Include the footer partial in the template files.
	//
	// files := []string{
	// 	"./ui/html/home.page.html",
	// 	"./ui/html/base.layout.html",
	// 	"./ui/html/footer.partial.html",
	// }

	// Use the template.ParseFiles()) function to read the template file into a template set.
	// If there is an error, we log the detailed error message and the http.Error() function
	// to send a generic 500 Internal Server Error response to the user.
	// Notice that we can pass the slice of file paths as a variadic parameter.
	//
	// ts, err := template.ParseFiles(files...)
	// if err != nil {
	// 	app.serverError(w, err) // Use the serveError() helper.
	// 	return
	// }

	// We then use the Execute() method on the template set to write the template content
	// as the response body. The last parameter to Execute() represents dynamic data that
	// we want to pass in. Pass in the templateData struct when executing the template.
	//
	// err = ts.Execute(w, data)
	// if err != nil {
	// 	app.serverError(w, err) // Use the serveError() helper.
	// }
}

// Change the signature of the showSnippet handler so it is defined as a method against
// *application.
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// Pat doesn't strip the colon from the named capture key, so we need to 
	// get the value of ":id" from the query string instead of "id".
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w) // Use the notFound() helper.
		return
	}

	// Use the SnippetModel object's Get method to retrieve the data for a
	// specific record based on its ID. If no matching record is found,
	// return a 404 Not Found response.
	s, err := app.snippets.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Use the render() helper.
	app.render(w, r, "show.page.html", &templateData{Snippet: s})

	// Create an instance of a templateData struct holding the snippet data.
	//
	// data := &templateData{Snippet: s}

	// Initiallize a slice containing the paths to the show.page.html file,
	// plus the base layout and footer partials that we made earlier.
	//
	// files := []string{
	// 	"./ui/html/show.page.html",
	// 	"./ui/html/base.layout.html",
	// 	"./ui/html/footer.partial.html",
	// }

	// Parse the template files...
	//
	// ts, err := template.ParseFiles(files...)
	// if err != nil {
	// 	app.serverError(w, err)
	// 	return
	// }

	// And then execute them, Notice how we are passing in the snippets
	// data (a models.Snippet struct) as the final parameter.
	//
	// err = ts.Execute(w, data)
	// if err != nil {
	// 	app.serverError(w, err)
	// }
}

// Add a new createSnippetForm handler, which for now returns a placeholder response.
func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new snippet..."))
}

// Change the signature of the createSnippet handler so it is defined as a method against
// *application.
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	// The check of r.Method != "POST" is now superfluous and can be removed.
	// if r.Method != "POST" {
	// 	w.Header().Set("Allow", "POST")
	// 	app.clientError(w, http.StatusMethodNotAllowed) // Use the clientError() helper.
	// 	return
	// }

	// Create some variables holding dummy data. We'll remove these later on
	// during the build.
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi"
	expires := "7"

	// Pass the data to the SnippetModel.Insert() method, receiving the
	// ID of the new record back.
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Redirect the user to the relevant page for the snippet.
	// http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)
	// Change the redirect to use the new semantic URL style of /snippet/:id
	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}
