# Contented (Angular and GoBuffalo Setup / Playground)

This is just a simple app for iterating over directories of images (eventually video) given a root path.  It can also create
previews for larger images.

Mostly it is an example code playground so you can browse images, see Angular keypress navigation events and hack around
with the usual http downloads, headers and ajax support in GoLang.

## Database Setup

This is done via docker right now, if you just run docker compose up -d db it will get you postgres, but that is not required.
To enable database support you then edit the .env file and set USE_DATABASE="true".  By default it will just use memory
to load up the contents in your DIR environment file (fully qualified).

The first thing you need to do is open up the "database.yml" file and edit it to use the correct 
usernames, passwords, hosts, etc... that are appropriate for your environment.  These are correct 
for the default dev instance.

### Create Your Databases (Not required)

There is a docker compose file that you can use to bring up a postgres db for development.  You can also then reset the database using buff

    $ docker-compose up -d db
    $ buffalo pop create -a
    $ buffalo pop reset

To populate the DB with the test directory you can set an environment variable in DIR (fully qualified) and then
run the db grift to populate the db.

    $ buffalo task db:seed

Creating previews for larger images can be done only with the database version at this time (not memory).

    $ buffalo task db:preview

Nuking out the current set of loaded data that is in the db..

    $ export NUKE_IT=y && buffalo task db:nuke


###  Development
Start by running yarn install in order to get all the required javascript and typescript installed.

    See a list of all the available tasks
    $ yarn run gulp --tasks 

    For doing rapid development on javascript I run these in two different consoles
    $ yarn run ng build contented --configuration=dev --watch=true --deploy-url /public/build/
    $ yr ng test contented --watch=true

TODO: Make it so the base gulp run is a little smarter, buffalo dev is SUPPOSED to do solid reloads on file
change but it is actually kinda sketchy.
TODO: Maybe stop using buffalo dev and instead rebuild and kick it manually...?

## Starting the Application

The app is currently split into an angular setup, a gulp helpfile (in Typescript) and the buffalo system.  Buffalo dev
mode seems to mostly restart buffalo correctly but sometimes doesn't notice a file save.  Save again...

    $ yarn install
    $ yarn run gulp buildDeploy  # Running just yarn run gulp will kick off the dev builds and watchers
	$ export DIR=`pwd`/mocks/content/ && buffalo dev
	$ export DIR=`pwd`/mocks/content/ && buffalo test

Alternatively you can run it as a docker setup but that is a little rougher for dev.

    $docker build -f Dockerfile -t contented . 
    $docker-compose up -d

## What Next?

Need to add in something around tests and exporting the dir path maybe just have the gulp process do a watch
and kick off the tests again.  I keep hoping buffalo will add something like ng test --watch=true.

## What this does

This is pretty much just a holder for getting GoLang, GoBuffalo and Angular trying to play together.

Browse over a set of images (eventually it should expand to video).

wasd => Navigate up and down
f => Fully load all the directory contents
x => Download. 
e => Fullscreen
q => Close Fullscreen

[Powered by Buffalo](http://gobuffalo.io)
