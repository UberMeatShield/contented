# Content View and Encoding

This is now a simple app for viewing and re-encoding images and video. Initially it was a learning project for GoLang using the Buffalo framework (buffalo branch). After GoBuffalo went into an archived mode it was ported to Gin and GORM.

Currently the main focus is on encoding video for the web with a focus on high quality. It uses ffmpeg for transcoding and libvmaf for quality assessment. It can also attempt to detect if there are duplicate videos and skip encoding them. By duplicate I mean shots within video that are identical not that they are the same file and encoding and that is done by slicing the video into 1 second segments and then comparing the hash of each segment using ssim.

The UI is an angular app that is served by Gin routes. This allows for a SPA with browser based navigation that is served by Go and lets you quickly view what is in a directory. There is a UI version of the re-encoding, preview generation and duplicate detection but the makefile has command line versions that work either with a DB or purely in memory with no dependencies.

## Database Setup

This is done via docker right now, if you just run docker compose up -d db it will get you postgres, but that is not required.

To enable database support you then edit the .env file and set USE_DATABASE="true".  By default it will just use memory to load up the contents in your DIR environment variable (fully qualified).  You can also set various settings in the .env file included in the project.


### Create Your Databases (Not required)

There is a docker compose file that you can use to bring up a postgres db for development.  You can also then reset the database using makefile comands.

    $ docker-compose up -d db
    $ make install
    $ make setup  # Load up the contents of the DIR env variable and run the server

To populate the DB with the test directory you can set an environment variable in DIR (fully qualified) and then run the db grift to populate the db.

    $ export DIR="/full/path/" && make db-populate

Creating previews for larger images and an initial preview image for video can be done as follows:

    // Note that if using a db the db-populate needs to be run first
    $ export DIR="/full/path/" && make preview

Nuking out the current set of loaded data that is in the db.

    $ make db-reset

To find all video files in the directory and subdirectory and re-encode them to h265 using ffmpeg and libvmaf for quality assessment run the following:

    $ export DIR="/full/path/" && make encode

###  Development in the UI
Start by running yarn install in order to get all the required javascript and typescript installed.

    See a list of all the available tasks
    $ yarn run gulp --tasks 

    For doing rapid development on javascript I run these in two different consoles
    $ yarn run ng build contented --configuration=dev --watch=true --deploy-url /public/build/
    $ yarn run ng test contented --watch=true

## Starting the Application

The app is currently split into the angular app on top and a Gin webserver.

    $ make install  # install dependencies
    $ make typescript # Compile the typescript
	$ make dev # Run the webserver
	$ make ngtest

TODO: Buffalo is busted and the docker image for Gin still needs work.
Alternatively you can run it as a docker setup but that is a little rougher for dev without restarts etc.

    $docker build -f Dockerfile -t contented . 
    $docker-compose up -d

## What Next?

Further works is being done to make development with the Gin webserver easier like restarts after a file save etc. Allow for viewing just a single directory instead of looking at all directories under the root. Specifying the DIR as a root allows for some easier safety checks when managing content. 

## What this does

Browse over a set of images or videos in subdirectories. Create previews or webm version snapshots from the screens. It makes looking for content a bit easier when you have a lot of video and images mixed together and you want to browse it.

    DIR=/root/to/view and under view you have directories a, b, c, d with mixed images and video.

    wasd => Navigate up and down
    f => Fully load all the directory contents
    x => Download. 
    e => Fullscreen
    q => Close Fullscreen


[Example](mocks/content/ExampleLoaded.png)
