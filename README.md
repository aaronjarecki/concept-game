# Concept-Game Service

## Running the service
1. Clone this repo
1. Populate the *assets* directory with clue images using the '001.png' naming scheme
1. `cf push` the code to Cloud Foundry or Pez. Call the app "concept-game".
1. Bind a MySQL service to the app.
1. Restage the app.

## Using the service
Most of the time you'll want to hit the [create endpoint](http://concept-game.cfapps.pez.pivotal.io/create).  
If you've already created a puzzle, you can use the [view endpoint](http://concept-game.cfapps.pez.pivotal.io/view?puzzleId=ugliest-famous-Porygon) to see it.  
You can also hit the [watchRecent endpoint](http://concept-game.cfapps.pez.pivotal.io/watchRecent) to watch someone build a puzzle (using the create page) in real time.

### The Create Page
This page features:
- Clickable **coloured icons** for choosing between concepts or sub-concepts.
- Clickable **clue icons**. Clicking these will add a new clue to the puzzle.
- A **puzzle identifier** at the top of the page, consisting of a three-word string.
- An image representing the current puzzle.
- A **save** button with associated **solution** and **author** fields. Saving a puzzle will store it to the database, ensuring it will be available even after the service is restarted.  
Once a puzzle has been saved, a *view* link is provided. A view link can also be created by pasting the *puzzle identifier* as the **puzzleId** property of the view endpoint. EG. http://concept-game.cfapps.pez.pivotal.io/view?puzzleId=some-puzzle-id

### The View Page
The view endpoint can take an optional **asPng** param. When this param is set, the page will display the puzzle as a PNG image instead of html.

# Contributing
Feel free to fork this repo and write your own endpoints and endpoint templates. Send a Pull Request to get your changes add to the project.
